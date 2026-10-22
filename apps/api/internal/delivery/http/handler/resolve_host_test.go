package handler

import "testing"

func TestResolveDBHost(t *testing.T) {
	t.Setenv("DBOKE_LOCALHOST_ALIAS", "")
	if got := resolveDBHost("localhost"); got != "localhost" {
		t.Errorf("no alias: got %q", got)
	}
	t.Setenv("DBOKE_LOCALHOST_ALIAS", "host.docker.internal")
	for _, h := range []string{"localhost", "LOCALHOST", "127.0.0.1", "::1"} {
		if got := resolveDBHost(h); got != "host.docker.internal" {
			t.Errorf("resolveDBHost(%q) = %q, want alias", h, got)
		}
	}
	if got := resolveDBHost("db.example.com"); got != "db.example.com" {
		t.Errorf("remote host rewritten: %q", got)
	}
}
