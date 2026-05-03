#!/bin/bash
set -e

CLUSTER_NAME="${CLUSTER_NAME:-idp-test}"
ARGOCD_VERSION="${ARGOCD_VERSION:-v2.11.0}"
ARGOCD_NAMESPACE="argocd"
TIMEOUT="${TIMEOUT:-600}"

echo "=== Setting up Kind cluster for IDP integration tests ==="

# Check if kind is installed
if ! command -v kind &> /dev/null; then
    echo "Error: kind is not installed. Please install it first:"
    echo "  go install sigs.k8s.io/kind@latest"
    exit 1
fi

# Check if kubectl is installed
if ! command -v kubectl &> /dev/null; then
    echo "Error: kubectl is not installed."
    exit 1
fi

# Check if cluster already exists
if kind get clusters 2>/dev/null | grep -q "^${CLUSTER_NAME}$"; then
    echo "Cluster '${CLUSTER_NAME}' already exists."
    read -p "Do you want to delete and recreate it? (y/N) " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        echo "Deleting existing cluster..."
        kind delete cluster --name "${CLUSTER_NAME}"
    else
        echo "Using existing cluster."
    fi
fi

# Create cluster if it doesn't exist
if ! kind get clusters 2>/dev/null | grep -q "^${CLUSTER_NAME}$"; then
    echo "Creating Kind cluster '${CLUSTER_NAME}'..."
    kind create cluster --name "${CLUSTER_NAME}" --config dev/kind-config.yaml
    echo "Cluster created successfully!"
fi

# Set kubectl context
echo "Setting kubectl context..."
kubectl config use-context "kind-${CLUSTER_NAME}"

# Wait for cluster to be ready
echo "Waiting for cluster to be ready..."
kubectl wait --for=condition=Ready nodes --all --timeout=120s

# Install ArgoCD
echo "=== Installing ArgoCD ==="

# Create ArgoCD namespace
kubectl create namespace "${ARGOCD_NAMESPACE}" --dry-run=client -o yaml | kubectl apply -f -

# Install ArgoCD
echo "Installing ArgoCD ${ARGOCD_VERSION}..."
kubectl apply -n "${ARGOCD_NAMESPACE}" -f "https://raw.githubusercontent.com/argoproj/argo-cd/${ARGOCD_VERSION}/manifests/install.yaml"

# Wait for ArgoCD CRDs to be established
echo "Waiting for ArgoCD CRDs..."
sleep 10

# Wait for each ArgoCD component with longer timeout
echo "Waiting for ArgoCD components to be ready (this may take several minutes)..."

# Function to wait for deployment with retries
wait_for_deployment() {
    local deployment=$1
    local namespace=$2
    local max_attempts=30
    local attempt=1

    echo "Waiting for ${deployment}..."
    while [ $attempt -le $max_attempts ]; do
        if kubectl wait --for=condition=Available --timeout=60s deployment/${deployment} -n ${namespace} 2>/dev/null; then
            echo "✓ ${deployment} is ready"
            return 0
        fi
        echo "  Attempt ${attempt}/${max_attempts} - ${deployment} not ready yet..."
        sleep 5
        attempt=$((attempt + 1))
    done
    echo "✗ ${deployment} failed to become ready"
    return 1
}

# Wait for each component
wait_for_deployment "argocd-server" "${ARGOCD_NAMESPACE}" || true
wait_for_deployment "argocd-repo-server" "${ARGOCD_NAMESPACE}" || true
wait_for_deployment "argocd-application-controller" "${ARGOCD_NAMESPACE}" || true
wait_for_deployment "argocd-dex-server" "${ARGOCD_NAMESPACE}" || true
wait_for_deployment "argocd-redis" "${ARGOCD_NAMESPACE}" || true

# Check overall status
echo ""
echo "Checking ArgoCD status..."
kubectl get pods -n "${ARGOCD_NAMESPACE}"

# Verify ArgoCD is functional
echo ""
echo "Verifying ArgoCD API..."
max_retries=10
retry=0
while [ $retry -lt $max_retries ]; do
    if kubectl get applications -n "${ARGOCD_NAMESPACE}" >/dev/null 2>&1; then
        echo "✓ ArgoCD API is ready"
        break
    fi
    echo "Waiting for ArgoCD API..."
    sleep 5
    retry=$((retry + 1))
done

echo ""
echo "=== Setup Complete! ==="
echo ""
echo "Cluster: ${CLUSTER_NAME}"
echo "ArgoCD namespace: ${ARGOCD_NAMESPACE}"
echo ""
echo "To run integration tests:"
echo "  make test-all-integration"
echo ""
echo "To access ArgoCD UI (optional):"
echo "  kubectl port-forward svc/argocd-server -n argocd 8080:443"
echo "  Initial password: $(kubectl -n argocd get secret argocd-initial-admin-secret -o jsonpath='{.data.password}' 2>/dev/null | base64 -d 2>/dev/null || echo 'check secret manually')"
echo ""
