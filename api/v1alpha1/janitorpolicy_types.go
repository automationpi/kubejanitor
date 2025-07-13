package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// JanitorPolicySpec defines the desired state of JanitorPolicy
type JanitorPolicySpec struct {
	// DryRun mode - when true, only simulate actions without performing them
	// +kubebuilder:default=true
	DryRun bool `json:"dryRun,omitempty"`

	// Schedule defines when cleanup should run (cron format)
	// +kubebuilder:validation:Pattern=`^(\*|([0-9]|1[0-9]|2[0-9]|3[0-9]|4[0-9]|5[0-9])|\*\/([0-9]|1[0-9]|2[0-9]|3[0-9]|4[0-9]|5[0-9])) (\*|([0-9]|1[0-9]|2[0-3])|\*\/([0-9]|1[0-9]|2[0-3])) (\*|([1-9]|1[0-9]|2[0-9]|3[0-1])|\*\/([1-9]|1[0-9]|2[0-9]|3[0-1])) (\*|([1-9]|1[0-2])|\*\/([1-9]|1[0-2])) (\*|([0-6])|\*\/([0-6]))$`
	Schedule string `json:"schedule,omitempty"`

	// Cleanup configuration for different resource types
	Cleanup CleanupConfig `json:"cleanup,omitempty"`

	// ProtectedLabels - resources with these labels will never be cleaned up
	ProtectedLabels []string `json:"protectedLabels,omitempty"`

	// IgnoreNamespaces - namespaces to completely skip during cleanup
	IgnoreNamespaces []string `json:"ignoreNamespaces,omitempty"`

	// BackupConfig - optional backup configuration before deletion
	BackupConfig *BackupConfig `json:"backupConfig,omitempty"`

	// NotificationConfig - optional notification settings
	NotificationConfig *NotificationConfig `json:"notificationConfig,omitempty"`
}

// CleanupConfig defines cleanup configuration for different resource types
type CleanupConfig struct {
	// PVC cleanup configuration
	PVC *PVCCleanupConfig `json:"pvc,omitempty"`

	// Jobs cleanup configuration
	Jobs *JobsCleanupConfig `json:"jobs,omitempty"`

	// ConfigMaps cleanup configuration
	ConfigMaps *ConfigMapsCleanupConfig `json:"configMaps,omitempty"`

	// Secrets cleanup configuration
	Secrets *SecretsCleanupConfig `json:"secrets,omitempty"`

	// Services cleanup configuration
	Services *ServicesCleanupConfig `json:"services,omitempty"`

	// TLSSecrets cleanup configuration
	TLSSecrets *TLSSecretsCleanupConfig `json:"tlsSecrets,omitempty"`

	// TerminatingPods cleanup configuration
	TerminatingPods *TerminatingPodsCleanupConfig `json:"terminatingPods,omitempty"`

	// StaleHelmReleases cleanup configuration
	StaleHelmReleases *StaleHelmReleasesCleanupConfig `json:"staleHelmReleases,omitempty"`

	// ResourceGaps configuration
	ResourceGaps *ResourceGapsConfig `json:"resourceGaps,omitempty"`

	// CrashLoopPods configuration
	CrashLoopPods *CrashLoopPodsConfig `json:"crashLoopPods,omitempty"`

	// RBACCheck configuration
	RBACCheck *RBACCheckConfig `json:"rbacCheck,omitempty"`
}

// PVCCleanupConfig defines PVC cleanup parameters
type PVCCleanupConfig struct {
	// Enabled - whether PVC cleanup is enabled
	Enabled bool `json:"enabled,omitempty"`

	// UnusedFor - how long a PVC must be unused before cleanup
	// +kubebuilder:validation:Pattern=`^([0-9]+(\.[0-9]+)?(ns|us|µs|ms|s|m|h))+$`
	UnusedFor string `json:"unusedFor,omitempty"`

	// IgnorePatterns - PVC name patterns to ignore
	IgnorePatterns []string `json:"ignorePatterns,omitempty"`
}

// JobsCleanupConfig defines Jobs cleanup parameters
type JobsCleanupConfig struct {
	// Enabled - whether Jobs cleanup is enabled
	Enabled bool `json:"enabled,omitempty"`

	// OlderThan - delete jobs older than this duration
	// +kubebuilder:validation:Pattern=`^([0-9]+(\.[0-9]+)?(ns|us|µs|ms|s|m|h))+$`
	OlderThan string `json:"olderThan,omitempty"`

	// Statuses - job statuses to clean up
	// +kubebuilder:validation:Enum=Failed;Complete;Active
	Statuses []string `json:"statuses,omitempty"`

	// KeepSuccessfulJobs - number of successful jobs to keep
	KeepSuccessfulJobs *int32 `json:"keepSuccessfulJobs,omitempty"`

	// KeepFailedJobs - number of failed jobs to keep
	KeepFailedJobs *int32 `json:"keepFailedJobs,omitempty"`
}

