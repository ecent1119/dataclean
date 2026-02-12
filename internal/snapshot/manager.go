package snapshot

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"gopkg.in/yaml.v3"

	"github.com/stackgen-cli/dataclean/internal/docker"
	"github.com/stackgen-cli/dataclean/internal/models"
)

// Manager handles snapshot operations
type Manager struct {
	client *docker.Client
	cfg    *models.Config
}

// CreateOptions controls snapshot creation
type CreateOptions struct {
	Tags        []string
	Description string
	Metadata    map[string]string
	Incremental bool    // Create incremental snapshot
	ParentName  string  // Name of parent snapshot for incremental
}

// NewManager creates a new snapshot manager
func NewManager(client *docker.Client, cfg *models.Config) *Manager {
	return &Manager{
		client: client,
		cfg:    cfg,
	}
}

// Create creates a new snapshot of the specified volumes
func (m *Manager) Create(name string, volumes []models.Volume) (*models.Snapshot, error) {
	return m.CreateWithOptions(name, volumes, CreateOptions{})
}

// CreateWithOptions creates a new snapshot with additional options
func (m *Manager) CreateWithOptions(name string, volumes []models.Volume, opts CreateOptions) (*models.Snapshot, error) {
	snapshotDir := filepath.Join(m.cfg.SnapshotDir, name)

	// Create snapshot directory
	if err := os.MkdirAll(snapshotDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create snapshot directory: %w", err)
	}

	// Stop containers for consistent snapshot
	m.client.StopContainers(volumes)
	defer m.client.StartContainers(volumes)

	// Export each volume
	var totalSize int64
	var snapshotVolumes []models.Volume

	for _, vol := range volumes {
		tarPath := filepath.Join(snapshotDir, fmt.Sprintf("%s.tar.gz", sanitizeName(vol.Name)))

		if err := m.client.ExportVolume(vol, tarPath); err != nil {
			return nil, fmt.Errorf("failed to export volume %s: %w", vol.Name, err)
		}

		// Get file size
		info, err := os.Stat(tarPath)
		if err == nil {
			totalSize += info.Size()
			vol.SizeBytes = info.Size()
			vol.SizeHuman = models.FormatSize(info.Size())
		}

		snapshotVolumes = append(snapshotVolumes, vol)
	}

	// Merge tags
	allTags := append(m.cfg.DefaultTags, opts.Tags...)

	// Create snapshot metadata
	snapshot := &models.Snapshot{
		Name:        name,
		Timestamp:   time.Now(),
		Volumes:     snapshotVolumes,
		SizeBytes:   totalSize,
		SizeHuman:   models.FormatSize(totalSize),
		Path:        snapshotDir,
		Tags:        allTags,
		Description: opts.Description,
		Metadata:    opts.Metadata,
		Incremental: opts.Incremental,
		ParentName:  opts.ParentName,
	}

	// Save metadata
	metadataPath := filepath.Join(snapshotDir, "metadata.yaml")
	metadataBytes, err := yaml.Marshal(snapshot)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal metadata: %w", err)
	}
	if err := os.WriteFile(metadataPath, metadataBytes, 0644); err != nil {
		return nil, fmt.Errorf("failed to write metadata: %w", err)
	}

	return snapshot, nil
}

// Restore restores volumes from a named snapshot
func (m *Manager) Restore(name string) error {
	snapshotDir := filepath.Join(m.cfg.SnapshotDir, name)

	// Load metadata
	snapshot, err := m.loadMetadata(snapshotDir)
	if err != nil {
		return err
	}

	// Create pre-restore backup if configured
	if m.cfg.BackupBeforeRestore {
		backupName := fmt.Sprintf("_pre-restore-%s", time.Now().Format("20060102-150405"))
		m.Create(backupName, snapshot.Volumes)
	}

	// Stop containers
	m.client.StopContainers(snapshot.Volumes)
	defer m.client.StartContainers(snapshot.Volumes)

	// Import each volume
	for _, vol := range snapshot.Volumes {
		tarPath := filepath.Join(snapshotDir, fmt.Sprintf("%s.tar.gz", sanitizeName(vol.Name)))

		if err := m.client.ImportVolume(tarPath, vol); err != nil {
			return fmt.Errorf("failed to import volume %s: %w", vol.Name, err)
		}
	}

	return nil
}

