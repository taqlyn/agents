package tools

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/taqlyn/agents/internal/auth"
	"github.com/taqlyn/agents/internal/config"
)

func TestRequireWriteDenied(t *testing.T) {
	dir := t.TempDir()
	store, err := auth.NewStore(filepath.Join(dir, "mcp.json"))
	if err != nil {
		t.Fatal(err)
	}
	if err := store.Put(auth.Snapshot{
		Token:       "t",
		Environment: auth.EnvSandbox,
		Permission:  auth.PermRead,
	}); err != nil {
		t.Fatal(err)
	}
	d := Deps{Cfg: config.Config{APIURL: "http://127.0.0.1:8080"}, Store: store}
	if _, _, err := d.requireWrite("sandbox"); err != auth.ErrReadOnly {
		t.Fatalf("got %v", err)
	}
	if _, _, err := d.requireRead("production"); err == nil {
		t.Fatal("expected production denied")
	}
}

func TestStoreRoundTrip(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "mcp.json")
	store, err := auth.NewStore(p)
	if err != nil {
		t.Fatal(err)
	}
	if err := store.Put(auth.Snapshot{Token: "abc", Email: "a@b.c", Environment: auth.EnvBoth, Permission: auth.PermWrite}); err != nil {
		t.Fatal(err)
	}
	raw, err := os.ReadFile(p)
	if err != nil {
		t.Fatal(err)
	}
	var m map[string]any
	if err := json.Unmarshal(raw, &m); err != nil {
		t.Fatal(err)
	}
	if m["token"] != "abc" {
		t.Fatalf("%s", raw)
	}
	st2, err := auth.NewStore(p)
	if err != nil {
		t.Fatal(err)
	}
	if st2.Get().Email != "a@b.c" {
		t.Fatal(st2.Get())
	}
}

func TestWorkspaceRootEnv(t *testing.T) {
	t.Setenv("TAQLYN_WORKSPACE", "/tmp/app")
	got, err := workspaceRoot("")
	if err != nil {
		t.Fatal(err)
	}
	if got != "/tmp/app" {
		t.Fatal(got)
	}
	got, err = workspaceRoot("/explicit")
	if err != nil || got != "/explicit" {
		t.Fatalf("%s %v", got, err)
	}
}
