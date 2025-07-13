#!/bin/bash
set -e

CLUSTER_NAME="kubejanitor-test"
KUBECONFIG_PATH="$HOME/.kube/config"

echo "🚀 Setting up KubeJanitor test cluster..."

# Check if kind is available
if ! command -v kind &> /dev/null; then
    echo "❌ kind not found. Please install kind first: https://kind.sigs.k8s.io/docs/user/quick-start/"
    exit 1
fi

# Check if kubectl is available
if ! command -v kubectl &> /dev/null; then
    echo "❌ kubectl not found. Please install kubectl first"
    exit 1
fi

# Check if cluster already exists
if kind get clusters | grep -q "^${CLUSTER_NAME}$"; then
    echo "⚠️  Cluster ${CLUSTER_NAME} already exists. Deleting..."
    kind delete cluster --name ${CLUSTER_NAME}
fi

# Create kind cluster configuration
cat > /tmp/kind-config.yaml <<EOF
kind: Cluster
apiVersion: kind.x-k8s.io/v1alpha4
name: ${CLUSTER_NAME}
nodes:
- role: control-plane
  kubeadmConfigPatches:
  - |
    kind: InitConfiguration
    nodeRegistration:
      kubeletExtraArgs:
        node-labels: "ingress-ready=true"
  extraPortMappings:
  - containerPort: 80
    hostPort: 8080
    protocol: TCP
  - containerPort: 443
    hostPort: 8443
    protocol: TCP
- role: worker
- role: worker
EOF

echo "📦 Creating kind cluster: ${CLUSTER_NAME}"
kind create cluster --config /tmp/kind-config.yaml --wait 300s

# Verify cluster is ready
echo "🔍 Verifying cluster is ready..."
kubectl cluster-info --context kind-${CLUSTER_NAME}

# Wait for all nodes to be ready
echo "⏳ Waiting for all nodes to be ready..."
kubectl wait --for=condition=Ready nodes --all --timeout=300s

# Create test namespaces
echo "🏗️  Creating test namespaces..."
kubectl create namespace test-workloads || true
kubectl create namespace test-system || true

# Label test namespaces
kubectl label namespace test-workloads environment=test || true
kubectl label namespace test-system environment=system || true

# Install some test resources for cleanup testing
echo "📋 Installing test resources..."

# Create test PVCs (some unused)
kubectl apply -f - <<EOF
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: unused-pvc-1
  namespace: test-workloads
  labels:
    test: "true"
spec:
  accessModes:
    - ReadWriteOnce
  resources:
    requests:
      storage: 1Gi
---
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: unused-pvc-2
  namespace: test-workloads
  labels:
    test: "true"
  annotations:
    created: "$(date -d '2 days ago' --iso-8601=seconds)"
spec:
  accessModes:
    - ReadWriteOnce
  resources:
    requests:
      storage: 1Gi
---
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: protected-pvc
  namespace: test-workloads
  labels:
    janitor.k8s.io/keep: "true"
spec:
  accessModes:
    - ReadWriteOnce
  resources:
    requests:
      storage: 1Gi
EOF

# Create test ConfigMaps and Secrets
kubectl apply -f - <<EOF
apiVersion: v1
kind: ConfigMap
metadata:
  name: unused-config-1
  namespace: test-workloads
  labels:
    test: "true"
data:
  key1: value1
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: unused-config-2
  namespace: test-workloads
  labels:
    test: "true"
data:
  key2: value2
---
apiVersion: v1
kind: Secret
metadata:
  name: unused-secret-1
  namespace: test-workloads
  labels:
    test: "true"
type: Opaque
data:
  password: dGVzdC1wYXNzd29yZA==
---
apiVersion: v1
kind: Secret
metadata:
  name: unused-secret-2
  namespace: test-workloads
  labels:
    test: "true"
type: Opaque
data:
  token: dGVzdC10b2tlbg==
EOF

# Create test Jobs (some failed, some completed)
kubectl apply -f - <<EOF
apiVersion: batch/v1
kind: Job
metadata:
  name: completed-job-1
  namespace: test-workloads
  labels:
    test: "true"
spec:
  template:
    spec:
      containers:
      - name: test
        image: busybox:1.36
        command: ["sh", "-c", "echo 'Job completed successfully' && sleep 5"]
      restartPolicy: Never
  backoffLimit: 0
---
apiVersion: batch/v1
kind: Job
metadata:
  name: failed-job-1
  namespace: test-workloads
  labels:
    test: "true"
spec:
  template:
    spec:
      containers:
      - name: test
        image: busybox:1.36
        command: ["sh", "-c", "echo 'Job will fail' && exit 1"]
      restartPolicy: Never
  backoffLimit: 0
EOF

# Create test Services (some without endpoints)
kubectl apply -f - <<EOF
apiVersion: v1
kind: Service
metadata:
  name: orphaned-service-1
  namespace: test-workloads
  labels:
    test: "true"
spec:
  selector:
    app: non-existent-app
  ports:
  - port: 80
    targetPort: 8080
---
apiVersion: v1
kind: Service
metadata:
  name: orphaned-service-2
  namespace: test-workloads
  labels:
    test: "true"
spec:
  selector:
    app: another-non-existent-app
  ports:
  - port: 443
    targetPort: 8443
EOF

# Wait a bit for jobs to complete/fail
echo "⏳ Waiting for test jobs to complete..."
sleep 30

echo "✅ Test cluster setup completed!"
echo ""
echo "📋 Cluster Information:"
echo "  Name: ${CLUSTER_NAME}"
echo "  Context: kind-${CLUSTER_NAME}"
echo "  Nodes: $(kubectl get nodes --no-headers | wc -l)"
echo ""
echo "📦 Test Resources Created:"
echo "  - 3 PVCs (2 unused, 1 protected)"
echo "  - 2 ConfigMaps (unused)"
echo "  - 2 Secrets (unused)"
echo "  - 2 Jobs (1 completed, 1 failed)"
echo "  - 2 Services (orphaned)"
echo ""
echo "🔧 Useful Commands:"
echo "  kubectl config use-context kind-${CLUSTER_NAME}"
echo "  kubectl get all -A"
echo "  kubectl get pvc -A"
echo "  make deploy-local"
echo "  make helm-install"
echo ""
echo "🧹 To clean up:"
echo "  make test-cluster-down"