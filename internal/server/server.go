package server

import (
	"context"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/taqlyn/agents/internal/api"
	"github.com/taqlyn/agents/internal/auth"
	"github.com/taqlyn/agents/internal/config"
	"github.com/taqlyn/agents/internal/tools"
)

const Version = "0.1.0"

func New(cfg config.Config, store *auth.Store) *mcp.Server {
	token := func() string { return store.Get().Token }
	s := mcp.NewServer(&mcp.Implementation{
		Name:    "taqlyn",
		Title:   "Taqlyn",
		Version: Version,
	}, nil)
	tools.Register(s, tools.Deps{
		Cfg:   cfg,
		Store: store,
		API:   api.New(cfg.APIURL, token),
	})
	return s
}

func Run(ctx context.Context, cfg config.Config) error {
	store, err := auth.NewStore(cfg.TokenPath)
	if err != nil {
		return err
	}
	if err := bootstrap(ctx, cfg, store); err != nil {
		log.Printf("bootstrap login skipped: %v", err)
	}
	srv := New(cfg, store)
	switch cfg.Transport {
	case "http":
		return serveHTTP(ctx, cfg, srv)
	default:
		log.SetOutput(os.Stderr)
		return srv.Run(ctx, &mcp.StdioTransport{})
	}
}

func serveHTTP(ctx context.Context, cfg config.Config, srv *mcp.Server) error {
	handler := mcp.NewStreamableHTTPHandler(func(*http.Request) *mcp.Server {
		return srv
	}, &mcp.StreamableHTTPOptions{Stateless: true})
	mux := http.NewServeMux()
	mux.Handle("/mcp", handler)
	mux.Handle("/", handler)
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"status":"ok","service":"taqlyn-mcp"}`))
	})
	hs := &http.Server{
		Addr:              cfg.HTTPAddr,
		Handler:           withCORS(mux),
		ReadHeaderTimeout: 10 * time.Second,
	}
	go func() {
		<-ctx.Done()
		_ = hs.Shutdown(context.Background())
	}()
	log.Printf("taqlyn-mcp http on %s (POST /mcp)", cfg.HTTPAddr)
	if err := hs.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return err
	}
	return nil
}

func withCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type, Accept, Mcp-Session-Id")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, DELETE, OPTIONS")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func bootstrap(ctx context.Context, cfg config.Config, store *auth.Store) error {
	cur := store.Get()
	if cur.Token != "" {
		return nil
	}
	if cfg.SessionTok != "" {
		return store.Put(auth.Snapshot{
			APIURL:      cfg.APIURL,
			Token:       cfg.SessionTok,
			Environment: cfg.Environment,
			Permission:  cfg.Permission,
		})
	}
	if cfg.Email == "" || cfg.Password == "" {
		return nil
	}
	cli := api.New(cfg.APIURL, func() string { return "" })
	out, err := cli.Login(ctx, cfg.Email, cfg.Password)
	if err != nil {
		return err
	}
	if mfa, _ := out["mfaRequired"].(bool); mfa {
		return nil
	}
	tok, _ := out["token"].(string)
	if strings.TrimSpace(tok) == "" {
		return nil
	}
	return store.Put(auth.Snapshot{
		APIURL:      cfg.APIURL,
		Token:       tok,
		UserID:      str(out, "userId"),
		OrgID:       str(out, "orgId"),
		Email:       str(out, "email"),
		Role:        str(out, "role"),
		Environment: cfg.Environment,
		Permission:  cfg.Permission,
	})
}

func str(m map[string]any, k string) string {
	v, _ := m[k].(string)
	return v
}
