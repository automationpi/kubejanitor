# Configuration Guide

This guide provides comprehensive information about configuring KubeJanitor Operator and JanitorPolicy resources.

## Operator Configuration

### Helm Chart Values

The operator can be configured through Helm chart values. Here are the key configuration options:

#### Image Configuration

```yaml
image:
  registry: ghcr.io
  repository: automationpi/kubejanitor
  tag: "v0.1.0"
  pullPolicy: IfNotPresent
```

#### Resource Configuration

```yaml
resources:
  limits:
    cpu: 500m
    memory: 256Mi
  requests:
    cpu: 100m
    memory: 64Mi
```

#### Manager Configuration

```yaml
manager:
  leaderElection: true
  leaderElectionNamespace: ""
  healthProbeBindAddress: ":8081"
  metricsBindAddress: ":8080"
  logLevel: info
  logFormat: json
```

#### Security Configuration

```yaml
podSecurityContext:
  runAsNonRoot: true
  runAsUser: 65532
  runAsGroup: 65532
  fsGroup: 65532

securityContext:
  allowPrivilegeEscalation: false
  capabilities:
    drop:
    - ALL
  readOnlyRootFilesystem: true
  runAsNonRoot: true
```

### Environment Variables

Environment variables can be configured in the Helm chart:

```yaml
extraEnvVars:
  - name: LOG_LEVEL
    value: "debug"
  - name: ENABLE_WEBHOOK
    value: "false"
  - name: CLEANUP_INTERVAL
    value: "5m"
```

## JanitorPolicy Configuration

### Basic Structure

```yaml
apiVersion: ops.k8s.io/v1alpha1
kind: JanitorPolicy
metadata:
  name: example-policy
  namespace: kubejanitor-system
spec:
  dryRun: true
  schedule: "0 2 * * *"
  cleanup: {}
  protectedLabels: []
  ignoreNamespaces: []
  backupConfig: {}
  notificationConfig: {}
```

### Global Settings

#### Dry Run Mode

```yaml
spec:
  dryRun: true  # Only simulate actions, don't actually delete resources
```

#### Scheduling

```yaml
spec:
  schedule: "0 2 * * *"  # Cron format: daily at 2 AM
```

Common schedule examples:
- `"*/30 * * * *"` - Every 30 minutes
- `"0 */6 * * *"` - Every 6 hours
- `"0 2 * * 0"` - Weekly on Sunday at 2 AM
- `"0 2 1 * *"` - Monthly on the 1st at 2 AM

### Resource-Specific Cleanup Configuration

#### PVC Cleanup

```yaml
spec:
  cleanup:
    pvc:
      enabled: true
      unusedFor: "48h"  # Duration a PVC must be unused
      ignorePatterns:    # Regex patterns to ignore
        - "^database-.*"
        - ".*-backup$"
```

#### Jobs Cleanup

```yaml
spec:
  cleanup:
    jobs:
      enabled: true
      olderThan: "24h"
      statuses: ["Failed", "Complete"]
      keepSuccessfulJobs: 3  # Keep last 3 successful jobs
      keepFailedJobs: 1      # Keep last 1 failed job
```

#### ConfigMaps and Secrets Cleanup

```yaml
spec:
  cleanup:
    configMaps:
      enabled: true
      olderThan: "168h"      # 7 days
      checkReferences: true  # Check if referenced before deletion
    
    secrets:
      enabled: true
      olderThan: "168h"
      checkReferences: true
      excludeTypes:          # Secret types to exclude
        - "kubernetes.io/service-account-token"
        - "kubernetes.io/dockercfg"
```

#### Services Cleanup

```yaml
spec:
  cleanup:
    services:
      enabled: true
      checkEndpoints: true  # Check for backing endpoints
```

#### TLS Secrets Cleanup

```yaml
spec:
  cleanup:
    tlsSecrets:
      enabled: true
      expiredOnly: true      # Only clean expired certificates
      expiringWithin: "720h" # Clean certificates expiring within 30 days
```

