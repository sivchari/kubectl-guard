package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestConfig_AddContext(t *testing.T) {
	cfg := &Config{}

	cfg.AddContext("prod", nil)
	if len(cfg.GuardedContexts) != 1 {
		t.Fatalf("expected 1 context, got %d", len(cfg.GuardedContexts))
	}
	if cfg.GuardedContexts[0].Name != "prod" {
		t.Errorf("expected context name 'prod', got %s", cfg.GuardedContexts[0].Name)
	}

	// Adding same context should update, not duplicate
	cfg.AddContext("prod", []string{"default", "kube-system"})
	if len(cfg.GuardedContexts) != 1 {
		t.Fatalf("expected 1 context after update, got %d", len(cfg.GuardedContexts))
	}
	if len(cfg.GuardedContexts[0].Namespaces) != 2 {
		t.Errorf("expected 2 namespaces, got %d", len(cfg.GuardedContexts[0].Namespaces))
	}
}

func TestConfig_RemoveContext(t *testing.T) {
	cfg := &Config{
		GuardedContexts: []GuardedContext{
			{Name: "prod"},
			{Name: "staging"},
		},
	}

	removed := cfg.RemoveContext("prod")
	if !removed {
		t.Error("expected RemoveContext to return true")
	}
	if len(cfg.GuardedContexts) != 1 {
		t.Fatalf("expected 1 context after removal, got %d", len(cfg.GuardedContexts))
	}
	if cfg.GuardedContexts[0].Name != "staging" {
		t.Errorf("expected remaining context to be 'staging', got %s", cfg.GuardedContexts[0].Name)
	}

	// Remove non-existent context
	removed = cfg.RemoveContext("non-existent")
	if removed {
		t.Error("expected RemoveContext to return false for non-existent context")
	}
}

func TestConfig_IsGuarded(t *testing.T) {
	cfg := &Config{
		GuardedContexts: []GuardedContext{
			{Name: "prod"},
		},
	}

	if !cfg.IsGuarded("prod") {
		t.Error("expected 'prod' to be guarded")
	}
	if cfg.IsGuarded("dev") {
		t.Error("expected 'dev' to not be guarded")
	}
}

func TestConfig_IsNamespaceGuarded(t *testing.T) {
	cfg := &Config{
		GuardedContexts: []GuardedContext{
			{Name: "prod"}, // all namespaces
			{Name: "staging", Namespaces: []string{"production", "critical"}},
		},
	}

	// Context with all namespaces guarded
	if !cfg.IsNamespaceGuarded("prod", "default") {
		t.Error("expected 'prod/default' to be guarded")
	}
	if !cfg.IsNamespaceGuarded("prod", "kube-system") {
		t.Error("expected 'prod/kube-system' to be guarded")
	}

	// Context with specific namespaces
	if !cfg.IsNamespaceGuarded("staging", "production") {
		t.Error("expected 'staging/production' to be guarded")
	}
	if cfg.IsNamespaceGuarded("staging", "default") {
		t.Error("expected 'staging/default' to not be guarded")
	}

	// Non-guarded context
	if cfg.IsNamespaceGuarded("dev", "default") {
		t.Error("expected 'dev/default' to not be guarded")
	}
}

func TestConfig_SaveAndLoad(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "guard.yaml")

	cfg := &Config{
		GuardedContexts: []GuardedContext{
			{Name: "prod"},
			{Name: "staging", Namespaces: []string{"production"}},
		},
	}

	if err := cfg.SaveTo(path); err != nil {
		t.Fatalf("failed to save config: %v", err)
	}

	loaded, err := LoadFrom(path)
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}

	if len(loaded.GuardedContexts) != 2 {
		t.Fatalf("expected 2 contexts, got %d", len(loaded.GuardedContexts))
	}
	if loaded.GuardedContexts[0].Name != "prod" {
		t.Errorf("expected first context to be 'prod', got %s", loaded.GuardedContexts[0].Name)
	}
}

func TestLoadFrom_NonExistent(t *testing.T) {
	cfg, err := LoadFrom("/non/existent/path/guard.yaml")
	if err != nil {
		t.Fatalf("expected no error for non-existent file, got %v", err)
	}
	if len(cfg.GuardedContexts) != 0 {
		t.Errorf("expected empty config, got %d contexts", len(cfg.GuardedContexts))
	}
}

func TestLoadFrom_InvalidYAML(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "guard.yaml")

	if err := os.WriteFile(path, []byte("invalid: yaml: content:"), 0o644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	_, err := LoadFrom(path)
	if err == nil {
		t.Error("expected error for invalid YAML")
	}
}
