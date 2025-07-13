package controllers

import (
	"context"
	"fmt"
	"time"

	"github.com/go-logr/logr"
	"github.com/robfig/cron/v3"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"

	opsv1alpha1 "github.com/automationpi/kubejanitor/api/v1alpha1"
	"github.com/automationpi/kubejanitor/pkg/cleanup"
	"github.com/automationpi/kubejanitor/pkg/metrics"
)

const (
	// FinalizerName is the finalizer name for JanitorPolicy
	FinalizerName = "janitorpolicy.ops.k8s.io/finalizer"

	// ConditionTypeReady represents the ready condition
	ConditionTypeReady = "Ready"

	// ConditionTypeScheduled represents the scheduled condition
	ConditionTypeScheduled = "Scheduled"

	// ReasonSucceeded represents successful operation
	ReasonSucceeded = "Succeeded"

	// ReasonFailed represents failed operation
	ReasonFailed = "Failed"

	// ReasonScheduled represents scheduled operation
	ReasonScheduled = "Scheduled"

	// EventTypeNormal represents normal event
	EventTypeNormal = "Normal"

	// EventTypeWarning represents warning event
	EventTypeWarning = "Warning"
)

// JanitorPolicyReconciler reconciles a JanitorPolicy object
type JanitorPolicyReconciler struct {
	client.Client
	Scheme   *runtime.Scheme
	Recorder record.EventRecorder
	Log      logr.Logger

	// Internal state
	cronScheduler *cron.Cron
	cleanupEngine *cleanup.Engine
	metricsServer *metrics.Server
}

//+kubebuilder:rbac:groups=ops.k8s.io,resources=janitorpolicies,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=ops.k8s.io,resources=janitorpolicies/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=ops.k8s.io,resources=janitorpolicies/finalizers,verbs=update
//+kubebuilder:rbac:groups="",resources=pods;persistentvolumeclaims;configmaps;secrets;services;events,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=apps,resources=deployments;replicasets;daemonsets;statefulsets,verbs=get;list;watch
//+kubebuilder:rbac:groups=batch,resources=jobs;cronjobs,verbs=get;list;watch;delete
//+kubebuilder:rbac:groups=rbac.authorization.k8s.io,resources=roles;rolebindings;clusterroles;clusterrolebindings,verbs=get;list;watch;delete

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
func (r *JanitorPolicyReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := r.Log.WithValues("janitorpolicy", req.NamespacedName)

	// Fetch the JanitorPolicy instance
	var janitorPolicy opsv1alpha1.JanitorPolicy
	if err := r.Get(ctx, req.NamespacedName, &janitorPolicy); err != nil {
		if errors.IsNotFound(err) {
			log.Info("JanitorPolicy resource not found. Ignoring since object must be deleted")
			return ctrl.Result{}, nil
		}
		log.Error(err, "Failed to get JanitorPolicy")
		return ctrl.Result{}, err
	}

	// Handle deletion
	if janitorPolicy.DeletionTimestamp != nil {
		return r.handleDeletion(ctx, &janitorPolicy)
	}

	// Add finalizer if not present
	if !controllerutil.ContainsFinalizer(&janitorPolicy, FinalizerName) {
		controllerutil.AddFinalizer(&janitorPolicy, FinalizerName)
		if err := r.Update(ctx, &janitorPolicy); err != nil {
			log.Error(err, "Failed to add finalizer")
			return ctrl.Result{}, err
		}
		return ctrl.Result{Requeue: true}, nil
	}

	// Initialize status if needed
	if janitorPolicy.Status.Phase == "" {
		janitorPolicy.Status.Phase = "Active"
		if err := r.Status().Update(ctx, &janitorPolicy); err != nil {
			log.Error(err, "Failed to update status")
			return ctrl.Result{}, err
		}
	}

	// Schedule cleanup if schedule is configured
	if janitorPolicy.Spec.Schedule != "" {
		if err := r.scheduleCleanup(ctx, &janitorPolicy); err != nil {
			log.Error(err, "Failed to schedule cleanup")
			r.updateCondition(&janitorPolicy, ConditionTypeScheduled, metav1.ConditionFalse, ReasonFailed, err.Error())
			r.Recorder.Event(&janitorPolicy, EventTypeWarning, ReasonFailed, fmt.Sprintf("Failed to schedule cleanup: %v", err))
			return ctrl.Result{}, err
		}
		r.updateCondition(&janitorPolicy, ConditionTypeScheduled, metav1.ConditionTrue, ReasonScheduled, "Cleanup scheduled successfully")
	}

	// Update ready condition
	r.updateCondition(&janitorPolicy, ConditionTypeReady, metav1.ConditionTrue, ReasonSucceeded, "JanitorPolicy is ready")

	// Update status
	if err := r.Status().Update(ctx, &janitorPolicy); err != nil {
		log.Error(err, "Failed to update status")
		return ctrl.Result{}, err
	}

	// Requeue to check for schedule updates
	return ctrl.Result{RequeueAfter: time.Minute * 5}, nil
}

