# 🧼 KubeJanitor Operator

> *A Helm-deployable Kubernetes Operator to automate cleanup of stale, unused, or misconfigured resources — with built-in guardrails to ensure production safety.*

[![Go Version](https://img.shields.io/github/go-mod/go-version/automationpi/kubejanitor)](https://golang.org/)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)
[![Release](https://img.shields.io/github/v/release/automationpi/kubejanitor)](https://github.com/automationpi/kubejanitor/releases)
[![Docker Pulls](https://img.shields.io/docker/pulls/automationpi/kubejanitor)](https://hub.docker.com/r/automationpi/kubejanitor)

---

## 🎯 Purpose

KubeJanitor Operator helps platform engineers **identify and clean up unused or misconfigured resources** in Kubernetes. It supports **safe, policy-based deletion and remediation**, reducing operational burden and infrastructure waste without disrupting workloads.

---

## 📦 Supported Cleanup Tasks

| Task                                      | Description                                                   |
| ----------------------------------------- | ------------------------------------------------------------- |
| 🗑️ **Delete unattached PVCs**            | Removes PersistentVolumeClaims not used by any Pod            |
| 🔁 **Restart CrashLoopBackOff Pods**      | Detect and restart (or notify) Pods stuck in CrashLoopBackOff |
| 📄 **Clean up old Jobs**                  | Delete completed or failed Jobs older than a specified age    |
| 📦 **Delete unused ConfigMaps & Secrets** | Detect unreferenced ConfigMaps and Secrets                    |
| 🌐 **Remove orphaned Services**           | Delete Services without backing Pods or endpoints             |
| 📜 **Expired TLS certs**                  | Detect and optionally delete expired TLS Secrets              |
| 📁 **Stale volumes**                      | Identify unused PersistentVolumes (bound to deleted PVCs)     |
| 🧼 **Evict terminated Pods**              | Clean up Pods stuck in "Terminating" for too long             |
| 📈 **Resource hog alerts**                | Alert on pods that consistently exceed CPU/memory limits      |
| 📊 **Resource gap checker**               | Detect Pods/Deployments missing `requests` or `limits`        |
| ⏳ **Stale Helm releases**                 | Detect Helm releases in failed or orphaned state              |
| ⚠️ **Misconfigured RBAC**                 | Identify Roles/Bindings with non-existent subjects            |

---

## 🚀 Quick Start

### Installation via Helm

```bash
# Add the Helm repository
helm repo add kubejanitor https://automationpi.github.io/kubejanitor
helm repo update

# Install the operator
helm install kubejanitor kubejanitor/kubejanitor-operator \
  --namespace kubejanitor-system \
  --create-namespace \
  --set dryRun=true
```

### Create your first cleanup policy

```bash
kubectl apply -f - <<EOF
apiVersion: ops.k8s.io/v1alpha1
kind: JanitorPolicy
metadata:
  name: basic-cleanup
  namespace: kubejanitor-system
spec:
  dryRun: true
  schedule: "0 2 * * *"  # Daily at 2 AM
  cleanup:
    jobs:
      enabled: true
      olderThan: 24h
      statuses: ["Failed", "Complete"]
    pvc:
      enabled: true
      unusedFor: 48h
EOF
```

---

## 📚 Documentation

- [**Installation Guide**](docs/installation.md) - Detailed installation instructions
- [**Configuration**](docs/configuration.md) - Complete configuration reference
- [**Policy Examples**](docs/examples/) - Real-world cleanup policy examples
- [**Safety & Guardrails**](docs/safety.md) - Understanding safety mechanisms
- [**API Reference**](docs/api-reference.md) - CRD schema and field descriptions
- [**Troubleshooting**](docs/troubleshooting.md) - Common issues and solutions

---

## 🧾 JanitorPolicy CRD

KubeJanitor uses a single CRD to declare what to clean and how:

```yaml
apiVersion: ops.k8s.io/v1alpha1
kind: JanitorPolicy
metadata:
  name: comprehensive-cleanup
  namespace: platform
spec:
  dryRun: false
  schedule: "*/30 * * * *"

  cleanup:
    pvc:
      enabled: true
      unusedFor: 4h
    jobs:
      enabled: true
      olderThan: 12h
      statuses: ["Failed", "Complete"]
    configMaps:
      enabled: true
      olderThan: 72h
    secrets:
      enabled: true
      olderThan: 168h
    terminatingPods:
      enabled: true
      stuckFor: 15m

  protectedLabels:
    - app.kubernetes.io/managed-by=Helm
    - janitor.k8s.io/keep=true

  ignoreNamespaces:
    - kube-system
    - ingress-nginx
```

---

## 🛡️ Safety First

- **🔍 Dry Run Mode**: Default behavior - simulate all actions first
- **📝 Audit Logging**: Every action is logged with justification
- **🏷️ Label Protection**: Preserve Helm-managed and protected resources
- **⏰ TTL Thresholds**: Prevent premature cleanup
- **📋 Namespace Controls**: Whitelist/blacklist critical namespaces
- **💾 Backup Options**: Export deleted objects before removal

---

## 🧪 Development & Testing

### Local Development

```bash
# Clone the repository
git clone https://github.com/automationpi/kubejanitor.git
cd kubejanitor

# Start local test cluster
make test-cluster-up

# Run tests
make test

# Deploy locally
make deploy-local

# Cleanup
make test-cluster-down
```

### Running Tests

```bash
# Unit tests
make test-unit

# Integration tests
make test-integration

# End-to-end tests
make test-e2e
```

---

## 🤝 Contributing

We welcome contributions! Please see:

- [Contributing Guide](CONTRIBUTING.md)
- [Code of Conduct](CODE_OF_CONDUCT.md)
- [Development Setup](docs/development.md)

---

## 📄 License

This project is licensed under the Apache License 2.0 - see the [LICENSE](LICENSE) file for details.

---

## 🙏 Acknowledgments

- Inspired by the need for automated Kubernetes resource lifecycle management
- Built with [controller-runtime](https://github.com/kubernetes-sigs/controller-runtime)
- Thanks to all [contributors](https://github.com/automationpi/kubejanitor/graphs/contributors)

---

## 📞 Support

- 📖 [Documentation](docs/)
- 🐛 [Issue Tracker](https://github.com/automationpi/kubejanitor/issues)
- 💬 [Discussions](https://github.com/automationpi/kubejanitor/discussions)
- 📧 [Security Issues](mailto:security@automationpi.com)