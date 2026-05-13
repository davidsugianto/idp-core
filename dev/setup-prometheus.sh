#!/bin/bash
# Setup Prometheus in the Kind cluster for OpenCost integration
set -e

CLUSTER_NAME="${CLUSTER_NAME:-idp-test}"
PROMETHEUS_NAMESPACE="${PROMETHEUS_NAMESPACE:-monitoring}"
PROMETHEUS_CHART_VERSION="${PROMETHEUS_CHART_VERSION:-25.30.0}"

echo "=== Setting up Prometheus for IDP integration tests ==="

# Check prerequisites
if ! command -v kubectl &> /dev/null; then
    echo "Error: kubectl is not installed."
    exit 1
fi

if ! command -v helm &> /dev/null; then
    echo "Error: helm is not installed. Please install it first:"
    echo "  brew install helm"
    echo "  or visit https://helm.sh/docs/intro/install/"
    exit 1
fi

# Check if connected to a cluster
if ! kubectl cluster-info &>/dev/null; then
    echo "Error: Not connected to a Kubernetes cluster."
    echo "Run 'make dev-k8s-setup' first."
    exit 1
fi

# Add Prometheus community Helm repo
echo "Adding Prometheus community Helm repo..."
helm repo add prometheus-community https://prometheus-community.github.io/helm-charts 2>/dev/null || true
helm repo update

# Create namespace
kubectl create namespace "${PROMETHEUS_NAMESPACE}" --dry-run=client -o yaml | kubectl apply -f -

# Install Prometheus (server only, no alertmanager or pushgateway for dev)
echo "Installing Prometheus (this may take a few minutes)..."
helm upgrade --install prometheus prometheus-community/prometheus \
    --namespace "${PROMETHEUS_NAMESPACE}" \
    --version "${PROMETHEUS_CHART_VERSION}" \
    --set alertmanager.enabled=false \
    --set pushgateway.enabled=false \
    --set server.persistentVolume.enabled=false \
    --set server.resources.requests.cpu=100m \
    --set server.resources.requests.memory=256Mi \
    --set server.resources.limits.cpu=500m \
    --set server.resources.limits.memory=512Mi \
    --wait \
    --timeout 5m

echo ""
echo "Waiting for Prometheus pods to be ready..."
kubectl wait --for=condition=Ready pods -l app.kubernetes.io/name=prometheus -n "${PROMETHEUS_NAMESPACE}" --timeout=120s || true

echo ""
echo "Checking Prometheus status..."
kubectl get pods -n "${PROMETHEUS_NAMESPACE}"

# Get the Prometheus service endpoint
PROMETHEUS_SVC=$(kubectl get svc -n "${PROMETHEUS_NAMESPACE}" prometheus-server -o jsonpath='{.spec.clusterIP}' 2>/dev/null || echo "")
if [ -z "$PROMETHEUS_SVC" ]; then
    PROMETHEUS_SVC="prometheus-server.${PROMETHEUS_NAMESPACE}.svc.cluster.local"
fi

echo ""
echo "=== Prometheus Setup Complete! ==="
echo ""
echo "Namespace:       ${PROMETHEUS_NAMESPACE}"
echo "Service:         prometheus-server"
echo "Internal URL:    http://${PROMETHEUS_SVC}:80"
echo ""
echo "To port-forward Prometheus UI:"
echo "  kubectl port-forward svc/prometheus-server -n ${PROMETHEUS_NAMESPACE} 9090:80"
echo "  Then open http://localhost:9090"
echo ""
echo "To verify Prometheus is scraping metrics:"
echo "  kubectl port-forward svc/prometheus-server -n ${PROMETHEUS_NAMESPACE} 9090:80 &"
echo "  curl -s http://localhost:9090/api/v1/targets | jq '.data.activeTargets[] | select(.health==\"up\") | .labels'"
echo ""