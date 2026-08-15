package config

import (
	"fmt"
	"net"
	"net/url"
	"os"
	"strings"

	"github.com/taqlyn/agents/internal/auth"
)

const (
	// DefaultAPIURL is the live Taqlyn control plane. MCP never uses a local API.
	DefaultAPIURL    = "https://api.rutvik.qzz.io"
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
	if err := AssertLiveAPI(cfg.APIURL); err != nil {
		return Config{}, err
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

// AssertLiveAPI rejects loopback, Docker Desktop host, and private LAN targets.
func AssertLiveAPI(raw string) error {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return fmt.Errorf("TAQLYN_API_URL is empty")
	}
	u, err := url.Parse(raw)
	if err != nil || u.Host == "" {
		return fmt.Errorf("TAQLYN_API_URL is not a valid URL")
	}
	if u.Scheme != "https" {
		return fmt.Errorf("TAQLYN_API_URL must be https (live Taqlyn only, not a local API)")
	}
	host := strings.ToLower(u.Hostname())
	if isLocalHost(host) {
		return fmt.Errorf("TAQLYN_API_URL cannot be local (%s); MCP uses live Taqlyn data only", host)
	}
	return nil
}

func isLocalHost(host string) bool {
	switch host {
	case "localhost", "127.0.0.1", "::1", "0.0.0.0", "host.docker.internal", "host.containers.internal":
		return true
	}
	if strings.HasSuffix(host, ".localhost") || strings.HasSuffix(host, ".local") {
		return true
	}
	ip := net.ParseIP(host)
	if ip == nil {
		return false
	}
	return ip.IsLoopback() || ip.IsPrivate() || ip.IsUnspecified() || ip.IsLinkLocalUnicast()
}

func getenv(key, fallback string) string {
	v := strings.TrimSpace(os.Getenv(key))
	if v == "" {
		return fallback
	}
	return v
}
