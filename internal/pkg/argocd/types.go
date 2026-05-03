package argocd

// ApplicationSpec defines the ArgoCD Application specification
type ApplicationSpec struct {
	Name      string
	Namespace string // Target namespace
	RepoURL   string
	Revision  string
	Path      string
	ServerURL string
	Project   string // ArgoCD project (default: "default")
}
