package config

import "testing"

func TestAssertLiveAPI(t *testing.T) {
	ok := []string{
		"https://api.rutvik.qzz.io",
		"https://api.example.com/v1",
	}
	for _, u := range ok {
		if err := AssertLiveAPI(u); err != nil {
			t.Fatalf("%s: %v", u, err)
		}
	}
	bad := []string{
		"http://127.0.0.1:8080",
		"https://localhost",
		"http://host.docker.internal:8080",
		"https://10.0.0.5",
		"https://192.168.1.10",
		"http://api.rutvik.qzz.io",
		"",
	}
	for _, u := range bad {
		if err := AssertLiveAPI(u); err == nil {
			t.Fatalf("expected reject %s", u)
		}
	}
}

func TestLoadDefaultIsLive(t *testing.T) {
	t.Setenv("TAQLYN_API_URL", "")
	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.APIURL != DefaultAPIURL {
		t.Fatal(cfg.APIURL)
	}
}

func TestLoadRejectsLocal(t *testing.T) {
	t.Setenv("TAQLYN_API_URL", "http://127.0.0.1:8080")
	if _, err := Load(); err == nil {
		t.Fatal("expected local API rejected")
	}
}
