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
	if v := os.Getenv("ARGOCD_TOKEN"); v != "" {
		c.ArgoCD.Token = v
	}
}

type Config struct {
	Server     ServerConfig     `json:"server" yaml:"server"`
	Database   DatabaseConfig   `json:"database" yaml:"database"`
	Auth       AuthConfig       `json:"auth" yaml:"auth"`
	CORS       CORSConfig       `json:"cors" yaml:"cors"`
	Kubernetes KubernetesConfig `json:"kubernetes" yaml:"kubernetes"`
	ArgoCD     ArgoCDConfig     `json:"argocd" yaml:"argocd"`
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
	BaseURL string `json:"base_url" yaml:"base_url"`
	Token   string `json:"token" yaml:"token"`
}
