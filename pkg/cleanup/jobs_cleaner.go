package cleanup

import (
	"context"
	"time"

	batchv1 "k8s.io/api/batch/v1"

	opsv1alpha1 "github.com/automationpi/kubejanitor/api/v1alpha1"
)

// JobsCleaner handles cleanup of old Jobs
type JobsCleaner struct{}

// NewJobsCleaner creates a new Jobs cleaner
func NewJobsCleaner() *JobsCleaner {
	return &JobsCleaner{}
}

// Name returns the name of the cleaner
func (c *JobsCleaner) Name() string {
	return "jobs"
}

// Execute performs Jobs cleanup
func (c *JobsCleaner) Execute(ctx context.Context, cleanupCtx *Context) (*opsv1alpha1.ResourceTypeStats, error) {
	log := cleanupCtx.Logger.WithName("jobs-cleaner")
	stats := &opsv1alpha1.ResourceTypeStats{}

	config := cleanupCtx.Policy.Spec.Cleanup.Jobs
	if config == nil || !config.Enabled {
		return stats, nil
	}

	// Parse duration
	olderThan, err := time.ParseDuration(config.OlderThan)
	if err != nil {
		log.Error(err, "Failed to parse olderThan duration", "duration", config.OlderThan)
		stats.Errors++
		return stats, err
	}

	cutoffTime := time.Now().Add(-olderThan)

	// Get all Jobs
	var jobList batchv1.JobList
	if err := cleanupCtx.Client.List(ctx, &jobList); err != nil {
		log.Error(err, "Failed to list Jobs")
		stats.Errors++
		return stats, err
	}

	stats.Scanned = int32(len(jobList.Items))

	// Process each Job
	for _, job := range jobList.Items {
		if c.shouldSkipJob(&job, config, cleanupCtx) {
			log.V(1).Info("Skipping Job", "name", job.Name, "namespace", job.Namespace)
			stats.Skipped++
			continue
		}

		// Check if Job is old enough
		if job.CreationTimestamp.Time.After(cutoffTime) {
			log.V(1).Info("Job is too new, skipping", "name", job.Name, "namespace", job.Namespace, "age", time.Since(job.CreationTimestamp.Time))
			stats.Skipped++
			continue
		}

		// Check if Job status matches cleanup criteria
		if !c.shouldCleanupJobByStatus(&job, config) {
			log.V(1).Info("Job status doesn't match cleanup criteria", "name", job.Name, "namespace", job.Namespace)
			stats.Skipped++
			continue
		}

		// Job is old enough and matches status criteria
		if cleanupCtx.DryRun {
			log.Info("Would delete old Job", "name", job.Name, "namespace", job.Namespace, "age", time.Since(job.CreationTimestamp.Time))
			cleanupCtx.EventRecorder.Event(&job, "Normal", "DryRun", "Would delete old Job")
		} else {
			log.Info("Deleting old Job", "name", job.Name, "namespace", job.Namespace, "age", time.Since(job.CreationTimestamp.Time))
			if err := cleanupCtx.Client.Delete(ctx, &job); err != nil {
				log.Error(err, "Failed to delete Job", "name", job.Name, "namespace", job.Namespace)
				stats.Errors++
				cleanupCtx.EventRecorder.Event(&job, "Warning", "DeleteFailed", "Failed to delete old Job")
				continue
			}
			cleanupCtx.EventRecorder.Event(&job, "Normal", "Deleted", "Deleted old Job")
		}

		stats.Cleaned++
	}

	log.Info("Jobs cleanup completed",
		"scanned", stats.Scanned,
		"cleaned", stats.Cleaned,
		"skipped", stats.Skipped,
		"errors", stats.Errors)

	return stats, nil
}

// shouldSkipJob determines if a Job should be skipped
func (c *JobsCleaner) shouldSkipJob(job *batchv1.Job, config *opsv1alpha1.JobsCleanupConfig, cleanupCtx *Context) bool {
	// Check if namespace is ignored
	if IsNamespaceIgnored(job.Namespace, cleanupCtx.Policy.Spec.IgnoreNamespaces) {
		return true
	}

	// Check if Job has protected labels
	if IsProtected(job.Labels, cleanupCtx.Policy.Spec.ProtectedLabels) {
		return true
	}

	// Skip if Job is in terminating state
	if job.DeletionTimestamp != nil {
		return true
	}

	// Skip system Jobs
	if c.isSystemJob(job) {
		return true
	}

	return false
}

// shouldCleanupJobByStatus checks if a Job should be cleaned up based on its status
func (c *JobsCleaner) shouldCleanupJobByStatus(job *batchv1.Job, config *opsv1alpha1.JobsCleanupConfig) bool {
	if len(config.Statuses) == 0 {
		return true // If no statuses specified, clean up all jobs
	}

	for _, status := range config.Statuses {
		switch status {
		case "Complete":
			if c.isJobComplete(job) {
				return true
			}
		case "Failed":
			if c.isJobFailed(job) {
				return true
			}
		case "Active":
			if c.isJobActive(job) {
				return true
			}
		}
	}

	return false
}

// isJobComplete checks if a Job has completed successfully
func (c *JobsCleaner) isJobComplete(job *batchv1.Job) bool {
	for _, condition := range job.Status.Conditions {
		if condition.Type == batchv1.JobComplete && condition.Status == "True" {
			return true
		}
	}
	return false
}

// isJobFailed checks if a Job has failed
func (c *JobsCleaner) isJobFailed(job *batchv1.Job) bool {
	for _, condition := range job.Status.Conditions {
		if condition.Type == batchv1.JobFailed && condition.Status == "True" {
			return true
		}
	}
	return false
}

// isJobActive checks if a Job is currently active
func (c *JobsCleaner) isJobActive(job *batchv1.Job) bool {
	return job.Status.Active > 0
}

// isSystemJob checks if a Job is a system Job that should not be deleted
func (c *JobsCleaner) isSystemJob(job *batchv1.Job) bool {
	// Check for system labels
	if job.Labels != nil {
		if component := job.Labels["k8s-app"]; component != "" {
			return true
		}
		if component := job.Labels["app.kubernetes.io/component"]; component == "controller" || component == "scheduler" {
			return true
		}
	}

	return false
}