// Reset clears all data from the specified volumes
func (m *Manager) Reset(volumes []models.Volume) error {
	// Create pre-reset backup if configured
	if m.cfg.BackupBeforeRestore {
		backupName := fmt.Sprintf("_pre-reset-%s", time.Now().Format("20060102-150405"))
		m.Create(backupName, volumes)
	}

	// Stop containers
	m.client.StopContainers(volumes)
	defer m.client.StartContainers(volumes)

	// Clear each volume
	for _, vol := range volumes {
		if err := m.client.ClearVolume(vol); err != nil {
			return fmt.Errorf("failed to clear volume %s: %w", vol.Name, err)
		}
	}

	return nil
}

// List returns all available snapshots
func (m *Manager) List() ([]models.Snapshot, error) {
	var snapshots []models.Snapshot

	// Create snapshot dir if it doesn't exist
	if _, err := os.Stat(m.cfg.SnapshotDir); os.IsNotExist(err) {
		return snapshots, nil
	}

	entries, err := os.ReadDir(m.cfg.SnapshotDir)
	if err != nil {
		return nil, err
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		snapshotDir := filepath.Join(m.cfg.SnapshotDir, entry.Name())
		snapshot, err := m.loadMetadata(snapshotDir)
		if err != nil {
			continue // Skip invalid snapshots
		}

		snapshots = append(snapshots, *snapshot)
	}

	// Sort by timestamp (newest first)
	sort.Slice(snapshots, func(i, j int) bool {
		return snapshots[i].Timestamp.After(snapshots[j].Timestamp)
	})

	return snapshots, nil
}

// Get returns a specific snapshot by name
func (m *Manager) Get(name string) (*models.Snapshot, error) {
	snapshotDir := filepath.Join(m.cfg.SnapshotDir, name)
	return m.loadMetadata(snapshotDir)
}

// Delete removes a snapshot
func (m *Manager) Delete(name string) error {
	snapshotDir := filepath.Join(m.cfg.SnapshotDir, name)
	return os.RemoveAll(snapshotDir)
}

// loadMetadata loads snapshot metadata from a directory
func (m *Manager) loadMetadata(snapshotDir string) (*models.Snapshot, error) {
	metadataPath := filepath.Join(snapshotDir, "metadata.yaml")
	data, err := os.ReadFile(metadataPath)
	if err != nil {
		return nil, fmt.Errorf("snapshot metadata not found: %w", err)
	}

	var snapshot models.Snapshot
	if err := yaml.Unmarshal(data, &snapshot); err != nil {
		return nil, fmt.Errorf("invalid snapshot metadata: %w", err)
	}

	// Update path in case directory was moved
	snapshot.Path = snapshotDir

	return &snapshot, nil
}

// sanitizeName converts a volume name to a safe filename
func sanitizeName(name string) string {
	// Replace characters that might be problematic in filenames
	result := name
	for _, char := range []string{"/", "\\", ":", "*", "?", "\"", "<", ">", "|"} {
		result = strings.ReplaceAll(result, char, "_")
	}
	return result
}

// ListByTag returns snapshots that have a specific tag
func (m *Manager) ListByTag(tag string) ([]models.Snapshot, error) {
	all, err := m.List()
	if err != nil {
		return nil, err
	}

	var filtered []models.Snapshot
	for _, s := range all {
		for _, t := range s.Tags {
			if t == tag {
				filtered = append(filtered, s)
				break
			}
		}
	}

	return filtered, nil
}

