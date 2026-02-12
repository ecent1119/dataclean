package docker

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/stackgen-cli/dataclean/internal/models"
)

// Client wraps Docker operations
type Client struct {
	ctx context.Context
}

// NewClient creates a new Docker client
func NewClient() (*Client, error) {
	// Verify docker is available
	_, err := exec.LookPath("docker")
	if err != nil {
		return nil, fmt.Errorf("docker not found in PATH: %w", err)
	}

	return &Client{
		ctx: context.Background(),
	}, nil
}

// Close releases any resources
func (c *Client) Close() error {
	return nil
}

// ComposeConfig represents the structure of docker-compose.yaml we care about
type ComposeConfig struct {
	Services map[string]ComposeService `yaml:"services"`
	Volumes  map[string]interface{}    `yaml:"volumes"`
}

// ComposeService represents a service in docker-compose.yaml
type ComposeService struct {
	Image         string   `yaml:"image"`
	ContainerName string   `yaml:"container_name"`
	Volumes       []string `yaml:"volumes"`
}

// DetectComposeVolumes finds volumes defined in docker-compose.yaml
func (c *Client) DetectComposeVolumes(cfg *models.Config) ([]models.Volume, error) {
	// Find compose file
	composeFile := cfg.ComposeFile
	if composeFile == "" {
		// Auto-detect
		candidates := []string{"compose.yaml", "compose.yml", "docker-compose.yaml", "docker-compose.yml"}
		for _, name := range candidates {
			if _, err := os.Stat(name); err == nil {
				composeFile = name
				break
			}
		}
	}

	if composeFile == "" {
		return nil, fmt.Errorf("no compose file found in current directory")
	}

	// Parse compose file
	data, err := os.ReadFile(composeFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read %s: %w", composeFile, err)
	}

	var compose ComposeConfig
	if err := yaml.Unmarshal(data, &compose); err != nil {
		return nil, fmt.Errorf("failed to parse %s: %w", composeFile, err)
	}

	// Extract volumes and infer datastore types
	var volumes []models.Volume
	projectName := c.getProjectName()

	for serviceName, service := range compose.Services {
		for _, volumeMount := range service.Volumes {
			// Parse volume mount (can be "volume:path" or "volume:path:mode")
			parts := strings.Split(volumeMount, ":")
			if len(parts) < 2 {
				continue
			}

			volumeName := parts[0]
			mountPath := parts[1]

			// Skip bind mounts (start with . or /)
			if strings.HasPrefix(volumeName, ".") || strings.HasPrefix(volumeName, "/") {
				continue
			}

			// Check if volume is in compose volumes section
			if compose.Volumes != nil {
				if _, exists := compose.Volumes[volumeName]; !exists {
					continue
				}
			}

			// Apply include/exclude filters
			if len(cfg.IncludeVolumes) > 0 && !contains(cfg.IncludeVolumes, volumeName) {
				continue
			}
			if contains(cfg.ExcludeVolumes, volumeName) {
				continue
			}

			// Determine datastore type
			datastoreType := c.inferDatastoreType(service.Image, mountPath, cfg.DatastoreHints[volumeName])

			// Docker Compose prefixes volume names with project name
			fullVolumeName := fmt.Sprintf("%s_%s", projectName, volumeName)

			volumes = append(volumes, models.Volume{
				Name:          fullVolumeName,
				DatastoreType: datastoreType,
				ContainerName: service.ContainerName,
				MountPath:     mountPath,
				ImageName:     service.Image,
			})

			// Store reference to short name for display
			_ = serviceName // used for context
		}
	}

	return volumes, nil
}

