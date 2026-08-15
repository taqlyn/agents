package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestClientLoginAndMe(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /v1/auth/login", func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{
			"token": "sess_1", "userId": "usr_1", "orgId": "org_1",
			"email": "a@b.c", "role": "owner", "emailVerified": true,
		})
	})
	mux.HandleFunc("GET /v1/auth/me", func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer sess_1" {
			w.WriteHeader(401)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "auth.unauthorized", "message": "nope"})
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"email": "a@b.c", "orgId": "org_1"})
	})
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)

	tok := ""
	c := New(srv.URL, func() string { return tok })
	out, err := c.Login(context.Background(), "a@b.c", "pw")
	if err != nil {
		t.Fatal(err)
	}
	tok, _ = out["token"].(string)
	me, err := c.Me(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if me["email"] != "a@b.c" {
		t.Fatalf("%v", me)
	}
}
