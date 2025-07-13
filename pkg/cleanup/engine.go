package cleanup

import (
	"context"
	"fmt"
	"time"

	"github.com/go-logr/logr"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client"

	opsv1alpha1 "github.com/automationpi/kubejanitor/api/v1alpha1"
)

// Context holds the cleanup execution context
type Context struct {
	Client        client.Client
	Policy        *opsv1alpha1.JanitorPolicy
	DryRun        bool
	Logger        logr.Logger
	EventRecorder record.EventRecorder
}

// Engine handles the cleanup execution
type Engine struct {
	cleaners map[string]Cleaner
}

// Cleaner interface defines the cleanup behavior for specific resource types
type Cleaner interface {
	Name() string
	Execute(ctx context.Context, cleanupCtx *Context) (*opsv1alpha1.ResourceTypeStats, error)
}

// NewEngine creates a new cleanup engine
func NewEngine() *Engine {
	engine := &Engine{
		cleaners: make(map[string]Cleaner),
	}

	// Register all cleaners
	engine.registerCleaners()
	return engine
}

// registerCleaners registers all available cleaners
func (e *Engine) registerCleaners() {
	cleaners := []Cleaner{
		NewPVCCleaner(),
		NewJobsCleaner(),
		NewConfigMapsCleaner(),
		NewSecretsCleaner(),
		NewServicesCleaner(),
		NewTLSSecretsCleaner(),
		NewTerminatingPodsCleaner(),
		NewCrashLoopPodsCleaner(),
		NewResourceGapsChecker(),
		NewRBACChecker(),
		NewStaleHelmReleasesCleaner(),
	}

	for _, cleaner := range cleaners {
		e.cleaners[cleaner.Name()] = cleaner
	}
}

// Execute runs the cleanup process based on the policy configuration
func (e *Engine) Execute(ctx context.Context, cleanupCtx *Context) (*opsv1alpha1.CleanupStats, error) {
	log := cleanupCtx.Logger.WithName("cleanup-engine")
	
	stats := &opsv1alpha1.CleanupStats{
		ByResourceType: make(map[string]opsv1alpha1.ResourceTypeStats),
	}

	log.Info("Starting cleanup execution", "dryRun", cleanupCtx.DryRun)

	// Execute PVC cleanup
	if cleanupCtx.Policy.Spec.Cleanup.PVC != nil && cleanupCtx.Policy.Spec.Cleanup.PVC.Enabled {
		if err := e.executeCleaner(ctx, cleanupCtx, "pvc", stats); err != nil {
			log.Error(err, "PVC cleanup failed")
			stats.ErrorsEncountered++
		}
	}

	// Execute Jobs cleanup
	if cleanupCtx.Policy.Spec.Cleanup.Jobs != nil && cleanupCtx.Policy.Spec.Cleanup.Jobs.Enabled {
		if err := e.executeCleaner(ctx, cleanupCtx, "jobs", stats); err != nil {
			log.Error(err, "Jobs cleanup failed")
			stats.ErrorsEncountered++
		}
	}

	// Execute ConfigMaps cleanup
	if cleanupCtx.Policy.Spec.Cleanup.ConfigMaps != nil && cleanupCtx.Policy.Spec.Cleanup.ConfigMaps.Enabled {
		if err := e.executeCleaner(ctx, cleanupCtx, "configmaps", stats); err != nil {
			log.Error(err, "ConfigMaps cleanup failed")
			stats.ErrorsEncountered++
		}
	}

	// Execute Secrets cleanup
	if cleanupCtx.Policy.Spec.Cleanup.Secrets != nil && cleanupCtx.Policy.Spec.Cleanup.Secrets.Enabled {
		if err := e.executeCleaner(ctx, cleanupCtx, "secrets", stats); err != nil {
			log.Error(err, "Secrets cleanup failed")
			stats.ErrorsEncountered++
		}
	}

	// Execute Services cleanup
	if cleanupCtx.Policy.Spec.Cleanup.Services != nil && cleanupCtx.Policy.Spec.Cleanup.Services.Enabled {
		if err := e.executeCleaner(ctx, cleanupCtx, "services", stats); err != nil {
			log.Error(err, "Services cleanup failed")
			stats.ErrorsEncountered++
		}
	}

	// Execute TLS Secrets cleanup
	if cleanupCtx.Policy.Spec.Cleanup.TLSSecrets != nil && cleanupCtx.Policy.Spec.Cleanup.TLSSecrets.Enabled {
		if err := e.executeCleaner(ctx, cleanupCtx, "tlssecrets", stats); err != nil {
			log.Error(err, "TLS Secrets cleanup failed")
			stats.ErrorsEncountered++
		}
	}

	// Execute Terminating Pods cleanup
	if cleanupCtx.Policy.Spec.Cleanup.TerminatingPods != nil && cleanupCtx.Policy.Spec.Cleanup.TerminatingPods.Enabled {
		if err := e.executeCleaner(ctx, cleanupCtx, "terminatingpods", stats); err != nil {
			log.Error(err, "Terminating Pods cleanup failed")
			stats.ErrorsEncountered++
		}
	}

	// Execute Crash Loop Pods handling
	if cleanupCtx.Policy.Spec.Cleanup.CrashLoopPods != nil && cleanupCtx.Policy.Spec.Cleanup.CrashLoopPods.Enabled {
		if err := e.executeCleaner(ctx, cleanupCtx, "crashlooppods", stats); err != nil {
			log.Error(err, "Crash Loop Pods handling failed")
			stats.ErrorsEncountered++
		}
	}

	// Execute Resource Gaps check
	if cleanupCtx.Policy.Spec.Cleanup.ResourceGaps != nil && cleanupCtx.Policy.Spec.Cleanup.ResourceGaps.Enabled {
		if err := e.executeCleaner(ctx, cleanupCtx, "resourcegaps", stats); err != nil {
			log.Error(err, "Resource Gaps check failed")
			stats.ErrorsEncountered++
		}
	}

	// Execute RBAC check
	if cleanupCtx.Policy.Spec.Cleanup.RBACCheck != nil && cleanupCtx.Policy.Spec.Cleanup.RBACCheck.Enabled {
		if err := e.executeCleaner(ctx, cleanupCtx, "rbaccheck", stats); err != nil {
			log.Error(err, "RBAC check failed")
			stats.ErrorsEncountered++
		}
	}

	// Execute Stale Helm Releases cleanup
	if cleanupCtx.Policy.Spec.Cleanup.StaleHelmReleases != nil && cleanupCtx.Policy.Spec.Cleanup.StaleHelmReleases.Enabled {
		if err := e.executeCleaner(ctx, cleanupCtx, "stalehelm", stats); err != nil {
			log.Error(err, "Stale Helm Releases cleanup failed")
			stats.ErrorsEncountered++
		}
	}

	log.Info("Cleanup execution completed", 
		"totalScanned", stats.ResourcesScanned,
		"totalCleaned", stats.ResourcesCleaned,
		"totalErrors", stats.ErrorsEncountered)

	return stats, nil
}