// inferDatastoreType determines the datastore type from image name or mount path
func (c *Client) inferDatastoreType(image, mountPath string, hint models.DatastoreType) models.DatastoreType {
	// Explicit hint takes precedence
	if hint != "" {
		return hint
	}

	imageLower := strings.ToLower(image)

	// Check image name patterns
	switch {
	case strings.Contains(imageLower, "postgres"):
		return models.DatastorePostgres
	case strings.Contains(imageLower, "mysql"), strings.Contains(imageLower, "mariadb"):
		return models.DatastoreMySQL
	case strings.Contains(imageLower, "redis"):
		return models.DatastoreRedis
	case strings.Contains(imageLower, "mongo"):
		return models.DatastoreMongoDB
	case strings.Contains(imageLower, "neo4j"):
		return models.DatastoreNeo4j
	}

	// Check mount path patterns
	switch {
	case strings.Contains(mountPath, "postgresql"), strings.Contains(mountPath, "pgdata"):
		return models.DatastorePostgres
	case strings.Contains(mountPath, "mysql"):
		return models.DatastoreMySQL
	case strings.Contains(mountPath, "redis"):
		return models.DatastoreRedis
	case strings.Contains(mountPath, "mongo"):
		return models.DatastoreMongoDB
	case strings.Contains(mountPath, "neo4j"):
		return models.DatastoreNeo4j
	}

	return models.DatastoreGeneric
}

// getProjectName returns the Docker Compose project name (directory name by default)
func (c *Client) getProjectName() string {
	cwd, err := os.Getwd()
	if err != nil {
		return "unknown"
	}
	return filepath.Base(cwd)
}

// StopContainers stops containers that use the specified volumes
func (c *Client) StopContainers(volumes []models.Volume) error {
	for _, v := range volumes {
		if v.ContainerName != "" {
			cmd := exec.CommandContext(c.ctx, "docker", "stop", v.ContainerName)
			cmd.Run() // Ignore errors - container might not be running
		}
	}
	return nil
}

// StartContainers starts containers that use the specified volumes
func (c *Client) StartContainers(volumes []models.Volume) error {
	for _, v := range volumes {
		if v.ContainerName != "" {
			cmd := exec.CommandContext(c.ctx, "docker", "start", v.ContainerName)
			cmd.Run() // Ignore errors - container might not exist
		}
	}
	return nil
}

// ExportVolume exports a volume's contents to a tar file
func (c *Client) ExportVolume(volume models.Volume, destPath string) error {
	// Create a temporary container to access the volume
	cmd := exec.CommandContext(c.ctx, "docker", "run", "--rm",
		"-v", fmt.Sprintf("%s:/data:ro", volume.Name),
		"-v", fmt.Sprintf("%s:/backup", filepath.Dir(destPath)),
		"alpine",
		"tar", "czf", fmt.Sprintf("/backup/%s", filepath.Base(destPath)), "-C", "/data", ".")

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("export failed: %s: %w", string(output), err)
	}

	return nil
}

// ImportVolume imports a tar file into a volume
func (c *Client) ImportVolume(srcPath string, volume models.Volume) error {
	// Clear existing data
	clearCmd := exec.CommandContext(c.ctx, "docker", "run", "--rm",
		"-v", fmt.Sprintf("%s:/data", volume.Name),
		"alpine",
		"sh", "-c", "rm -rf /data/* /data/.*")
	clearCmd.Run() // Ignore errors

	// Import from tar
	cmd := exec.CommandContext(c.ctx, "docker", "run", "--rm",
		"-v", fmt.Sprintf("%s:/data", volume.Name),
		"-v", fmt.Sprintf("%s:/backup:ro", filepath.Dir(srcPath)),
		"alpine",
		"tar", "xzf", fmt.Sprintf("/backup/%s", filepath.Base(srcPath)), "-C", "/data")

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("import failed: %s: %w", string(output), err)
	}

	return nil
}

// ClearVolume removes all data from a volume
func (c *Client) ClearVolume(volume models.Volume) error {
	cmd := exec.CommandContext(c.ctx, "docker", "run", "--rm",
		"-v", fmt.Sprintf("%s:/data", volume.Name),
		"alpine",
		"sh", "-c", "rm -rf /data/* /data/.*")

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("clear failed: %s: %w", string(output), err)
	}

	return nil
}

// GetVolumeSize returns the size of a volume in bytes
func (c *Client) GetVolumeSize(volume models.Volume) (int64, error) {
	cmd := exec.CommandContext(c.ctx, "docker", "run", "--rm",
		"-v", fmt.Sprintf("%s:/data:ro", volume.Name),
		"alpine",
		"du", "-sb", "/data")

	output, err := cmd.Output()
	if err != nil {
		return 0, err
	}

	var size int64
	fmt.Sscanf(string(output), "%d", &size)
	return size, nil
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