// ConfigMapsCleanupConfig defines ConfigMaps cleanup parameters
type ConfigMapsCleanupConfig struct {
	// Enabled - whether ConfigMaps cleanup is enabled
	Enabled bool `json:"enabled,omitempty"`

	// OlderThan - delete configmaps older than this duration
	// +kubebuilder:validation:Pattern=`^([0-9]+(\.[0-9]+)?(ns|us|µs|ms|s|m|h))+$`
	OlderThan string `json:"olderThan,omitempty"`

	// CheckReferences - whether to check for references before deletion
	// +kubebuilder:default=true
	CheckReferences bool `json:"checkReferences,omitempty"`
}

// SecretsCleanupConfig defines Secrets cleanup parameters
type SecretsCleanupConfig struct {
	// Enabled - whether Secrets cleanup is enabled
	Enabled bool `json:"enabled,omitempty"`

	// OlderThan - delete secrets older than this duration
	// +kubebuilder:validation:Pattern=`^([0-9]+(\.[0-9]+)?(ns|us|µs|ms|s|m|h))+$`
	OlderThan string `json:"olderThan,omitempty"`

	// CheckReferences - whether to check for references before deletion
	// +kubebuilder:default=true
	CheckReferences bool `json:"checkReferences,omitempty"`

	// ExcludeTypes - secret types to exclude from cleanup
	ExcludeTypes []string `json:"excludeTypes,omitempty"`
}

// ServicesCleanupConfig defines Services cleanup parameters
type ServicesCleanupConfig struct {
	// Enabled - whether Services cleanup is enabled
	Enabled bool `json:"enabled,omitempty"`

	// CheckEndpoints - whether to check for backing endpoints
	// +kubebuilder:default=true
	CheckEndpoints bool `json:"checkEndpoints,omitempty"`
}

// TLSSecretsCleanupConfig defines TLS Secrets cleanup parameters
type TLSSecretsCleanupConfig struct {
	// Enabled - whether TLS Secrets cleanup is enabled
	Enabled bool `json:"enabled,omitempty"`

	// ExpiredOnly - only clean up expired certificates
	// +kubebuilder:default=true
	ExpiredOnly bool `json:"expiredOnly,omitempty"`

	// ExpiringWithin - clean up certificates expiring within this duration
	// +kubebuilder:validation:Pattern=`^([0-9]+(\.[0-9]+)?(ns|us|µs|ms|s|m|h))+$`
	ExpiringWithin string `json:"expiringWithin,omitempty"`
}

// TerminatingPodsCleanupConfig defines terminating Pods cleanup parameters
type TerminatingPodsCleanupConfig struct {
	// Enabled - whether terminating Pods cleanup is enabled
	Enabled bool `json:"enabled,omitempty"`

	// StuckFor - how long a pod can be stuck in terminating state
	// +kubebuilder:validation:Pattern=`^([0-9]+(\.[0-9]+)?(ns|us|µs|ms|s|m|h))+$`
	StuckFor string `json:"stuckFor,omitempty"`
}

// StaleHelmReleasesCleanupConfig defines Helm releases cleanup parameters
type StaleHelmReleasesCleanupConfig struct {
	// Enabled - whether Helm releases cleanup is enabled
	Enabled bool `json:"enabled,omitempty"`

	// FailedOnly - only clean up failed releases
	// +kubebuilder:default=true
	FailedOnly bool `json:"failedOnly,omitempty"`

	// OlderThan - delete releases older than this duration
	// +kubebuilder:validation:Pattern=`^([0-9]+(\.[0-9]+)?(ns|us|µs|ms|s|m|h))+$`
	OlderThan string `json:"olderThan,omitempty"`
}

// ResourceGapsConfig defines resource gaps detection parameters
type ResourceGapsConfig struct {
	// Enabled - whether resource gaps detection is enabled
	Enabled bool `json:"enabled,omitempty"`

	// Check - what to check for (limits, requests, or both)
	// +kubebuilder:validation:Enum=limits;requests;both
	Check []string `json:"check,omitempty"`

	// ReportOnly - only report gaps, don't attempt to fix
	// +kubebuilder:default=true
	ReportOnly bool `json:"reportOnly,omitempty"`
}

// CrashLoopPodsConfig defines crash loop pods handling parameters
type CrashLoopPodsConfig struct {
	// Enabled - whether crash loop pods handling is enabled
	Enabled bool `json:"enabled,omitempty"`

	// RestartThreshold - restart threshold to consider a pod in crash loop
	// +kubebuilder:default=5
	RestartThreshold int32 `json:"restartThreshold,omitempty"`

	// Action - what action to take (restart, alert, delete)
	// +kubebuilder:validation:Enum=restart;alert;delete
	// +kubebuilder:default=alert
	Action string `json:"action,omitempty"`
}

// RBACCheckConfig defines RBAC validation parameters
type RBACCheckConfig struct {
	// Enabled - whether RBAC check is enabled
	Enabled bool `json:"enabled,omitempty"`

	// FixMode - how to handle misconfigurations (manual, suggest, auto)
	// +kubebuilder:validation:Enum=manual;suggest;auto
	// +kubebuilder:default=manual
	FixMode string `json:"fixMode,omitempty"`
}

