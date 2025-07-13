package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"

	opsv1alpha1 "github.com/automationpi/kubejanitor/api/v1alpha1"
)

// Server handles metrics collection and exposure
type Server struct {
	resourcesScanned *prometheus.CounterVec
	resourcesCleaned *prometheus.CounterVec
	errorsTotal      *prometheus.CounterVec
	cleanupDuration  *prometheus.HistogramVec
}

// NewServer creates a new metrics server
func NewServer() *Server {
	return &Server{
		resourcesScanned: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "kubejanitor_resources_scanned_total",
				Help: "Total number of resources scanned by type",
			},
			[]string{"resource_type", "namespace", "policy"},
		),
		resourcesCleaned: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "kubejanitor_resources_cleaned_total",
				Help: "Total number of resources cleaned by type",
			},
			[]string{"resource_type", "namespace", "policy"},
		),
		errorsTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "kubejanitor_errors_total",
				Help: "Total number of errors encountered",
			},
			[]string{"resource_type", "namespace", "policy", "error_type"},
		),
		cleanupDuration: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "kubejanitor_cleanup_duration_seconds",
				Help:    "Time spent cleaning up resources",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"resource_type", "namespace", "policy"},
		),
	}
}

// RecordCleanupMetrics records metrics for cleanup operations
func (s *Server) RecordCleanupMetrics(stats *opsv1alpha1.CleanupStats, err error) {
	if stats == nil {
		return
	}

	// Record overall metrics
	if err != nil {
		s.errorsTotal.WithLabelValues("overall", "", "", "cleanup_failed").Inc()
	}

	// Record per-resource-type metrics
	for resourceType, resourceStats := range stats.ByResourceType {
		s.resourcesScanned.WithLabelValues(resourceType, "", "").Add(float64(resourceStats.Scanned))
		s.resourcesCleaned.WithLabelValues(resourceType, "", "").Add(float64(resourceStats.Cleaned))

		if resourceStats.Errors > 0 {
			s.errorsTotal.WithLabelValues(resourceType, "", "", "resource_error").Add(float64(resourceStats.Errors))
		}
	}
}