// handleDeletion handles the deletion of JanitorPolicy
func (r *JanitorPolicyReconciler) handleDeletion(ctx context.Context, janitorPolicy *opsv1alpha1.JanitorPolicy) (ctrl.Result, error) {
	log := r.Log.WithValues("janitorpolicy", janitorPolicy.Name)

	// Remove from scheduler if scheduled
	if janitorPolicy.Spec.Schedule != "" {
		r.removeFromScheduler(janitorPolicy)
	}

	// Remove finalizer
	controllerutil.RemoveFinalizer(janitorPolicy, FinalizerName)
	if err := r.Update(ctx, janitorPolicy); err != nil {
		log.Error(err, "Failed to remove finalizer")
		return ctrl.Result{}, err
	}

	log.Info("JanitorPolicy deleted successfully")
	return ctrl.Result{}, nil
}

// scheduleCleanup schedules the cleanup job based on the cron schedule
func (r *JanitorPolicyReconciler) scheduleCleanup(ctx context.Context, janitorPolicy *opsv1alpha1.JanitorPolicy) error {
	log := r.Log.WithValues("janitorpolicy", janitorPolicy.Name)

	// Remove existing schedule if any
	r.removeFromScheduler(janitorPolicy)

	// Parse and validate cron schedule
	schedule, err := cron.ParseStandard(janitorPolicy.Spec.Schedule)
	if err != nil {
		return fmt.Errorf("invalid cron schedule: %w", err)
	}

	// Add to scheduler
	entryID, err := r.cronScheduler.AddFunc(janitorPolicy.Spec.Schedule, func() {
		r.executeCleanup(ctx, janitorPolicy)
	})
	if err != nil {
		return fmt.Errorf("failed to add to scheduler: %w", err)
	}

	// Update next run time
	nextRun := schedule.Next(time.Now())
	janitorPolicy.Status.NextRun = &metav1.Time{Time: nextRun}

	log.Info("Cleanup scheduled", "schedule", janitorPolicy.Spec.Schedule, "nextRun", nextRun, "entryID", entryID)
	return nil
}

// removeFromScheduler removes the cleanup job from the scheduler
func (r *JanitorPolicyReconciler) removeFromScheduler(janitorPolicy *opsv1alpha1.JanitorPolicy) {
	// This is a simplified implementation
	// In a real implementation, you would track entry IDs and remove them
	log := r.Log.WithValues("janitorpolicy", janitorPolicy.Name)
	log.Info("Removing from scheduler")
}