#### Terminating Pods Cleanup

```yaml
spec:
  cleanup:
    terminatingPods:
      enabled: true
      stuckFor: "15m"  # How long a pod can be stuck in terminating state
```

#### Crash Loop Pods Handling

```yaml
spec:
  cleanup:
    crashLoopPods:
      enabled: true
      restartThreshold: 5      # Restart count threshold
      action: "alert"          # Options: alert, restart, delete
```

#### Resource Gaps Detection

```yaml
spec:
  cleanup:
    resourceGaps:
      enabled: true
      check: ["limits", "requests"]  # What to check for
      reportOnly: true               # Only report, don't fix
```

#### RBAC Validation

```yaml
spec:
  cleanup:
    rbacCheck:
      enabled: true
      fixMode: "manual"  # Options: manual, suggest, auto
```

#### Stale Helm Releases

```yaml
spec:
  cleanup:
    staleHelmReleases:
      enabled: true
      failedOnly: true    # Only clean up failed releases
      olderThan: "72h"    # Clean releases older than 3 days
```

### Protection Mechanisms

#### Protected Labels

Resources with these labels will never be cleaned up:

```yaml
spec:
  protectedLabels:
    - "app.kubernetes.io/managed-by=Helm"
    - "janitor.k8s.io/keep=true"
    - "backup.velero.io/backup-name"
```

#### Ignored Namespaces

Namespaces to completely skip during cleanup:

```yaml
spec:
  ignoreNamespaces:
    - "kube-system"
    - "kube-public"
    - "kube-node-lease"
    - "ingress-nginx"
    - "cert-manager"
    - "monitoring"
```

### Backup Configuration

```yaml
spec:
  backupConfig:
    enabled: true
    type: "git"                    # Options: git, s3, local
    location: "git@github.com:org/backups.git"
    retentionDays: 30
```

#### Git Backup

```yaml
spec:
  backupConfig:
    type: "git"
    location: "https://github.com/org/k8s-backups.git"
    # Additional git-specific configuration
```

#### S3 Backup

```yaml
spec:
  backupConfig:
    type: "s3"
    location: "s3://my-bucket/kubejanitor-backups/"
    # S3 credentials should be provided via environment variables or IAM roles
```

#### Local Backup

```yaml
spec:
  backupConfig:
    type: "local"
    location: "/tmp/kubejanitor-backups"
```

### Notification Configuration

#### Slack Notifications

```yaml
spec:
  notificationConfig:
    slack:
      enabled: true
      webhookURL: "https://hooks.slack.com/services/..."
      channel: "#alerts"
```

#### Email Notifications

```yaml
spec:
  notificationConfig:
    email:
      enabled: true
      smtpServer: "smtp.gmail.com"
      smtpPort: 587
      username: "alerts@company.com"
      password: "app-password"  # Use Kubernetes secrets in production
      to:
        - "devops@company.com"
        - "platform@company.com"
```

#### Webhook Notifications

```yaml
spec:
  notificationConfig:
    webhook:
      enabled: true
      url: "https://api.company.com/webhooks/kubejanitor"
      headers:
        Authorization: "Bearer token123"
        Content-Type: "application/json"
```

## Advanced Configuration Examples

### Production Setup

