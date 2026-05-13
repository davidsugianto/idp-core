#!/bin/bash
# Setup OpenCost in the Kind cluster for cost tracking integration tests
# Requires Prometheus to already be installed (run setup-prometheus.sh first)
set -e

CLUSTER_NAME="${CLUSTER_NAME:-idp-test}"
OPENCOST_NAMESPACE="${OPENCOST_NAMESPACE:-opencost}"
PROMETHEUS_NAMESPACE="${PROMETHEUS_NAMESPACE:-monitoring}"
OPENCOST_CHART_VERSION="${OPENCOST_CHART_VERSION:-1.43.0}"

echo "=== Setting up OpenCost for IDP integration tests ==="

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

# Check if Prometheus is installed
if ! kubectl get svc prometheus-server -n "${PROMETHEUS_NAMESPACE}" &>/dev/null; then
    echo "Error: Prometheus is not installed in namespace '${PROMETHEUS_NAMESPACE}'."
    echo "Run 'make dev-prometheus-setup' first."
    exit 1
fi

# Build the Prometheus internal URL
PROMETHEUS_URL="http://prometheus-server.${PROMETHEUS_NAMESPACE}.svc.cluster.local:80"

# Add OpenCost Helm repo
echo "Adding OpenCost Helm repo..."
helm repo add opencost https://opencost.github.io/opencost-helm-chart 2>/dev/null || true
helm repo update

# Create namespace
kubectl create namespace "${OPENCOST_NAMESPACE}" --dry-run=client -o yaml | kubectl apply -f -

# Install OpenCost
echo "Installing OpenCost (this may take a few minutes)..."
helm upgrade --install opencost opencost/opencost \
    --namespace "${OPENCOST_NAMESPACE}" \
    --version "${OPENCOST_CHART_VERSION}" \
    --set opencost.prometheus.external.url="${PROMETHEUS_URL}" \
    --set opencost.exporter.defaultClusterId="${CLUSTER_NAME}" \
    --set opencost.ui.enabled=false \
    --set opencost.exporter.resources.requests.cpu=50m \
    --set opencost.exporter.resources.requests.memory=128Mi \
    --set opencost.exporter.resources.limits.cpu=200m \
    --set opencost.exporter.resources.limits.memory=256Mi \
    --wait \
    --timeout 5m

echo ""
echo "Waiting for OpenCost pods to be ready..."
kubectl wait --for=condition=Ready pods -l app.kubernetes.io/name=opencost -n "${OPENCOST_NAMESPACE}" --timeout=120s || true

echo ""
echo "Checking OpenCost status..."
kubectl get pods -n "${OPENCOST_NAMESPACE}"

# Get the OpenCost service endpoint
OPENCOST_SVC=$(kubectl get svc -n "${OPENCOST_NAMESPACE}" opencost -o jsonpath='{.spec.clusterIP}' 2>/dev/null || echo "")

echo ""
echo "=== OpenCost Setup Complete! ==="
echo ""
echo "Namespace:       ${OPENCOST_NAMESPACE}"
echo "Service:         opencost"
echo "Internal URL:    http://opencost.${OPENCOST_NAMESPACE}.svc.cluster.local:9003"
echo "Prometheus URL:  ${PROMETHEUS_URL}"
echo ""
echo "To verify OpenCost is working:"
echo "  kubectl port-forward svc/opencost -n ${OPENCOST_NAMESPACE} 9003:9003 &"
echo "  curl -s http://localhost:9003/allocation?window=1h | jq ."
echo ""
echo "Update your config.yaml to enable OpenCost:"
echo "  finops:"
echo "    enabled: true"
echo "    opencost:"
echo "      base_url: \"http://opencost.${OPENCOST_NAMESPACE}.svc.cluster.local:9003\""
echo "      poll_interval: \"1h\""
echo "    prometheus:"
echo "      url: \"${PROMETHEUS_URL}\""
echo ""