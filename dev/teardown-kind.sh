#!/bin/bash
set -e

CLUSTER_NAME="${CLUSTER_NAME:-idp-test}"

echo "=== Tearing down Kind cluster ==="

if kind get clusters 2>/dev/null | grep -q "^${CLUSTER_NAME}$"; then
    echo "Deleting cluster '${CLUSTER_NAME}'..."
    kind delete cluster --name "${CLUSTER_NAME}"
    echo "Cluster deleted successfully!"
else
    echo "Cluster '${CLUSTER_NAME}' not found."
fi