```yaml
apiVersion: ops.k8s.io/v1alpha1
kind: JanitorPolicy
metadata:
  name: production-cleanup
  namespace: kubejanitor-system
spec:
  dryRun: false
  schedule: "0 3 * * *"  # Daily at 3 AM
  
  cleanup:
    # Conservative PVC cleanup
    pvc:
      enabled: true
      unusedFor: "168h"  # 7 days
      ignorePatterns:
        - "^database-.*"
        - ".*-persistent-.*"
    
    # Aggressive job cleanup
    jobs:
      enabled: true
      olderThan: "48h"
      statuses: ["Failed", "Complete"]
      keepSuccessfulJobs: 5
      keepFailedJobs: 2
    
    # Conservative secret/configmap cleanup
    secrets:
      enabled: true
      olderThan: "720h"  # 30 days
      checkReferences: true
      excludeTypes:
        - "kubernetes.io/service-account-token"
        - "kubernetes.io/dockercfg"
        - "kubernetes.io/dockerconfigjson"
    
    # Certificate management
    tlsSecrets:
      enabled: true
      expiredOnly: false
      expiringWithin: "168h"  # 7 days
    
    # Stuck pod cleanup
    terminatingPods:
      enabled: true
      stuckFor: "10m"
    
    # Alert on crash loops
    crashLoopPods:
      enabled: true
      restartThreshold: 3
      action: "alert"
  
  protectedLabels:
    - "app.kubernetes.io/managed-by=Helm"
    - "janitor.k8s.io/keep=true"
    - "backup.velero.io/backup-name"
    - "app.kubernetes.io/component=database"
  
  ignoreNamespaces:
    - "kube-system"
    - "kube-public"
    - "kube-node-lease"
    - "ingress-nginx"
    - "cert-manager"
    - "monitoring"
    - "logging"
    - "velero"
  
  backupConfig:
    enabled: true
    type: "git"
    location: "git@github.com:company/k8s-resource-backups.git"
    retentionDays: 90
  
  notificationConfig:
    slack:
      enabled: true
      webhookURL: "https://hooks.slack.com/services/..."
      channel: "#platform-alerts"
```

### Development/Testing Setup

```yaml
apiVersion: ops.k8s.io/v1alpha1
kind: JanitorPolicy
metadata:
  name: dev-cleanup
  namespace: kubejanitor-system
spec:
  dryRun: false
  schedule: "*/15 * * * *"  # Every 15 minutes
  
  cleanup:
    # Aggressive cleanup for dev environments
    pvc:
      enabled: true
      unusedFor: "1h"
    
    jobs:
      enabled: true
      olderThan: "2h"
      statuses: ["Failed", "Complete"]
      keepSuccessfulJobs: 1
      keepFailedJobs: 1
    
    configMaps:
      enabled: true
      olderThan: "24h"
      checkReferences: true
    
    secrets:
      enabled: true
      olderThan: "24h"
      checkReferences: true
    
    services:
      enabled: true
      checkEndpoints: true
    
    terminatingPods:
      enabled: true
      stuckFor: "5m"
  
  protectedLabels:
    - "janitor.k8s.io/keep=true"
    - "environment=production"  # Protect any prod resources that might be in dev
  
  ignoreNamespaces:
    - "kube-system"
    - "kube-public"
    - "kube-node-lease"
```

## Validation and Schema

KubeJanitor includes comprehensive validation for all configuration options:

### Duration Format

All duration fields must follow Go's duration format:
- `"1h"` - 1 hour
- `"30m"` - 30 minutes
- `"24h"` - 24 hours
- `"168h"` - 7 days
- `"2h30m"` - 2 hours and 30 minutes

### Cron Schedule Format

Schedule fields must follow standard cron format:
```
┌───────────── minute (0 - 59)
│ ┌───────────── hour (0 - 23)
│ │ ┌───────────── day of the month (1 - 31)
│ │ │ ┌───────────── month (1 - 12)
│ │ │ │ ┌───────────── day of the week (0 - 6) (Sunday to Saturday)
│ │ │ │ │
│ │ │ │ │
* * * * *
```

## Best Practices

1. **Start with Dry Run**: Always begin with `dryRun: true` to understand the impact
2. **Use Conservative Timeouts**: Start with longer durations and adjust based on your environment
3. **Protect Critical Resources**: Use `protectedLabels` extensively
4. **Monitor Regularly**: Set up notifications and check logs regularly
5. **Test in Non-Production**: Validate policies in development environments first
6. **Use Backups**: Configure backup options for important resources
7. **Gradual Rollout**: Enable cleanup types incrementally

For more information, see the [Safety Guide](safety.md) and [Examples](examples/).