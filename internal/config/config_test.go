// Package config_test tests the configuration handling
package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stackgen-cli/dataclean/internal/models"
)

func TestLoadConfig_Default(t *testing.T) {
	// Create temp directory without config file
	tmpDir, err := os.MkdirTemp("", "dataclean-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	origDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(origDir)

	cfg, err := Load("")
	if err != nil {
		t.Fatalf("Load() failed: %v", err)
	}

	// Should return default config
	if cfg.SnapshotDir != ".dataclean" {
		t.Errorf("SnapshotDir = %q, want %q", cfg.SnapshotDir, ".dataclean")
	}
	if !cfg.BackupBeforeRestore {
		t.Error("BackupBeforeRestore should be true by default")
	}
}

func TestLoadConfig_FromFile(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "dataclean-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create config file
	configContent := `
compose_file: custom-compose.yaml
include_volumes:
  - pgdata
  - redisdata
exclude_volumes:
  - tempdata
snapshot_dir: custom-snapshots
backup_before_restore: false
datastore_hints:
  myvolume: postgres
`
	configPath := filepath.Join(tmpDir, ".dataclean.yaml")
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	origDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(origDir)

	cfg, err := Load("")
	if err != nil {
		t.Fatalf("Load() failed: %v", err)
	}

	if cfg.ComposeFile != "custom-compose.yaml" {
		t.Errorf("ComposeFile = %q, want %q", cfg.ComposeFile, "custom-compose.yaml")
	}
	if len(cfg.IncludeVolumes) != 2 {
		t.Errorf("IncludeVolumes len = %d, want 2", len(cfg.IncludeVolumes))
	}
	if cfg.SnapshotDir != "custom-snapshots" {
		t.Errorf("SnapshotDir = %q, want %q", cfg.SnapshotDir, "custom-snapshots")
	}
	if cfg.BackupBeforeRestore {
		t.Error("BackupBeforeRestore should be false")
	}
	if cfg.DatastoreHints["myvolume"] != models.DatastorePostgres {
		t.Error("DatastoreHint not loaded correctly")
	}
}

func TestLoadConfig_YML(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "dataclean-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create .yml variant config file
	configContent := `snapshot_dir: yml-snapshots`
	configPath := filepath.Join(tmpDir, ".dataclean.yml")
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	origDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(origDir)

	cfg, err := Load("")
	if err != nil {
		t.Fatalf("Load() failed: %v", err)
	}

	if cfg.SnapshotDir != "yml-snapshots" {
		t.Errorf("SnapshotDir = %q, want %q", cfg.SnapshotDir, "yml-snapshots")
	}
}

func TestLoadConfig_ExplicitPath(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "dataclean-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create config at explicit path
	configContent := `snapshot_dir: explicit-path`
	configPath := filepath.Join(tmpDir, "my-config.yaml")
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	cfg, err := Load(configPath)
	if err != nil {
		t.Fatalf("Load() failed: %v", err)
	}

	if cfg.SnapshotDir != "explicit-path" {
		t.Errorf("SnapshotDir = %q, want %q", cfg.SnapshotDir, "explicit-path")
	}
}

func TestLoadConfig_InvalidYAML(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "dataclean-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create invalid YAML file
	configContent := `this is: not: valid: yaml: [[[[`
	configPath := filepath.Join(tmpDir, "invalid.yaml")
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	_, err = Load(configPath)
	if err == nil {
		t.Error("expected error for invalid YAML, got nil")
	}
}
