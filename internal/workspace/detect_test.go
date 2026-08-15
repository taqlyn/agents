package workspace

import (
	"os"
	"path/filepath"
	"testing"
)

func TestInspectAndroid(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, "app"), 0o755); err != nil {
		t.Fatal(err)
	}
	gradle := "android {\n    defaultConfig {\n        applicationId \"com.example.app\"\n    }\n}\n"
	if err := os.WriteFile(filepath.Join(dir, "app", "build.gradle"), []byte(gradle), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "settings.gradle"), []byte(""), 0o644); err != nil {
		t.Fatal(err)
	}
	r, err := Inspect(dir)
	if err != nil {
		t.Fatal(err)
	}
	if r.ApplicationID != "com.example.app" {
		t.Fatalf("got %q", r.ApplicationID)
	}
	if !contains(r.Platforms, Android) {
		t.Fatalf("%v", r.Platforms)
	}
}

func TestInspectFlutter(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "pubspec.yaml"), []byte("name: demo\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	r, err := Inspect(dir)
	if err != nil {
		t.Fatal(err)
	}
	if !contains(r.Platforms, Flutter) {
		t.Fatalf("%v", r.Platforms)
	}
}