// BackupConfig defines backup configuration
type BackupConfig struct {
	// Enabled - whether backup is enabled
	Enabled bool `json:"enabled,omitempty"`

	// Type - backup type (git, s3, local)
	// +kubebuilder:validation:Enum=git;s3;local
	Type string `json:"type,omitempty"`

	// Location - backup location (URL, path, etc.)
	Location string `json:"location,omitempty"`

	// RetentionDays - how long to keep backups
	RetentionDays int32 `json:"retentionDays,omitempty"`
}

// NotificationConfig defines notification configuration
type NotificationConfig struct {
	// Slack configuration
	Slack *SlackConfig `json:"slack,omitempty"`

	// Email configuration
	Email *EmailConfig `json:"email,omitempty"`

	// Webhook configuration
	Webhook *WebhookConfig `json:"webhook,omitempty"`
}

// SlackConfig defines Slack notification configuration
type SlackConfig struct {
	// Enabled - whether Slack notifications are enabled
	Enabled bool `json:"enabled,omitempty"`

	// WebhookURL - Slack webhook URL
	WebhookURL string `json:"webhookURL,omitempty"`

	// Channel - Slack channel to send notifications to
	Channel string `json:"channel,omitempty"`
}

// EmailConfig defines email notification configuration
type EmailConfig struct {
	// Enabled - whether email notifications are enabled
	Enabled bool `json:"enabled,omitempty"`

	// SMTP server configuration
	SMTPServer string `json:"smtpServer,omitempty"`
	SMTPPort   int32  `json:"smtpPort,omitempty"`
	Username   string `json:"username,omitempty"`
	Password   string `json:"password,omitempty"`

	// Recipients
	To []string `json:"to,omitempty"`
}

// WebhookConfig defines webhook notification configuration
type WebhookConfig struct {
	// Enabled - whether webhook notifications are enabled
	Enabled bool `json:"enabled,omitempty"`

	// URL - webhook URL
	URL string `json:"url,omitempty"`

	// Headers - custom headers to send
	Headers map[string]string `json:"headers,omitempty"`
}

// JanitorPolicyStatus defines the observed state of JanitorPolicy
type JanitorPolicyStatus struct {
	// LastRun - timestamp of the last cleanup run
	LastRun *metav1.Time `json:"lastRun,omitempty"`

	// NextRun - timestamp of the next scheduled cleanup run
	NextRun *metav1.Time `json:"nextRun,omitempty"`

	// Phase - current phase of the policy
	// +kubebuilder:validation:Enum=Active;Paused;Error
	Phase string `json:"phase,omitempty"`

	// Message - human readable message about the current status
	Message string `json:"message,omitempty"`

	// Stats - cleanup statistics from the last run
	Stats *CleanupStats `json:"stats,omitempty"`

	// Conditions - conditions array
	Conditions []metav1.Condition `json:"conditions,omitempty"`
}

// CleanupStats defines cleanup statistics
type CleanupStats struct {
	// ResourcesScanned - total number of resources scanned
	ResourcesScanned int32 `json:"resourcesScanned,omitempty"`

	// ResourcesCleaned - total number of resources cleaned up
	ResourcesCleaned int32 `json:"resourcesCleaned,omitempty"`

	// ErrorsEncountered - number of errors encountered
	ErrorsEncountered int32 `json:"errorsEncountered,omitempty"`

	// Duration - how long the cleanup took
	Duration string `json:"duration,omitempty"`

	// ByResourceType - breakdown by resource type
	ByResourceType map[string]ResourceTypeStats `json:"byResourceType,omitempty"`
}

// ResourceTypeStats defines statistics for a specific resource type
type ResourceTypeStats struct {
	// Scanned - number of resources scanned
	Scanned int32 `json:"scanned,omitempty"`

	// Cleaned - number of resources cleaned
	Cleaned int32 `json:"cleaned,omitempty"`

	// Skipped - number of resources skipped
	Skipped int32 `json:"skipped,omitempty"`

	// Errors - number of errors
	Errors int32 `json:"errors,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:printcolumn:name="Phase",type=string,JSONPath=`.status.phase`
//+kubebuilder:printcolumn:name="Schedule",type=string,JSONPath=`.spec.schedule`
//+kubebuilder:printcolumn:name="DryRun",type=boolean,JSONPath=`.spec.dryRun`
//+kubebuilder:printcolumn:name="Last Run",type=date,JSONPath=`.status.lastRun`
//+kubebuilder:printcolumn:name="Age",type=date,JSONPath=`.metadata.creationTimestamp`

// JanitorPolicy is the Schema for the janitorpolicies API
type JanitorPolicy struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   JanitorPolicySpec   `json:"spec,omitempty"`
	Status JanitorPolicyStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// JanitorPolicyList contains a list of JanitorPolicy
type JanitorPolicyList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []JanitorPolicy `json:"items"`
}

func init() {
	SchemeBuilder.Register(&JanitorPolicy{}, &JanitorPolicyList{})
}