// AddTag adds a tag to an existing snapshot
func (m *Manager) AddTag(name, tag string) error {
	snapshot, err := m.Get(name)
	if err != nil {
		return err
	}

	// Check if tag already exists
	for _, t := range snapshot.Tags {
		if t == tag {
			return nil // Already has tag
		}
	}

	snapshot.Tags = append(snapshot.Tags, tag)
	return m.saveMetadata(snapshot)
}

// RemoveTag removes a tag from an existing snapshot
func (m *Manager) RemoveTag(name, tag string) error {
	snapshot, err := m.Get(name)
	if err != nil {
		return err
	}

	var newTags []string
	for _, t := range snapshot.Tags {
		if t != tag {
			newTags = append(newTags, t)
		}
	}

	snapshot.Tags = newTags
	return m.saveMetadata(snapshot)
}

// UpdateDescription updates a snapshot's description
func (m *Manager) UpdateDescription(name, description string) error {
	snapshot, err := m.Get(name)
	if err != nil {
		return err
	}

	snapshot.Description = description
	return m.saveMetadata(snapshot)
}

// UpdateMetadata updates a snapshot's metadata
func (m *Manager) UpdateMetadata(name string, metadata map[string]string) error {
	snapshot, err := m.Get(name)
	if err != nil {
		return err
	}

	if snapshot.Metadata == nil {
		snapshot.Metadata = make(map[string]string)
	}
	for k, v := range metadata {
		snapshot.Metadata[k] = v
	}

	return m.saveMetadata(snapshot)
}

// saveMetadata saves snapshot metadata
func (m *Manager) saveMetadata(snapshot *models.Snapshot) error {
	metadataPath := filepath.Join(snapshot.Path, "metadata.yaml")
	metadataBytes, err := yaml.Marshal(snapshot)
	if err != nil {
		return err
	}
	return os.WriteFile(metadataPath, metadataBytes, 0644)
}

// GetSizeReport generates a size report for all volumes and snapshots
func (m *Manager) GetSizeReport(volumes []models.Volume) (*models.SizeReport, error) {
	report := &models.SizeReport{
		ByDatastore: make(map[string]models.DatastoreSizeInfo),
		ByVolume:    make(map[string]int64),
	}

	// Get volume sizes
	for _, vol := range volumes {
		size, err := m.client.GetVolumeSize(vol)
		if err != nil {
			continue
		}
		report.TotalSize += size
		report.ByVolume[vol.Name] = size

		// Aggregate by datastore type
		dsType := string(vol.DatastoreType)
		info := report.ByDatastore[dsType]
		info.Type = vol.DatastoreType
		info.TotalSize += size
		info.Count++
		report.ByDatastore[dsType] = info
	}

	// Update human-readable sizes
	report.TotalSizeHuman = models.FormatSize(report.TotalSize)
	for dsType, info := range report.ByDatastore {
		info.SizeHuman = models.FormatSize(info.TotalSize)
		report.ByDatastore[dsType] = info
	}

	// Get snapshot sizes
	snapshots, err := m.List()
	if err == nil {
		report.SnapshotCount = len(snapshots)
		for _, s := range snapshots {
			report.SnapshotSize += s.SizeBytes
		}
	}

	return report, nil
}

// CleanupOldSnapshots removes snapshots older than retention days
func (m *Manager) CleanupOldSnapshots() ([]string, error) {
	if m.cfg.RetentionDays <= 0 {
		return nil, nil
	}

	cutoff := time.Now().AddDate(0, 0, -m.cfg.RetentionDays)
	snapshots, err := m.List()
	if err != nil {
		return nil, err
	}

	var deleted []string
	for _, s := range snapshots {
		// Skip system backups (prefixed with _)
		if strings.HasPrefix(s.Name, "_") {
			continue
		}

		if s.Timestamp.Before(cutoff) {
			if err := m.Delete(s.Name); err == nil {
				deleted = append(deleted, s.Name)
			}
		}
	}

	return deleted, nil
}
