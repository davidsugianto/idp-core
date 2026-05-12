package config

import (
	"os"
	"strconv"

	"github.com/davidsugianto/go-pkgs/config"
)

func Load(path string) (*Config, error) {
	if path == "" {
		path = "configs/config.yaml"
	}
	cfg, err := config.Load[Config](path)
	if err != nil {
		return nil, err
	}

	// Override with environment variables if set
	cfg.applyEnvOverrides()

	return &cfg, nil
}

func (c *Config) applyEnvOverrides() {
	// Server
	if v := os.Getenv("SERVER_PORT"); v != "" {
		if port, err := strconv.Atoi(v); err == nil {
			c.Server.Port = port
		}
	}

	// Database
	if v := os.Getenv("DB_HOST"); v != "" {
		c.Database.Host = v
	}
	if v := os.Getenv("DB_PORT"); v != "" {
		if port, err := strconv.Atoi(v); err == nil {
			c.Database.Port = port
		}
	}
	if v := os.Getenv("DB_USER"); v != "" {
		c.Database.User = v
	}
	if v := os.Getenv("DB_PASSWORD"); v != "" {
		c.Database.Password = v
	}
	if v := os.Getenv("DB_NAME"); v != "" {
		c.Database.Name = v
	}

	// Auth
	if v := os.Getenv("JWT_SECRET"); v != "" {
		c.Auth.JWTSecret = v
	}
	if v := os.Getenv("JWT_EXPIRY"); v != "" {
		c.Auth.JWTExpiry = v
	}

	// Kubernetes
	if v := os.Getenv("K8S_IN_CLUSTER"); v != "" {
		if inCluster, err := strconv.ParseBool(v); err == nil {
			c.Kubernetes.InCluster = inCluster
		}
	}
	if v := os.Getenv("KUBECONFIG_PATH"); v != "" {
		c.Kubernetes.KubeconfigPath = v
	}

	// ArgoCD
	if v := os.Getenv("ARGOCD_BASE_URL"); v != "" {
		c.ArgoCD.BaseURL = v
	}
	if v := os.Getenv("ARGOCD_NAMESPACE"); v != "" {
		c.ArgoCD.Namespace = v
	}
	if v := os.Getenv("ARGOCD_TOKEN"); v != "" {
		c.ArgoCD.Token = v
	}

	// OIDC
	if v := os.Getenv("OIDC_ENABLED"); v != "" {
		if enabled, err := strconv.ParseBool(v); err == nil {
			c.OIDC.Enabled = enabled
		}
	}
	if v := os.Getenv("OIDC_ISSUER_URL"); v != "" {
		c.OIDC.IssuerURL = v
	}
	if v := os.Getenv("OIDC_CLIENT_ID"); v != "" {
		c.OIDC.ClientID = v
	}
	if v := os.Getenv("OIDC_CLIENT_SECRET"); v != "" {
		c.OIDC.ClientSecret = v
	}
	if v := os.Getenv("OIDC_REDIRECT_URL"); v != "" {
		c.OIDC.RedirectURL = v
	}
	if v := os.Getenv("OIDC_GROUPS_CLAIM"); v != "" {
		c.OIDC.GroupsClaim = v
	}
	if v := os.Getenv("OIDC_ADMIN_GROUP"); v != "" {
		c.OIDC.AdminGroup = v
	}
}

type Config struct {
	Server     ServerConfig     `json:"server" yaml:"server"`
	Database   DatabaseConfig   `json:"database" yaml:"database"`
	Auth       AuthConfig       `json:"auth" yaml:"auth"`
	CORS       CORSConfig       `json:"cors" yaml:"cors"`
	Kubernetes KubernetesConfig `json:"kubernetes" yaml:"kubernetes"`
	ArgoCD     ArgoCDConfig     `json:"argocd" yaml:"argocd"`
	OIDC       OIDCConfig       `json:"oidc" yaml:"oidc"`
}

type ServerConfig struct {
	Port int `json:"port" yaml:"port"`
}

type DatabaseConfig struct {
	Host     string `json:"host" yaml:"host"`
	Port     int    `json:"port" yaml:"port"`
	User     string `json:"user" yaml:"user"`
	Password string `json:"password" yaml:"password"`
	Name     string `json:"name" yaml:"name"`
	SSLMode  string `json:"sslmode" yaml:"sslmode"`
}

type AuthConfig struct {
	JWTSecret string `json:"jwt_secret" yaml:"jwt_secret"`
	JWTExpiry string `json:"jwt_expiry" yaml:"jwt_expiry"`
}

type CORSConfig struct {
	AllowedOrigins   []string `json:"allowed_origins" yaml:"allowed_origins"`
	AllowedMethods   []string `json:"allowed_methods" yaml:"allowed_methods"`
	AllowedHeaders   []string `json:"allowed_headers" yaml:"allowed_headers"`
	AllowCredentials bool     `json:"allow_credentials" yaml:"allow_credentials"`
}

type KubernetesConfig struct {
	InCluster      bool   `json:"in_cluster" yaml:"in_cluster"`
	KubeconfigPath string `json:"kubeconfig_path" yaml:"kubeconfig_path"`
}

type ArgoCDConfig struct {
	BaseURL   string `json:"base_url" yaml:"base_url"`
	Namespace string `json:"namespace" yaml:"namespace"`
	Token     string `json:"token" yaml:"token"`
}

type OIDCConfig struct {
	Enabled      bool     `json:"enabled" yaml:"enabled"`
	IssuerURL    string   `json:"issuer_url" yaml:"issuer_url"`
	ClientID     string   `json:"client_id" yaml:"client_id"`
	ClientSecret string   `json:"client_secret" yaml:"client_secret"`
	RedirectURL  string   `json:"redirect_url" yaml:"redirect_url"`
	Scopes       []string `json:"scopes" yaml:"scopes"`
	GroupsClaim  string   `json:"groups_claim" yaml:"groups_claim"`
	AdminGroup   string   `json:"admin_group" yaml:"admin_group"`
}
