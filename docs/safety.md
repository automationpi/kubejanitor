# Safety Guide

KubeJanitor is designed with safety as the top priority. This guide covers all the safety mechanisms and best practices to ensure your cluster remains stable while cleaning up unused resources.

## Built-in Safety Mechanisms

### 1. Dry Run Mode (Default)

**Description**: By default, all policies run in dry-run mode, which means they only simulate actions without actually deleting resources.

**Configuration**:
```yaml
spec:
  dryRun: true  # Default value
```

**Benefits**:
- See what would be cleaned up without risk
- Validate policy configuration
- Build confidence before enabling actual cleanup

**Best Practice**: Always start with dry-run mode and monitor logs before switching to active mode.

### 2. Protected Labels

**Description**: Resources with specific labels are automatically protected from cleanup.

**Default Protected Labels**:
```yaml
spec:
  protectedLabels:
    - "app.kubernetes.io/managed-by=Helm"
    - "janitor.k8s.io/keep=true"
```

**Custom Protection**:
```yaml
spec:
  protectedLabels:
    - "app.kubernetes.io/managed-by=Helm"
    - "janitor.k8s.io/keep=true"
    - "backup.velero.io/backup-name"
    - "environment=production"
    - "criticality=high"
```

**Best Practice**: Label all critical resources with protection labels.

### 3. Namespace Exclusion

**Description**: Entire namespaces can be excluded from cleanup operations.

**Recommended Exclusions**:
```yaml
spec:
  ignoreNamespaces:
    - "kube-system"
    - "kube-public"
    - "kube-node-lease"
    - "ingress-nginx"
    - "cert-manager"
    - "monitoring"
    - "logging"
    - "velero"
    - "istio-system"
    - "linkerd"
```

**Best Practice**: Always exclude system namespaces and any namespace containing critical infrastructure.

### 4. Reference Checking

**Description**: For ConfigMaps and Secrets, the operator checks if they are referenced by other resources before deletion.

**Configuration**:
```yaml
spec:
  cleanup:
    configMaps:
      checkReferences: true  # Default
    secrets:
      checkReferences: true  # Default
```

**Reference Types Checked**:
- Environment variables (`env`, `envFrom`)
- Volume mounts (`volumes`)
- Image pull secrets
- Service account secrets
- Webhook configurations

### 5. Time-based Safeguards

**Description**: Resources must meet age requirements before being eligible for cleanup.

**Conservative Defaults**:
```yaml
spec:
  cleanup:
    pvc:
      unusedFor: "48h"      # PVCs unused for 48+ hours
    jobs:
      olderThan: "24h"      # Jobs older than 24 hours
    secrets:
      olderThan: "168h"     # Secrets older than 7 days
```

**Best Practice**: Use conservative time windows, especially in production.

### 6. Resource Type Exclusions

**Description**: Certain types of secrets and other critical resources are excluded by default.

**Default Secret Exclusions**:
```yaml
spec:
  cleanup:
    secrets:
      excludeTypes:
        - "kubernetes.io/service-account-token"
        - "kubernetes.io/dockercfg"
        - "kubernetes.io/dockerconfigjson"
        - "bootstrap.kubernetes.io/token"
```

### 7. Audit Logging

**Description**: All cleanup actions are logged with detailed information for audit purposes.

**Log Information Includes**:
- Resource details (name, namespace, type)
- Reason for cleanup
- Timestamp
- Policy that triggered the action
- Dry-run vs actual action

**Example Log Entry**:
```json
{
  "level": "info",
  "timestamp": "2024-01-15T10:30:00Z",
  "msg": "Resource cleaned up",
  "policy": "production-cleanup",
  "resource": "unused-pvc-1",
  "namespace": "app-namespace",
  "type": "PersistentVolumeClaim",
  "reason": "Unused for 72h",
  "dryRun": false
}
```

### 8. Event Generation

**Description**: Kubernetes events are generated for all cleanup actions.

**Event Types**:
- `CleanupCandidateFound` - Resource identified for cleanup
- `CleanupSimulated` - Dry-run action performed
- `CleanupCompleted` - Actual cleanup performed
- `CleanupSkipped` - Resource skipped due to protection
- `CleanupFailed` - Cleanup action failed

**Viewing Events**:
```bash
# View all janitor events
kubectl get events --field-selector source=kubejanitor-operator

# View events for specific resource
kubectl describe pvc unused-pvc-1
```

## Safety Best Practices

### 1. Phased Rollout Strategy

#### Phase 1: Observation (1-2 weeks)
```yaml
spec:
  dryRun: true
  schedule: "0 2 * * *"  # Daily
  cleanup:
    # Start with safest cleanup types
    jobs:
      enabled: true
      olderThan: "72h"
      statuses: ["Failed"]
```

#### Phase 2: Conservative Cleanup (2-4 weeks)
```yaml
spec:
  dryRun: false
  cleanup:
    jobs:
      enabled: true
      olderThan: "48h"
      statuses: ["Failed", "Complete"]
      keepSuccessfulJobs: 5
    terminatingPods:
      enabled: true
      stuckFor: "30m"
```

#### Phase 3: Full Deployment
```yaml
spec:
  dryRun: false
  cleanup:
    pvc:
      enabled: true
      unusedFor: "168h"  # Conservative 7 days
    # ... other cleanup types
```

### 2. Environment-Specific Policies

