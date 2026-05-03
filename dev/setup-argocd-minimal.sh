#!/bin/bash
# Ultra-minimal ArgoCD setup for CI/testing - just CRDs + basic controller
set -e

CLUSTER_NAME="${CLUSTER_NAME:-idp-test}"
ARGOCD_NAMESPACE="argocd"

echo "=== Ultra-Minimal ArgoCD Setup for Testing ==="

# Check prerequisites
if ! command -v kubectl &> /dev/null; then
    echo "Error: kubectl is not installed."
    exit 1
fi

# Check if connected to a cluster
if ! kubectl cluster-info &>/dev/null; then
    echo "Error: Not connected to a Kubernetes cluster."
    exit 1
fi

# Create namespace
kubectl create namespace "${ARGOCD_NAMESPACE}" --dry-run=client -o yaml | kubectl apply -f -

# Download and apply ArgoCD CRDs locally (avoid GitHub timeout)
echo "Downloading ArgoCD CRDs..."
ARGOCD_CRD_URL="https://raw.githubusercontent.com/argoproj/argo-cd/v2.11.0/manifests/crds/application-crd.yaml"

# Try to download CRDs with retry
MAX_RETRIES=3
RETRY=0
CRD_FILE="/tmp/argocd-crd.yaml"

while [ $RETRY -lt $MAX_RETRIES ]; do
    if curl -sSL --connect-timeout 30 --max-time 120 "$ARGOCD_CRD_URL" -o "$CRD_FILE"; then
        echo "✓ CRDs downloaded"
        break
    fi
    RETRY=$((RETRY + 1))
    echo "Retry $RETRY/$MAX_RETRIES..."
    sleep 5
done

if [ ! -f "$CRD_FILE" ] || [ ! -s "$CRD_FILE" ]; then
    echo "Failed to download CRDs, applying inline..."
    # Inline minimal CRD
    kubectl apply -f - <<EOF
apiVersion: apiextensions.k8s.io/v1
kind: CustomResourceDefinition
metadata:
  name: applications.argoproj.io
spec:
  group: argoproj.io
  names:
    kind: Application
    listKind: ApplicationList
    plural: applications
    singular: application
  scope: Namespaced
  versions:
  - name: v1alpha1
    served: true
    storage: true
    schema:
      openAPIV3Schema:
        type: object
        x-kubernetes-preserve-unknown-fields: true
    subresources:
      status: {}
EOF
else
    echo "Applying CRDs..."
    kubectl apply -f "$CRD_FILE"
fi

# Create minimal RBAC for testing
echo "Creating RBAC..."
kubectl apply -n "${ARGOCD_NAMESPACE}" -f - <<EOF
---
apiVersion: v1
kind: ServiceAccount
metadata:
  name: argocd-application-controller
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: argocd-application-controller
rules:
- apiGroups: ["*"]
  resources: ["*"]
  verbs: ["*"]
- nonResourceURLs: ["*"]
  verbs: ["*"]
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: argocd-application-controller
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: argocd-application-controller
subjects:
- kind: ServiceAccount
  name: argocd-application-controller
  namespace: argocd
EOF

# Deploy minimal ArgoCD components with resource limits
echo "Deploying minimal ArgoCD controller..."
kubectl apply -n "${ARGOCD_NAMESPACE}" -f - <<EOF
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: argocd-repo-server
spec:
  replicas: 1
  selector:
    matchLabels:
      app.kubernetes.io/name: argocd-repo-server
  template:
    metadata:
      labels:
        app.kubernetes.io/name: argocd-repo-server
    spec:
      containers:
      - name: argocd-repo-server
        image: quay.io/argoproj/argocd:v2.11.0
        imagePullPolicy: IfNotPresent
        command:
        - argocd-repo-server
        resources:
          requests:
            cpu: 100m
            memory: 256Mi
          limits:
            cpu: 500m
            memory: 512Mi
---
apiVersion: v1
kind: Service
metadata:
  name: argocd-repo-server
spec:
  selector:
    app.kubernetes.io/name: argocd-repo-server
  ports:
  - port: 8081
    targetPort: 8081
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: argocd-application-controller
spec:
  replicas: 1
  selector:
    matchLabels:
      app.kubernetes.io/name: argocd-application-controller
  template:
    metadata:
      labels:
        app.kubernetes.io/name: argocd-application-controller
    spec:
      serviceAccountName: argocd-application-controller
      containers:
      - name: argocd-application-controller
        image: quay.io/argoproj/argocd:v2.11.0
        imagePullPolicy: IfNotPresent
        command:
        - argocd-application-controller
        - --repo-server
        - argocd-repo-server:8081
        resources:
          requests:
            cpu: 100m
            memory: 256Mi
          limits:
            cpu: 500m
            memory: 512Mi
EOF

# Wait for deployments
echo "Waiting for ArgoCD deployments..."
echo "This may take a few minutes on first run (image pull)..."

for deployment in argocd-repo-server argocd-application-controller; do
    echo "Waiting for $deployment..."
    kubectl rollout status deployment/$deployment -n "${ARGOCD_NAMESPACE}" --timeout=300s || {
        echo "Warning: $deployment may not be ready"
        kubectl get pods -n "${ARGOCD_NAMESPACE}" -l app.kubernetes.io/name=$deployment
    }
done

echo ""
echo "Checking ArgoCD pods..."
kubectl get pods -n "${ARGOCD_NAMESPACE}"

# Verify CRDs are installed
echo ""
if kubectl get crd applications.argoproj.io >/dev/null 2>&1; then
    echo "✓ ArgoCD Application CRD is installed"
else
    echo "✗ ArgoCD Application CRD not found"
fi

echo ""
echo "=== Setup Complete ==="
