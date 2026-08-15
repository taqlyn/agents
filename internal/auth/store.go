package auth

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

type Environment string

const (
	EnvSandbox    Environment = "sandbox"
	EnvProduction Environment = "production"
	EnvBoth       Environment = "both"
)

type Permission string

const (
	PermRead  Permission = "read"
	PermWrite Permission = "write"
)

// Snapshot is the persisted MCP login (never stores the user password).
type Snapshot struct {
	APIURL      string      `json:"apiUrl"`
	Token       string      `json:"token"`
	UserID      string      `json:"userId,omitempty"`
	OrgID       string      `json:"orgId,omitempty"`
	Email       string      `json:"email,omitempty"`
	Role        string      `json:"role,omitempty"`
	Environment Environment `json:"environment"`
	Permission  Permission  `json:"permission"`
	UpdatedAt   time.Time   `json:"updatedAt"`
}

func ParseEnvironment(v string) Environment {
	switch strings.ToLower(strings.TrimSpace(v)) {
	case "sandbox", "test":
		return EnvSandbox
	case "production", "prod", "live":
		return EnvProduction
	case "both", "all":
		return EnvBoth
	default:
		return ""
	}
}

func ParsePermission(v string) Permission {
	switch strings.ToLower(strings.TrimSpace(v)) {
	case "read", "readonly", "read-only":
		return PermRead
	case "write", "readwrite", "read-write":
		return PermWrite
	default:
		return ""
	}
}

func (s Snapshot) AllowsEnv(env Environment) error {
	if s.Token == "" {
		return ErrNotAuthenticated
	}
	want := canonicalEnv(env)
	if want == "" {
		return fmt.Errorf("environment must be sandbox or production")
	}
	switch s.Environment {
	case EnvBoth:
		return nil
	case want:
		return nil
	default:
		return fmt.Errorf("scope denied: token allows %s, requested %s", s.Environment, want)
	}
}

func (s Snapshot) AllowsWrite() error {
	if s.Token == "" {
		return ErrNotAuthenticated
	}
	if s.Permission != PermWrite {
		return ErrReadOnly
	}
	return nil
}

func canonicalEnv(env Environment) Environment {
	switch env {
	case EnvSandbox, EnvProduction:
		return env
	case "":
		return EnvSandbox
	default:
		return ParseEnvironment(string(env))
	}
}

var (
	ErrNotAuthenticated = errors.New("not authenticated: call auth_login")
	ErrReadOnly         = errors.New("scope denied: token is read-only")
)

// Store persists the session snapshot at 0600.
type Store struct {
	path string
	mu   sync.Mutex
	cur  Snapshot
}

func DefaultPath() (string, error) {
	if p := strings.TrimSpace(os.Getenv("TAQLYN_TOKEN_PATH")); p != "" {
		return p, nil
	}
	dir, err := os.UserConfigDir()
	if err != nil {
		home, homeErr := os.UserHomeDir()
		if homeErr != nil {
			return "", err
		}
		dir = filepath.Join(home, ".config")
	}
	return filepath.Join(dir, "taqlyn", "mcp.json"), nil
}

func NewStore(path string) (*Store, error) {
	if path == "" {
		var err error
		path, err = DefaultPath()
		if err != nil {
			return nil, err
		}
	}
	s := &Store{path: path}
	_ = s.load()
	return s, nil
}

func (s *Store) Path() string { return s.path }

func (s *Store) Get() Snapshot {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.cur
}

func (s *Store) Put(next Snapshot) error {
	if next.Environment == "" {
		next.Environment = EnvSandbox
	}
	if next.Permission == "" {
		next.Permission = PermWrite
	}
	next.UpdatedAt = time.Now().UTC()
	s.mu.Lock()
	defer s.mu.Unlock()
	s.cur = next
	return s.flushLocked()
}

func (s *Store) Clear() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.cur = Snapshot{}
	if err := os.Remove(s.path); err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}

func (s *Store) load() error {
	b, err := os.ReadFile(s.path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	var snap Snapshot
	if err := json.Unmarshal(b, &snap); err != nil {
		return err
	}
	s.mu.Lock()
	s.cur = snap
	s.mu.Unlock()
	return nil
}

func (s *Store) flushLocked() error {
	if err := os.MkdirAll(filepath.Dir(s.path), 0o700); err != nil {
		return err
	}
	b, err := json.MarshalIndent(s.cur, "", "  ")
	if err != nil {
		return err
	}
	tmp := s.path + ".tmp"
	if err := os.WriteFile(tmp, b, 0o600); err != nil {
		return err
	}
	return os.Rename(tmp, s.path)
}
