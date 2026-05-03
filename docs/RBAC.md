# Kubernetes RBAC Permissions

This document describes the Kubernetes RBAC permissions required for idp-core to function properly.

## ServiceAccount

The application runs under the `idp-core` ServiceAccount in the `idp-core` namespace.

```yaml
apiVersion: v1
kind: ServiceAccount
metadata:
  name: idp-core
  namespace: idp-core
```

## ClusterRole Permissions

idp-core requires cluster-wide permissions to manage environments across namespaces.

### Namespace Management

```yaml
- apiGroups: [""]
  resources: ["namespaces"]
  verbs: ["get", "list", "watch", "create", "update", "patch", "delete"]
```

**Why:** Each environment gets its own namespace. idp-core creates, updates, and deletes namespaces as part of environment lifecycle.

### Resource Quota Management

```yaml
- apiGroups: [""]
  resources: ["resourcequotas"]
  verbs: ["get", "list", "watch", "create", "update", "patch", "delete"]
```

**Why:** Optional resource quotas can be set per environment to limit CPU/memory usage.

### Network Policy Management

```yaml
- apiGroups: ["networking.k8s.io"]
  resources: ["networkpolicies"]
  verbs: ["get", "list", "watch", "create", "update", "patch", "delete"]
```

**Why:** Network policies are created to isolate environments from each other.

### Workload Observation

```yaml
- apiGroups: [""]
  resources: ["pods"]
  verbs: ["get", "list", "watch"]

- apiGroups: ["apps"]
  resources: ["deployments", "statefulsets", "daemonsets", "replicasets"]
  verbs: ["get", "list", "watch"]
```

**Why:** idp-core observes workloads to provide status information. It does not modify workloads directly.

### ArgoCD Application Management

```yaml
- apiGroups: ["argoproj.io"]
  resources: ["applications", "appprojects"]
  verbs: ["get", "list", "watch", "create", "update", "patch", "delete"]
```

**Why:** Each environment has an associated ArgoCD Application for GitOps. idp-core manages the full lifecycle.

## ClusterRoleBinding

The ClusterRole is bound cluster-wide:

```yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: idp-core
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: idp-core
subjects:
  - kind: ServiceAccount
    name: idp-core
    namespace: idp-core
```

## Security Considerations

### Why Cluster-Scope Permissions?

idp-core manages environments across the entire cluster, each in its own namespace. This requires:
- Creating new namespaces dynamically
- Managing resources in those namespaces
- Deleting namespaces when environments are removed

### Principle of Least Privilege

The permissions follow the principle of least privilege:
- **Read-only** for workloads (pods, deployments) - no modification needed
- **Full CRUD** only for resources idp-core owns (namespaces, resource quotas, network policies, ArgoCD applications)
- **No access** to secrets, configmaps, or other sensitive resources in other namespaces

### Recommended Additional Restrictions

For production deployments, consider:

1. **Namespace Labels**: Restrict operations to namespaces with specific labels
   ```yaml
   # Example: Only manage namespaces created by idp-core
   - apiGroups: [""]
     resources: ["namespaces"]
     verbs: ["update", "patch", "delete"]
     resourceNames: []  # Use admission webhook for dynamic restriction
   ```

2. **Admission Webhooks**: Use a validating webhook to ensure idp-core can only modify resources it owns (identified by labels)

3. **Audit Logging**: Enable audit logging for the idp-core ServiceAccount

## Installing RBAC

### Using kubectl

```bash
# Apply RBAC resources
kubectl apply -f deployments/kubernetes/base/rbac.yaml
```

### Using kustomize

```bash
kustomize build deployments/kubernetes/base/ | kubectl apply -f -
```

## Verifying Permissions

### Check ServiceAccount

```bash
kubectl get serviceaccount idp-core -n idp-core
```

### Check ClusterRole

```bash
kubectl get clusterrole idp-core
kubectl describe clusterrole idp-core
```

### Check ClusterRoleBinding

```bash
kubectl get clusterrolebinding idp-core
kubectl describe clusterrolebinding idp-core
```

### Test Permissions (can-i)

```bash
# Test if idp-core can create namespaces
kubectl auth can-i create namespaces --as=system:serviceaccount:idp-core:idp-core

# Test if idp-core can list pods in any namespace
kubectl auth can-i list pods --all-namespaces --as=system:serviceaccount:idp-core:idp-core

# Test if idp-core can manage ArgoCD applications
kubectl auth can-i create applications.argoproj.io --as=system:serviceaccount:idp-core:idp-core -n argocd
```

## Troubleshooting Permission Issues

### Permission Denied Errors

If idp-core logs show permission denied:

1. Verify the ServiceAccount exists:
   ```bash
   kubectl get sa idp-core -n idp-core
   ```

2. Verify the ClusterRoleBinding:
   ```bash
   kubectl get clusterrolebinding idp-core
   ```

3. Test specific permissions:
   ```bash
   kubectl auth can-i <verb> <resource> --as=system:serviceaccount:idp-core:idp-core
   ```

### ArgoCD Permission Issues

If ArgoCD operations fail:

1. Verify ArgoCD CRDs are installed:
   ```bash
   kubectl get crd | grep argoproj
   ```

2. Check if idp-core can access ArgoCD namespace:
   ```bash
   kubectl auth can-i get applications.argoproj.io --as=system:serviceaccount:idp-core:idp-core -n argocd
   ```

### Namespace Creation Failures

If namespace creation fails:

1. Check cluster-level namespace permissions:
   ```bash
   kubectl auth can-i create namespaces --as=system:serviceaccount:idp-core:idp-core
   ```

2. Check for LimitRanges or ResourceQuotas that might block creation
