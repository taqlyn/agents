package config

import (
	"fmt"
	"os"
	"strings"

	"github.com/taqlyn/agents/internal/auth"
)

const (
	DefaultAPIURL    = "http://127.0.0.1:8080"
	DefaultTransport = "stdio"
	DefaultHTTPAddr  = ":8787"
)

// Config is process-level MCP settings (env / flags). Session tokens live in auth.Store.
type Config struct {
	APIURL      string
	Transport   string
	HTTPAddr    string
	TokenPath   string
	Email       string
	Password    string
	SessionTok  string
	Environment auth.Environment
	Permission  auth.Permission
}

func Load() (Config, error) {
	cfg := Config{
		APIURL:      getenv("TAQLYN_API_URL", DefaultAPIURL),
		Transport:   strings.ToLower(getenv("TAQLYN_MCP_TRANSPORT", DefaultTransport)),
		HTTPAddr:    getenv("TAQLYN_MCP_ADDR", DefaultHTTPAddr),
		TokenPath:   os.Getenv("TAQLYN_TOKEN_PATH"),
		Email:       os.Getenv("TAQLYN_EMAIL"),
		Password:    os.Getenv("TAQLYN_PASSWORD"),
		SessionTok:  os.Getenv("TAQLYN_SESSION_TOKEN"),
		Environment: auth.ParseEnvironment(getenv("TAQLYN_ENV", string(auth.EnvSandbox))),
		Permission:  auth.ParsePermission(getenv("TAQLYN_PERMISSION", string(auth.PermWrite))),
	}
	if cfg.APIURL == "" {
		return Config{}, fmt.Errorf("TAQLYN_API_URL is empty")
	}
	switch cfg.Transport {
	case "stdio", "http":
	default:
		return Config{}, fmt.Errorf("TAQLYN_MCP_TRANSPORT must be stdio or http")
	}
	if cfg.Environment == "" {
		return Config{}, fmt.Errorf("TAQLYN_ENV must be sandbox, production, or both")
	}
	if cfg.Permission == "" {
		return Config{}, fmt.Errorf("TAQLYN_PERMISSION must be read or write")
	}
	return cfg, nil
}

func getenv(key, fallback string) string {
	v := strings.TrimSpace(os.Getenv(key))
	if v == "" {
		return fallback
	}
	return v
}