// executeCleaner executes a specific cleaner and updates stats
func (e *Engine) executeCleaner(ctx context.Context, cleanupCtx *Context, cleanerName string, stats *opsv1alpha1.CleanupStats) error {
	cleaner, exists := e.cleaners[cleanerName]
	if !exists {
		return fmt.Errorf("cleaner %s not found", cleanerName)
	}

	log := cleanupCtx.Logger.WithName(fmt.Sprintf("cleaner-%s", cleanerName))
	log.Info("Executing cleaner")

	start := time.Now()
	resourceStats, err := cleaner.Execute(ctx, cleanupCtx)
	duration := time.Since(start)

	if resourceStats != nil {
		stats.ResourcesScanned += resourceStats.Scanned
		stats.ResourcesCleaned += resourceStats.Cleaned
		stats.ErrorsEncountered += resourceStats.Errors
		stats.ByResourceType[cleanerName] = *resourceStats
	}

	log.Info("Cleaner execution completed", 
		"duration", duration,
		"scanned", resourceStats.Scanned,
		"cleaned", resourceStats.Cleaned,
		"errors", resourceStats.Errors,
		"skipped", resourceStats.Skipped)

	return err
}

// IsProtected checks if a resource is protected based on labels
func IsProtected(labels map[string]string, protectedLabels []string) bool {
	if labels == nil {
		return false
	}

	for _, protectedLabel := range protectedLabels {
		if value, exists := labels[protectedLabel]; exists && value == "true" {
			return true
		}
		// Also check for label key existence (for labels like app.kubernetes.io/managed-by=Helm)
		if _, exists := labels[protectedLabel]; exists {
			return true
		}
	}

	return false
}

// IsNamespaceIgnored checks if a namespace should be ignored
func IsNamespaceIgnored(namespace string, ignoreNamespaces []string) bool {
	for _, ignored := range ignoreNamespaces {
		if namespace == ignored {
			return true
		}
	}
	return false
}