// executeCleanup executes the cleanup operation
func (r *JanitorPolicyReconciler) executeCleanup(ctx context.Context, janitorPolicy *opsv1alpha1.JanitorPolicy) {
	log := r.Log.WithValues("janitorpolicy", janitorPolicy.Name)
	startTime := time.Now()

	log.Info("Starting cleanup execution")

	// Create cleanup context
	cleanupCtx := &cleanup.Context{
		Client:       r.Client,
		Policy:       janitorPolicy,
		DryRun:       janitorPolicy.Spec.DryRun,
		Logger:       log,
		EventRecorder: r.Recorder,
	}

	// Execute cleanup
	stats, err := r.cleanupEngine.Execute(ctx, cleanupCtx)
	duration := time.Since(startTime)

	// Update policy status
	var updatedPolicy opsv1alpha1.JanitorPolicy
	if err := r.Get(ctx, types.NamespacedName{Name: janitorPolicy.Name, Namespace: janitorPolicy.Namespace}, &updatedPolicy); err != nil {
		log.Error(err, "Failed to get JanitorPolicy for status update")
		return
	}

	// Update status
	now := metav1.Now()
	updatedPolicy.Status.LastRun = &now
	updatedPolicy.Status.Stats = stats
	updatedPolicy.Status.Stats.Duration = duration.String()

	if err != nil {
		updatedPolicy.Status.Message = fmt.Sprintf("Cleanup failed: %v", err)
		r.updateCondition(&updatedPolicy, ConditionTypeReady, metav1.ConditionFalse, ReasonFailed, err.Error())
		r.Recorder.Event(&updatedPolicy, EventTypeWarning, ReasonFailed, fmt.Sprintf("Cleanup failed: %v", err))
	} else {
		updatedPolicy.Status.Message = fmt.Sprintf("Cleanup completed successfully. Resources scanned: %d, cleaned: %d", 
			stats.ResourcesScanned, stats.ResourcesCleaned)
		r.updateCondition(&updatedPolicy, ConditionTypeReady, metav1.ConditionTrue, ReasonSucceeded, "Cleanup completed successfully")
		r.Recorder.Event(&updatedPolicy, EventTypeNormal, ReasonSucceeded, 
			fmt.Sprintf("Cleanup completed. Scanned: %d, Cleaned: %d", stats.ResourcesScanned, stats.ResourcesCleaned))
	}

	// Calculate next run
	if updatedPolicy.Spec.Schedule != "" {
		if schedule, parseErr := cron.ParseStandard(updatedPolicy.Spec.Schedule); parseErr == nil {
			nextRun := schedule.Next(time.Now())
			updatedPolicy.Status.NextRun = &metav1.Time{Time: nextRun}
		}
	}

	// Update status
	if updateErr := r.Status().Update(ctx, &updatedPolicy); updateErr != nil {
		log.Error(updateErr, "Failed to update status after cleanup")
	}

	// Update metrics
	if r.metricsServer != nil {
		r.metricsServer.RecordCleanupMetrics(stats, err)
	}

	log.Info("Cleanup execution completed", 
		"duration", duration, 
		"scanned", stats.ResourcesScanned, 
		"cleaned", stats.ResourcesCleaned,
		"errors", stats.ErrorsEncountered)
}

// updateCondition updates the condition in the JanitorPolicy status
func (r *JanitorPolicyReconciler) updateCondition(janitorPolicy *opsv1alpha1.JanitorPolicy, 
	conditionType string, status metav1.ConditionStatus, reason, message string) {
	
	condition := metav1.Condition{
		Type:               conditionType,
		Status:             status,
		Reason:             reason,
		Message:            message,
		LastTransitionTime: metav1.Now(),
	}

	// Update or add condition
	found := false
	for i, existingCondition := range janitorPolicy.Status.Conditions {
		if existingCondition.Type == conditionType {
			if existingCondition.Status != status {
				janitorPolicy.Status.Conditions[i] = condition
			}
			found = true
			break
		}
	}

	if !found {
		janitorPolicy.Status.Conditions = append(janitorPolicy.Status.Conditions, condition)
	}
}

// SetupWithManager sets up the controller with the Manager.
func (r *JanitorPolicyReconciler) SetupWithManager(mgr ctrl.Manager) error {
	// Initialize cron scheduler
	r.cronScheduler = cron.New(cron.WithSeconds())
	r.cronScheduler.Start()

	// Initialize cleanup engine
	r.cleanupEngine = cleanup.NewEngine()

	// Initialize metrics server
	r.metricsServer = metrics.NewServer()

	// Set up the controller
	return ctrl.NewControllerManagedBy(mgr).
		For(&opsv1alpha1.JanitorPolicy{}).
		Owns(&corev1.Event{}).
		Complete(r)
}