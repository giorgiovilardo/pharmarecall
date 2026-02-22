package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/giorgiovilardo/pharmarecall/internal/config"
)

func TestLoad(t *testing.T) {
	t.Run("loads all values from TOML file", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "config.toml")
		content := `
[server]
port = 9090

[db]
url = "postgres://user:pass@localhost:5432/testdb"

[session]
secret = "test-secret-key"

[lookahead]
days = 14
`
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}

		cfg, err := config.Load(path)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if cfg.Server.Port != 9090 {
			t.Errorf("server.port = %d, want 9090", cfg.Server.Port)
		}
		if cfg.DB.URL != "postgres://user:pass@localhost:5432/testdb" {
			t.Errorf("db.url = %q, want postgres URL", cfg.DB.URL)
		}
		if cfg.Session.Secret != "test-secret-key" {
			t.Errorf("session.secret = %q, want test-secret-key", cfg.Session.Secret)
		}
		if cfg.Lookahead.Days != 14 {
			t.Errorf("lookahead.days = %d, want 14", cfg.Lookahead.Days)
		}
	})

	t.Run("applies defaults for port and lookahead", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "config.toml")
		content := `
[db]
url = "postgres://localhost/pharmarecall"

[session]
secret = "s"
`
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}

		cfg, err := config.Load(path)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if cfg.Server.Port != 8080 {
			t.Errorf("server.port = %d, want default 8080", cfg.Server.Port)
		}
		if cfg.Lookahead.Days != 7 {
			t.Errorf("lookahead.days = %d, want default 7", cfg.Lookahead.Days)
		}
	})

	t.Run("returns error for missing file", func(t *testing.T) {
		_, err := config.Load("/nonexistent/config.toml")
		if err == nil {
			t.Fatal("expected error for missing file")
		}
	})
}
