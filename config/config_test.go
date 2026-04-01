package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadFromFile(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, ".kbn.yml")
	content := []byte(`vault: "/path/to/vault"
path: "my/project"
glob: "*.md"
fields:
  id: "ticket_id"
  title: "title"
  status: "status"
  priority: "priority"
  type: "type"
hidden_statuses:
  - "Closed"
  - "Archived"
`)
	os.WriteFile(cfgPath, content, 0644)

	cfg, err := LoadFromFile(cfgPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Vault != "/path/to/vault" {
		t.Errorf("vault = %q, want %q", cfg.Vault, "/path/to/vault")
	}
	if cfg.Path != "my/project" {
		t.Errorf("path = %q, want %q", cfg.Path, "my/project")
	}
	if cfg.Glob != "*.md" {
		t.Errorf("glob = %q, want %q", cfg.Glob, "*.md")
	}
	if cfg.Fields.Status != "status" {
		t.Errorf("fields.status = %q, want %q", cfg.Fields.Status, "status")
	}
	if cfg.Fields.ID != "ticket_id" {
		t.Errorf("fields.id = %q, want %q", cfg.Fields.ID, "ticket_id")
	}
	if len(cfg.HiddenStatuses) != 2 || cfg.HiddenStatuses[0] != "Closed" {
		t.Errorf("hidden_statuses = %v, want [Closed, Archived]", cfg.HiddenStatuses)
	}
}

func TestLoadFromFileDefaults(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, ".kbn.yml")
	content := []byte(`vault: "/path/to/vault"
path: "notes"
fields:
  status: "status"
`)
	os.WriteFile(cfgPath, content, 0644)

	cfg, err := LoadFromFile(cfgPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Glob != "*.md" {
		t.Errorf("default glob = %q, want %q", cfg.Glob, "*.md")
	}
	if cfg.HiddenStatuses != nil && len(cfg.HiddenStatuses) != 0 {
		t.Errorf("default hidden_statuses = %v, want empty", cfg.HiddenStatuses)
	}
}

func TestFullPath(t *testing.T) {
	cfg := &Config{Vault: "/vault", Path: "sub/dir"}
	got := cfg.FullPath()
	want := "/vault/sub/dir"
	if got != want {
		t.Errorf("FullPath() = %q, want %q", got, want)
	}
}
