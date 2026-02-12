// Package models_test tests the data models
package models

import (
	"testing"
	"time"
)

func TestFormatSize(t *testing.T) {
	tests := []struct {
		bytes    int64
		expected string
	}{
		{0, "0 B"},
		{100, "100 B"},
		{1024, "1.0 KB"},
		{1536, "1.5 KB"},
		{1048576, "1.0 MB"},
		{1572864, "1.5 MB"},
		{1073741824, "1.0 GB"},
		{1610612736, "1.5 GB"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			got := FormatSize(tt.bytes)
			if got != tt.expected {
				t.Errorf("FormatSize(%d) = %q, want %q", tt.bytes, got, tt.expected)
			}
		})
	}
}

func TestGetDatastoreInfo(t *testing.T) {
	tests := []struct {
		dt           DatastoreType
		expectedName string
		hasIcon      bool
	}{
		{DatastorePostgres, "PostgreSQL", true},
		{DatastoreMySQL, "MySQL/MariaDB", true},
		{DatastoreRedis, "Redis", true},
		{DatastoreMongoDB, "MongoDB", true},
		{DatastoreNeo4j, "Neo4j", true},
		{DatastoreGeneric, "Generic Volume", true},
		{"unknown", "Generic Volume", true},
	}

	for _, tt := range tests {
		t.Run(string(tt.dt), func(t *testing.T) {
			name, icon := GetDatastoreInfo(tt.dt)
			if name != tt.expectedName {
				t.Errorf("GetDatastoreInfo(%s) name = %q, want %q", tt.dt, name, tt.expectedName)
			}
			if tt.hasIcon && icon == "" {
				t.Errorf("GetDatastoreInfo(%s) should return an icon", tt.dt)
			}
		})
	}
}

func TestAvailableDatastores(t *testing.T) {
	datastores := AvailableDatastores()

	// Should include all known types
	expected := []DatastoreType{
		DatastorePostgres,
		DatastoreMySQL,
		DatastoreRedis,
		DatastoreMongoDB,
		DatastoreNeo4j,
		DatastoreGeneric,
	}

	if len(datastores) != len(expected) {
		t.Errorf("expected %d datastores, got %d", len(expected), len(datastores))
	}

	for _, exp := range expected {
		found := false
		for _, ds := range datastores {
			if ds == exp {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("expected datastore %s not found", exp)
		}
	}
}

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	if cfg.SnapshotDir != ".dataclean" {
		t.Errorf("default SnapshotDir = %q, want %q", cfg.SnapshotDir, ".dataclean")
	}

	if !cfg.BackupBeforeRestore {
		t.Error("BackupBeforeRestore should be true by default")
	}
}

func TestSnapshotStruct(t *testing.T) {
	now := time.Now()
	s := Snapshot{
		Name:      "test-snapshot",
		Timestamp: now,
		Volumes: []Volume{
			{Name: "vol1", DatastoreType: DatastorePostgres},
			{Name: "vol2", DatastoreType: DatastoreRedis},
		},
		SizeBytes: 1048576,
		SizeHuman: "1.0 MB",
		Path:      "/snapshots/test-snapshot",
	}

	if s.Name != "test-snapshot" {
		t.Error("snapshot name mismatch")
	}
	if len(s.Volumes) != 2 {
		t.Error("expected 2 volumes")
	}
	if s.SizeBytes != 1048576 {
		t.Error("size mismatch")
	}
}

func TestVolumeStruct(t *testing.T) {
	v := Volume{
		Name:          "myproject_pgdata",
		DatastoreType: DatastorePostgres,
		ContainerName: "myproject-db-1",
		MountPath:     "/var/lib/postgresql/data",
		ImageName:     "postgres:15",
	}

	if v.Name != "myproject_pgdata" {
		t.Error("volume name mismatch")
	}
	if v.DatastoreType != DatastorePostgres {
		t.Error("datastore type mismatch")
	}
}

func TestConfigWithIncludes(t *testing.T) {
	cfg := &Config{
		ComposeFile:    "docker-compose.yaml",
		IncludeVolumes: []string{"pgdata", "redisdata"},
		ExcludeVolumes: []string{"tempdata"},
		DatastoreHints: map[string]DatastoreType{
			"custom-vol": DatastorePostgres,
		},
	}

	if len(cfg.IncludeVolumes) != 2 {
		t.Error("expected 2 include volumes")
	}
	if len(cfg.ExcludeVolumes) != 1 {
		t.Error("expected 1 exclude volume")
	}
	if cfg.DatastoreHints["custom-vol"] != DatastorePostgres {
		t.Error("datastore hint mismatch")
	}
}
