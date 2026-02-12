// Package snapshot_test tests the snapshot manager (unit tests only, no Docker)
package snapshot

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stackgen-cli/dataclean/internal/models"
)

func TestSanitizeName(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"simple", "simple"},
		{"with-dash", "with-dash"},
		{"with_underscore", "with_underscore"},
		{"with/slash", "with_slash"},
		{"project_pgdata", "project_pgdata"},
		{"my.app.volume", "my.app.volume"},
		{"with:colon", "with_colon"},
		{"with*star", "with_star"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := sanitizeName(tt.input)
			if got != tt.expected {
				t.Errorf("sanitizeName(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}

func TestLoadMetadata(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "dataclean-snapshot-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create snapshot directory structure
	snapshotDir := filepath.Join(tmpDir, "test-snapshot")
	if err := os.MkdirAll(snapshotDir, 0755); err != nil {
		t.Fatalf("failed to create snapshot dir: %v", err)
	}

	// Create metadata file
	metadata := `
name: test-snapshot
timestamp: 2024-01-15T10:30:00Z
volumes:
  - name: project_pgdata
    datastore_type: postgres
  - name: project_redis
    datastore_type: redis
size_bytes: 104857600
size_human: 100.0 MB
path: /snapshots/test-snapshot
`
	metadataPath := filepath.Join(snapshotDir, "metadata.yaml")
	if err := os.WriteFile(metadataPath, []byte(metadata), 0644); err != nil {
		t.Fatalf("failed to write metadata: %v", err)
	}

	// Create manager with mock config
	cfg := &models.Config{SnapshotDir: tmpDir}
	m := &Manager{cfg: cfg}

	snapshot, err := m.loadMetadata(snapshotDir)
	if err != nil {
		t.Fatalf("loadMetadata failed: %v", err)
	}

	if snapshot.Name != "test-snapshot" {
		t.Errorf("Name = %q, want %q", snapshot.Name, "test-snapshot")
	}
	if len(snapshot.Volumes) != 2 {
		t.Errorf("Volumes len = %d, want 2", len(snapshot.Volumes))
	}
	if snapshot.SizeBytes != 104857600 {
		t.Errorf("SizeBytes = %d, want 104857600", snapshot.SizeBytes)
	}
}

func TestLoadMetadata_NotFound(t *testing.T) {
	cfg := &models.Config{SnapshotDir: "/nonexistent"}
	m := &Manager{cfg: cfg}

	_, err := m.loadMetadata("/nonexistent/snapshot")
	if err == nil {
		t.Error("expected error for nonexistent metadata, got nil")
	}
}

func TestListSnapshots_Empty(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "dataclean-list-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	cfg := &models.Config{SnapshotDir: tmpDir}
	m := &Manager{cfg: cfg}

	snapshots, err := m.List()
	if err != nil {
		t.Fatalf("List() failed: %v", err)
	}

	if len(snapshots) != 0 {
		t.Errorf("expected 0 snapshots in empty dir, got %d", len(snapshots))
	}
}

func TestListSnapshots_Multiple(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "dataclean-list-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create multiple snapshot directories with metadata
	names := []string{"snapshot-1", "snapshot-2", "snapshot-3"}
	for _, name := range names {
		dir := filepath.Join(tmpDir, name)
		os.MkdirAll(dir, 0755)

		metadata := `
name: ` + name + `
timestamp: 2024-01-15T10:30:00Z
volumes: []
size_bytes: 0
`
		os.WriteFile(filepath.Join(dir, "metadata.yaml"), []byte(metadata), 0644)
	}

	cfg := &models.Config{SnapshotDir: tmpDir}
	m := &Manager{cfg: cfg}

	snapshots, err := m.List()
	if err != nil {
		t.Fatalf("List() failed: %v", err)
	}

	if len(snapshots) != 3 {
		t.Errorf("expected 3 snapshots, got %d", len(snapshots))
	}
}

func TestListSnapshots_SkipsInvalidDirs(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "dataclean-list-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create valid snapshot
	validDir := filepath.Join(tmpDir, "valid-snapshot")
	os.MkdirAll(validDir, 0755)
	os.WriteFile(filepath.Join(validDir, "metadata.yaml"), []byte(`name: valid`), 0644)

	// Create dir without metadata (should be skipped)
	invalidDir := filepath.Join(tmpDir, "invalid-snapshot")
	os.MkdirAll(invalidDir, 0755)

	// Create a file (not a dir, should be skipped)
	os.WriteFile(filepath.Join(tmpDir, "somefile.txt"), []byte("not a snapshot"), 0644)

	cfg := &models.Config{SnapshotDir: tmpDir}
	m := &Manager{cfg: cfg}

	snapshots, err := m.List()
	if err != nil {
		t.Fatalf("List() failed: %v", err)
	}

	if len(snapshots) != 1 {
		t.Errorf("expected 1 valid snapshot, got %d", len(snapshots))
	}
}

func TestDelete(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "dataclean-delete-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create snapshot to delete
	snapshotDir := filepath.Join(tmpDir, "to-delete")
	os.MkdirAll(snapshotDir, 0755)
	os.WriteFile(filepath.Join(snapshotDir, "metadata.yaml"), []byte(`name: to-delete`), 0644)
	os.WriteFile(filepath.Join(snapshotDir, "vol.tar.gz"), []byte("fake archive"), 0644)

	cfg := &models.Config{SnapshotDir: tmpDir}
	m := &Manager{cfg: cfg}

	err = m.Delete("to-delete")
	if err != nil {
		t.Fatalf("Delete() failed: %v", err)
	}

	// Verify deleted
	if _, err := os.Stat(snapshotDir); !os.IsNotExist(err) {
		t.Error("snapshot directory should be deleted")
	}
}

func TestDelete_NotFound(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "dataclean-delete-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	cfg := &models.Config{SnapshotDir: tmpDir}
	m := &Manager{cfg: cfg}

	// os.RemoveAll succeeds even if path doesn't exist, which is the behavior
	// The Delete function should work without error for non-existent snapshots
	err = m.Delete("nonexistent")
	if err != nil {
		t.Errorf("Delete() for non-existent snapshot should not error: %v", err)
	}
}

func TestSnapshotTimestampSorting(t *testing.T) {
	// Test that snapshots are properly sortable by timestamp
	snapshots := []models.Snapshot{
		{Name: "third", Timestamp: time.Date(2024, 3, 1, 0, 0, 0, 0, time.UTC)},
		{Name: "first", Timestamp: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)},
		{Name: "second", Timestamp: time.Date(2024, 2, 1, 0, 0, 0, 0, time.UTC)},
	}

	// Sort by timestamp (newest first)
	for i := 0; i < len(snapshots)-1; i++ {
		for j := i + 1; j < len(snapshots); j++ {
			if snapshots[j].Timestamp.After(snapshots[i].Timestamp) {
				snapshots[i], snapshots[j] = snapshots[j], snapshots[i]
			}
		}
	}

	if snapshots[0].Name != "third" {
		t.Error("expected 'third' to be first after sorting")
	}
	if snapshots[2].Name != "first" {
		t.Error("expected 'first' to be last after sorting")
	}
}
