package models

import (
	"fmt"
	"time"
)

// DatastoreType represents the type of datastore backing a volume
type DatastoreType string

const (
	DatastorePostgres DatastoreType = "postgres"
	DatastoreMySQL    DatastoreType = "mysql"
	DatastoreRedis    DatastoreType = "redis"
	DatastoreMongoDB  DatastoreType = "mongodb"
	DatastoreNeo4j    DatastoreType = "neo4j"
	DatastoreGeneric  DatastoreType = "generic"
)

// Volume represents a Docker volume with datastore metadata
type Volume struct {
	Name          string        `yaml:"name" json:"name"`
	DatastoreType DatastoreType `yaml:"datastore_type" json:"datastore_type"`
	ContainerName string        `yaml:"container_name,omitempty" json:"container_name,omitempty"`
	MountPath     string        `yaml:"mount_path,omitempty" json:"mount_path,omitempty"`
	ImageName     string        `yaml:"image_name,omitempty" json:"image_name,omitempty"`
	SizeBytes     int64         `yaml:"size_bytes,omitempty" json:"size_bytes,omitempty"`
	SizeHuman     string        `yaml:"size_human,omitempty" json:"size_human,omitempty"`
}

// Snapshot represents a saved state of one or more volumes
type Snapshot struct {
	Name        string            `yaml:"name" json:"name"`
	Timestamp   time.Time         `yaml:"timestamp" json:"timestamp"`
	Volumes     []Volume          `yaml:"volumes" json:"volumes"`
	SizeBytes   int64             `yaml:"size_bytes" json:"size_bytes"`
	SizeHuman   string            `yaml:"size_human" json:"size_human"`
	Path        string            `yaml:"path" json:"path"`
	Checksum    string            `yaml:"checksum,omitempty" json:"checksum,omitempty"`
	Tags        []string          `yaml:"tags,omitempty" json:"tags,omitempty"`
	Description string            `yaml:"description,omitempty" json:"description,omitempty"`
	Metadata    map[string]string `yaml:"metadata,omitempty" json:"metadata,omitempty"`
	ParentName  string            `yaml:"parent_name,omitempty" json:"parent_name,omitempty"` // For incremental
	Incremental bool              `yaml:"incremental,omitempty" json:"incremental,omitempty"`
}

// SizeReport contains detailed size information
type SizeReport struct {
	TotalSize      int64         `json:"total_size"`
	TotalSizeHuman string        `json:"total_size_human"`
	ByDatastore    map[string]DatastoreSizeInfo `json:"by_datastore"`
	ByVolume       map[string]int64 `json:"by_volume"`
	SnapshotCount  int           `json:"snapshot_count"`
	SnapshotSize   int64         `json:"snapshot_size"`
}

// DatastoreSizeInfo holds size info for a datastore type
type DatastoreSizeInfo struct {
	Type      DatastoreType `json:"type"`
	TotalSize int64         `json:"total_size"`
	SizeHuman string        `json:"size_human"`
	Count     int           `json:"count"`
}

// Config represents dataclean configuration (auto-detected or from file)
type Config struct {
	// ComposeFile is the path to docker-compose.yaml (auto-detected if empty)
	ComposeFile string `yaml:"compose_file,omitempty"`

	// Volumes to explicitly include (if empty, auto-detect all)
	IncludeVolumes []string `yaml:"include_volumes,omitempty"`

	// Volumes to exclude from operations
	ExcludeVolumes []string `yaml:"exclude_volumes,omitempty"`

	// DatastoreHints maps volume names to datastore types (overrides auto-detection)
	DatastoreHints map[string]DatastoreType `yaml:"datastore_hints,omitempty"`

	// SnapshotDir is where snapshots are stored (default: .dataclean/)
	SnapshotDir string `yaml:"snapshot_dir,omitempty"`

	// BackupBeforeRestore creates automatic backup before restore/reset
	BackupBeforeRestore bool `yaml:"backup_before_restore,omitempty"`

	// DefaultTags are added to all snapshots
	DefaultTags []string `yaml:"default_tags,omitempty"`

	// RetentionDays is how long to keep snapshots (0 = forever)
	RetentionDays int `yaml:"retention_days,omitempty"`
}

// DefaultConfig returns a Config with sensible defaults
func DefaultConfig() *Config {
	return &Config{
		SnapshotDir:         ".dataclean",
		BackupBeforeRestore: true,
	}
}

// GetDatastoreInfo returns display information for a datastore type
func GetDatastoreInfo(dt DatastoreType) (name string, icon string) {
	switch dt {
	case DatastorePostgres:
		return "PostgreSQL", "🐘"
	case DatastoreMySQL:
		return "MySQL/MariaDB", "🐬"
	case DatastoreRedis:
		return "Redis", "🔴"
	case DatastoreMongoDB:
		return "MongoDB", "🍃"
	case DatastoreNeo4j:
		return "Neo4j", "🔵"
	default:
		return "Generic Volume", "📦"
	}
}

// AvailableDatastores returns all supported datastore types
func AvailableDatastores() []DatastoreType {
	return []DatastoreType{
		DatastorePostgres,
		DatastoreMySQL,
		DatastoreRedis,
		DatastoreMongoDB,
		DatastoreNeo4j,
		DatastoreGeneric,
	}
}

// FormatSize converts bytes to human-readable format
func FormatSize(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}