#### Production Policy
```yaml
metadata:
  name: production-policy
spec:
  dryRun: false
  schedule: "0 3 * * *"  # Once daily, off-hours
  cleanup:
    # Very conservative settings
    pvc:
      unusedFor: "720h"   # 30 days
    jobs:
      olderThan: "168h"   # 7 days
      keepSuccessfulJobs: 10
```

#### Development Policy
```yaml
metadata:
  name: development-policy
spec:
  dryRun: false
  schedule: "*/30 * * * *"  # Every 30 minutes
  cleanup:
    # More aggressive settings
    pvc:
      unusedFor: "2h"
    jobs:
      olderThan: "4h"
      keepSuccessfulJobs: 2
```

### 3. Monitoring and Alerting

#### Essential Metrics to Monitor
- Number of resources scanned
- Number of resources cleaned
- Number of errors encountered
- Cleanup duration
- Protected resources skipped

#### Recommended Alerts
```yaml
# Alert on high error rate
- alert: KubeJanitorHighErrorRate
  expr: kubejanitor_errors_total / kubejanitor_scanned_total > 0.1
  for: 5m

# Alert on unexpected high cleanup volume
- alert: KubeJanitorHighCleanupVolume
  expr: kubejanitor_cleaned_total > 100
  for: 1m
```

### 4. Backup Strategy

#### Git-based Backup
```yaml
spec:
  backupConfig:
    enabled: true
    type: "git"
    location: "git@github.com:company/k8s-backups.git"
    retentionDays: 90
```

#### S3-based Backup
```yaml
spec:
  backupConfig:
    enabled: true
    type: "s3"
    location: "s3://company-k8s-backups/kubejanitor/"
    retentionDays: 30
```

### 5. Testing and Validation

#### Pre-deployment Testing
```bash
# Test policy in development cluster
kubectl apply -f policy-dev.yaml

# Monitor for 24 hours
kubectl logs -f deployment/kubejanitor-operator -n kubejanitor-system

# Check events and metrics
kubectl get events --field-selector source=kubejanitor-operator
```

#### Canary Deployment
```yaml
# Start with limited scope
spec:
  cleanup:
    jobs:
      enabled: true
      # Only clean up test jobs initially
      ignorePatterns:
        - "^(?!test-).*"  # Only clean resources starting with "test-"
```

## Recovery Procedures

### 1. Emergency Stop

```bash
# Pause all cleanup by setting dry-run mode
kubectl patch janitorpolicy production-policy -p '{"spec":{"dryRun":true}}'

# Or scale down the operator
kubectl scale deployment kubejanitor-operator --replicas=0 -n kubejanitor-system
```

### 2. Resource Recovery

#### From Git Backup
```bash
# Clone backup repository
git clone git@github.com:company/k8s-backups.git
cd k8s-backups

# Find and restore specific resource
kubectl apply -f backups/2024-01-15/pvc-important-data.yaml
```

#### From Event Logs
```bash
# Find deletion events
kubectl get events --field-selector reason=CleanupCompleted,type=Warning

# Manually recreate if backup not available
# (Use the event details to reconstruct the resource)
```

### 3. Policy Rollback

```bash
# Revert to previous policy version
kubectl rollout undo deployment kubejanitor-operator -n kubejanitor-system

# Or apply previous policy configuration
kubectl apply -f policy-backup.yaml
```

## Common Pitfalls and How to Avoid Them

### 1. Overly Aggressive Timeouts

**Problem**: Setting very short timeouts that clean up resources still in use.

**Solution**:
```yaml
# Bad
spec:
  cleanup:
    pvc:
      unusedFor: "5m"  # Too aggressive

# Good
spec:
  cleanup:
    pvc:
      unusedFor: "48h"  # Conservative
```

### 2. Missing Protection Labels

**Problem**: Critical resources without protection labels get cleaned up.

**Solution**:
```bash
# Add protection labels to critical resources
kubectl label pvc important-database-pvc janitor.k8s.io/keep=true
kubectl label secret production-tls-cert janitor.k8s.io/keep=true
```

### 3. Not Testing in Non-Production

**Problem**: Deploying policies directly to production without testing.

**Solution**: Always test policies in development/staging environments first.

### 4. Ignoring Events and Logs

**Problem**: Not monitoring what the operator is doing.

**Solution**: Set up proper monitoring and alerting for janitor events.

## Security Considerations

### 1. RBAC Permissions

The operator requires extensive permissions. Ensure:
- Service account has minimal required permissions
- Regular audit of RBAC roles
- Use of admission controllers to prevent privilege escalation

### 2. Backup Security

```yaml
# Use secrets for sensitive backup configuration
spec:
  backupConfig:
    enabled: true
    type: "git"
    location: "git@github.com:company/k8s-backups.git"
    # SSH key should be provided via Kubernetes secret
```

### 3. Network Policies

Consider implementing network policies to restrict operator communication:

```yaml
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: kubejanitor-operator-netpol
spec:
  podSelector:
    matchLabels:
      app.kubernetes.io/name: kubejanitor-operator
  policyTypes:
  - Egress
  egress:
  - to: []  # Allow egress to Kubernetes API server only
```

## Compliance and Governance

### 1. Change Management

- Document all policy changes
- Use GitOps for policy management
- Require approvals for production policy changes

### 2. Audit Requirements

- Enable audit logging for all cleanup actions
- Retain logs for compliance periods
- Regular review of cleanup activities

### 3. Documentation

Maintain documentation for:
- Deployed policies and their rationale
- Emergency procedures
- Recovery processes
- Regular review schedules

Remember: The goal is automated cleanup with zero risk to production workloads. When in doubt, err on the side of caution.