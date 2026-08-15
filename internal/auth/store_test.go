package auth

import "testing"

func TestAllowsEnv(t *testing.T) {
	s := Snapshot{Token: "t", Environment: EnvSandbox, Permission: PermWrite}
	if err := s.AllowsEnv(EnvSandbox); err != nil {
		t.Fatal(err)
	}
	if err := s.AllowsEnv(EnvProduction); err == nil {
		t.Fatal("expected production denied")
	}
	both := Snapshot{Token: "t", Environment: EnvBoth, Permission: PermRead}
	if err := both.AllowsEnv(EnvProduction); err != nil {
		t.Fatal(err)
	}
}

func TestAllowsWrite(t *testing.T) {
	read := Snapshot{Token: "t", Permission: PermRead}
	if err := read.AllowsWrite(); err != ErrReadOnly {
		t.Fatalf("got %v", err)
	}
	write := Snapshot{Token: "t", Permission: PermWrite}
	if err := write.AllowsWrite(); err != nil {
		t.Fatal(err)
	}
	empty := Snapshot{}
	if err := empty.AllowsWrite(); err != ErrNotAuthenticated {
		t.Fatalf("got %v", err)
	}
}

func TestParse(t *testing.T) {
	if ParseEnvironment("prod") != EnvProduction {
		t.Fatal(ParseEnvironment("prod"))
	}
	if ParsePermission("read-only") != PermRead {
		t.Fatal(ParsePermission("read-only"))
	}
}
