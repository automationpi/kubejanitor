package cleanup

import (
	"context"
	"regexp"
	"strings"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	opsv1alpha1 "github.com/automationpi/kubejanitor/api/v1alpha1"
)

// PVCCleaner handles cleanup of unused PersistentVolumeClaims
type PVCCleaner struct{}

// NewPVCCleaner creates a new PVC cleaner
func NewPVCCleaner() *PVCCleaner {
	return &PVCCleaner{}
}

// Name returns the name of the cleaner
func (c *PVCCleaner) Name() string {
	return "pvc"
}

// Execute performs PVC cleanup
func (c *PVCCleaner) Execute(ctx context.Context, cleanupCtx *Context) (*opsv1alpha1.ResourceTypeStats, error) {
	log := cleanupCtx.Logger.WithName("pvc-cleaner")
	stats := &opsv1alpha1.ResourceTypeStats{}

	config := cleanupCtx.Policy.Spec.Cleanup.PVC
	if config == nil || !config.Enabled {
		return stats, nil
	}

	// Parse duration
	unusedFor, err := time.ParseDuration(config.UnusedFor)
	if err != nil {
		log.Error(err, "Failed to parse unusedFor duration", "duration", config.UnusedFor)
		stats.Errors++
		return stats, err
	}

	cutoffTime := time.Now().Add(-unusedFor)

	// Get all PVCs
	var pvcList corev1.PersistentVolumeClaimList
	if err := cleanupCtx.Client.List(ctx, &pvcList); err != nil {
		log.Error(err, "Failed to list PVCs")
		stats.Errors++
		return stats, err
	}

	stats.Scanned = int32(len(pvcList.Items))

	// Get all pods to check PVC usage
	var podList corev1.PodList
	if err := cleanupCtx.Client.List(ctx, &podList); err != nil {
		log.Error(err, "Failed to list pods")
		stats.Errors++
		return stats, err
	}

	// Build map of used PVCs
	usedPVCs := make(map[string]bool)
	for _, pod := range podList.Items {
		for _, volume := range pod.Spec.Volumes {
			if volume.PersistentVolumeClaim != nil {
				key := pod.Namespace + "/" + volume.PersistentVolumeClaim.ClaimName
				usedPVCs[key] = true
			}
		}
	}

	// Process each PVC
	for _, pvc := range pvcList.Items {
		if c.shouldSkipPVC(&pvc, config, cleanupCtx) {
			log.V(1).Info("Skipping PVC", "name", pvc.Name, "namespace", pvc.Namespace)
			stats.Skipped++
			continue
		}

		pvcKey := pvc.Namespace + "/" + pvc.Name

		// Check if PVC is being used
		if usedPVCs[pvcKey] {
			log.V(1).Info("PVC is in use, skipping", "name", pvc.Name, "namespace", pvc.Namespace)
			stats.Skipped++
			continue
		}

		// Check if PVC is old enough
		if pvc.CreationTimestamp.Time.After(cutoffTime) {
			log.V(1).Info("PVC is too new, skipping", "name", pvc.Name, "namespace", pvc.Namespace, "age", time.Since(pvc.CreationTimestamp.Time))
			stats.Skipped++
			continue
		}

		// PVC is unused and old enough to be cleaned
		if cleanupCtx.DryRun {
			log.Info("Would delete unused PVC", "name", pvc.Name, "namespace", pvc.Namespace, "age", time.Since(pvc.CreationTimestamp.Time))
			cleanupCtx.EventRecorder.Event(&pvc, "Normal", "DryRun", "Would delete unused PVC")
		} else {
			log.Info("Deleting unused PVC", "name", pvc.Name, "namespace", pvc.Namespace, "age", time.Since(pvc.CreationTimestamp.Time))
			if err := cleanupCtx.Client.Delete(ctx, &pvc); err != nil {
				log.Error(err, "Failed to delete PVC", "name", pvc.Name, "namespace", pvc.Namespace)
				stats.Errors++
				cleanupCtx.EventRecorder.Event(&pvc, "Warning", "DeleteFailed", "Failed to delete unused PVC")
				continue
			}
			cleanupCtx.EventRecorder.Event(&pvc, "Normal", "Deleted", "Deleted unused PVC")
		}

		stats.Cleaned++
	}

	log.Info("PVC cleanup completed", 
		"scanned", stats.Scanned, 
		"cleaned", stats.Cleaned, 
		"skipped", stats.Skipped, 
		"errors", stats.Errors)

	return stats, nil
}

// shouldSkipPVC determines if a PVC should be skipped
func (c *PVCCleaner) shouldSkipPVC(pvc *corev1.PersistentVolumeClaim, config *opsv1alpha1.PVCCleanupConfig, cleanupCtx *Context) bool {
	// Check if namespace is ignored
	if IsNamespaceIgnored(pvc.Namespace, cleanupCtx.Policy.Spec.IgnoreNamespaces) {
		return true
	}

	// Check if PVC has protected labels
	if IsProtected(pvc.Labels, cleanupCtx.Policy.Spec.ProtectedLabels) {
		return true
	}

	// Check ignore patterns
	for _, pattern := range config.IgnorePatterns {
		if matched, _ := regexp.MatchString(pattern, pvc.Name); matched {
			return true
		}
	}

	// Skip if PVC is in terminating state
	if pvc.DeletionTimestamp != nil {
		return true
	}

	// Skip system PVCs (those with certain annotations or labels)
	if c.isSystemPVC(pvc) {
		return true
	}

	return false
}

// isSystemPVC checks if a PVC is a system PVC that should not be deleted
func (c *PVCCleaner) isSystemPVC(pvc *corev1.PersistentVolumeClaim) bool {
	// Check for system annotations
	systemAnnotations := []string{
		"volume.beta.kubernetes.io/storage-provisioner",
		"volume.kubernetes.io/storage-provisioner",
	}

	for _, annotation := range systemAnnotations {
		if _, exists := pvc.Annotations[annotation]; exists {
			// Additional checks for system storage classes
			if pvc.Spec.StorageClassName != nil {
				storageClass := *pvc.Spec.StorageClassName
				if strings.Contains(storageClass, "system") || 
				   strings.Contains(storageClass, "default") ||
				   strings.Contains(storageClass, "gp2") {
					return false // These are user PVCs, not system PVCs
				}
			}
		}
	}

	// Check for system labels
	if pvc.Labels != nil {
		if component := pvc.Labels["k8s-app"]; component != "" {
			return true
		}
		if component := pvc.Labels["app.kubernetes.io/component"]; component == "controller" || component == "scheduler" {
			return true
		}
	}

	return false
}