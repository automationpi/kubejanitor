# Installation Guide

This guide covers different methods to install KubeJanitor Operator in your Kubernetes cluster.

## Prerequisites

- Kubernetes 1.20+
- Helm 3.x (for Helm installation)
- kubectl configured to access your cluster

## Installation Methods

### 1. Helm Installation (Recommended)

#### Add Helm Repository

```bash
helm repo add kubejanitor https://automationpi.github.io/kubejanitor
helm repo update
```

#### Install with Default Configuration

```bash
helm install kubejanitor kubejanitor/kubejanitor-operator \
  --namespace kubejanitor-system \
  --create-namespace
```

#### Install with Custom Configuration

```bash
# Create custom values file
cat > values-custom.yaml <<EOF
dryRun: false
defaultPolicy:
  cleanup:
    pvc:
      enabled: true
      unusedFor: "48h"
    jobs:
      enabled: true
      olderThan: "24h"
EOF

# Install with custom values
helm install kubejanitor kubejanitor/kubejanitor-operator \
  --namespace kubejanitor-system \
  --create-namespace \
  --values values-custom.yaml
```

### 2. Kubectl Installation

#### Install CRDs

```bash
kubectl apply -f https://github.com/automationpi/kubejanitor/releases/latest/download/crds.yaml
```

#### Install Operator

```bash
kubectl apply -f https://github.com/automationpi/kubejanitor/releases/latest/download/operator.yaml
```

### 3. Kustomize Installation

```bash
# Create kustomization.yaml
cat > kustomization.yaml <<EOF
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization

resources:
- https://github.com/automationpi/kubejanitor/config/default

namespace: kubejanitor-system

namePrefix: my-
EOF

# Apply with kustomize
kubectl apply -k .
```

## Configuration

### Environment Variables

The operator supports configuration through environment variables:

| Variable | Description | Default |
|----------|-------------|---------|
| `LOG_LEVEL` | Log level (debug, info, warn, error) | `info` |
| `LOG_FORMAT` | Log format (json, console) | `json` |
| `LEADER_ELECTION` | Enable leader election | `true` |
| `METRICS_BIND_ADDRESS` | Metrics server bind address | `:8080` |
| `HEALTH_PROBE_BIND_ADDRESS` | Health probe bind address | `:8081` |

### RBAC Configuration

The operator requires cluster-wide permissions to manage resources. The Helm chart automatically creates the necessary RBAC resources:

- `ClusterRole` for resource management
- `ClusterRoleBinding` to bind the role to the service account
- `Role` and `RoleBinding` for leader election

### Custom Resource Definitions

KubeJanitor uses a single CRD: `JanitorPolicy`. The CRD is automatically installed with the Helm chart or can be installed separately:

```bash
kubectl apply -f https://raw.githubusercontent.com/automationpi/kubejanitor/main/config/crd/bases/ops.k8s.io_janitorpolicies.yaml
```

## Post-Installation

### Verify Installation

```bash
# Check operator deployment
kubectl get deployment -n kubejanitor-system

# Check operator logs
kubectl logs -n kubejanitor-system deployment/kubejanitor-operator

# Check CRD installation
kubectl get crd janitorpolicies.ops.k8s.io
```

### Create Your First Policy

```bash
kubectl apply -f - <<EOF
apiVersion: ops.k8s.io/v1alpha1
kind: JanitorPolicy
metadata:
  name: basic-cleanup
  namespace: kubejanitor-system
spec:
  dryRun: true
  schedule: "0 2 * * *"
  cleanup:
    jobs:
      enabled: true
      olderThan: 24h
      statuses: ["Failed", "Complete"]
    pvc:
      enabled: true
      unusedFor: 48h
  protectedLabels:
    - "janitor.k8s.io/keep=true"
  ignoreNamespaces:
    - "kube-system"
    - "kube-public"
EOF
```

### Monitor Policy Status

```bash
# Check policy status
kubectl get janitorpolicy -A

# Get detailed policy information
kubectl describe janitorpolicy basic-cleanup -n kubejanitor-system

# View cleanup events
kubectl get events -n kubejanitor-system --field-selector reason=CleanupCompleted
```

## Upgrade

### Helm Upgrade

```bash
# Update repository
helm repo update

# Upgrade to latest version
helm upgrade kubejanitor kubejanitor/kubejanitor-operator \
  --namespace kubejanitor-system
```

### Manual Upgrade

```bash
# Update CRDs (if needed)
kubectl apply -f https://github.com/automationpi/kubejanitor/releases/latest/download/crds.yaml

# Update operator
kubectl apply -f https://github.com/automationpi/kubejanitor/releases/latest/download/operator.yaml
```

## Uninstallation

### Helm Uninstall

```bash
# Uninstall operator
helm uninstall kubejanitor --namespace kubejanitor-system

# Remove CRDs (optional, this will delete all JanitorPolicy resources)
kubectl delete crd janitorpolicies.ops.k8s.io

# Remove namespace
kubectl delete namespace kubejanitor-system
```

### Manual Uninstall

```bash
# Delete all JanitorPolicy resources
kubectl delete janitorpolicy --all --all-namespaces

# Delete operator
kubectl delete -f https://github.com/automationpi/kubejanitor/releases/latest/download/operator.yaml

# Delete CRDs
kubectl delete -f https://github.com/automationpi/kubejanitor/releases/latest/download/crds.yaml
```

## Troubleshooting

### Common Issues

#### Operator Not Starting

1. Check resource requirements:
   ```bash
   kubectl describe pod -n kubejanitor-system -l app.kubernetes.io/name=kubejanitor-operator
   ```

2. Verify RBAC permissions:
   ```bash
   kubectl auth can-i '*' '*' --as=system:serviceaccount:kubejanitor-system:kubejanitor-operator
   ```

#### Policy Not Working

1. Check policy status:
   ```bash
   kubectl get janitorpolicy -A -o wide
   ```

2. View operator logs:
   ```bash
   kubectl logs -n kubejanitor-system deployment/kubejanitor-operator -f
   ```

3. Check events:
   ```bash
   kubectl get events -A --field-selector involvedObject.kind=JanitorPolicy
   ```

#### CRD Issues

1. Verify CRD installation:
   ```bash
   kubectl get crd janitorpolicies.ops.k8s.io -o yaml
   ```

2. Check CRD version compatibility:
   ```bash
   kubectl api-versions | grep ops.k8s.io
   ```

For more troubleshooting information, see the [Troubleshooting Guide](troubleshooting.md).

## Next Steps

- [Configuration Guide](configuration.md) - Learn about detailed configuration options
- [Policy Examples](examples/) - Explore real-world policy examples
- [Safety Guide](safety.md) - Understand safety mechanisms and best practices
- [API Reference](api-reference.md) - Complete API documentation