package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/stackgen-cli/dataclean/internal/config"
	"github.com/stackgen-cli/dataclean/internal/docker"
	"github.com/stackgen-cli/dataclean/internal/models"
)

var detectCmd = &cobra.Command{
	Use:   "detect",
	Short: "Detect compose file, services, and snapshot-capable volumes",
	Long: `Scan the current directory for Docker Compose configuration and list:
  • Compose file location
  • Services defined
  • Volumes attached to services
  • Datastore types (inferred or configured)
  • Which volumes are snapshot-capable

This is a read-only operation that helps you understand what dataclean will operate on.`,
	RunE: runDetect,
}

func init() {
	rootCmd.AddCommand(detectCmd)
}

func runDetect(cmd *cobra.Command, args []string) error {
	// Load configuration
	cfg, err := config.Load(cfgFile)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Create Docker client
	client, err := docker.NewClient()
	if err != nil {
		return fmt.Errorf("failed to create Docker client: %w", err)
	}
	defer client.Close()

	// Detect volumes
	volumes, err := client.DetectComposeVolumes(cfg)
	if err != nil {
		return fmt.Errorf("failed to detect volumes: %w", err)
	}

	// Print results
	if !quiet {
		printDetectionResults(cfg, volumes)
	}

	return nil
}

func printDetectionResults(cfg *models.Config, volumes []models.Volume) {
	cyan := color.New(color.FgCyan, color.Bold)
	green := color.New(color.FgGreen)
	yellow := color.New(color.FgYellow)
	white := color.New(color.FgWhite)

	// Header
	cyan.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	cyan.Println("  dataclean detection results")
	cyan.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println()

	// Compose file
	white.Print("Compose file: ")
	composeFile := cfg.ComposeFile
	if composeFile == "" {
		composeFile = findComposeFile()
	}
	if composeFile != "" {
		green.Println(composeFile)
	} else {
		yellow.Println("not found")
	}
	fmt.Println()

	// Volumes summary
	if len(volumes) == 0 {
		yellow.Println("No snapshot-capable volumes detected.")
		fmt.Println()
		white.Println("Tip: Ensure your compose file defines named volumes for data services.")
		return
	}

	cyan.Printf("Detected %d snapshot-capable volume(s):\n\n", len(volumes))

	// Group volumes by datastore type
	byType := make(map[models.DatastoreType][]models.Volume)
	for _, v := range volumes {
		byType[v.DatastoreType] = append(byType[v.DatastoreType], v)
	}

	// Print volumes grouped by type
	for _, dsType := range models.AvailableDatastores() {
		vols := byType[dsType]
		if len(vols) == 0 {
			continue
		}

		name, icon := models.GetDatastoreInfo(dsType)
		cyan.Printf("  %s %s\n", icon, name)

		for _, v := range vols {
			white.Printf("    • ")
			green.Printf("%s", v.Name)
			if v.MountPath != "" {
				white.Printf(" → %s", v.MountPath)
			}
			if v.ContainerName != "" {
				white.Printf(" (container: %s)", v.ContainerName)
			}
			fmt.Println()
		}
		fmt.Println()
	}

	// Show excluded volumes if any
	if len(cfg.ExcludeVolumes) > 0 {
		yellow.Println("Excluded volumes (from config):")
		for _, v := range cfg.ExcludeVolumes {
			white.Printf("    • %s\n", v)
		}
		fmt.Println()
	}

	// Show snapshot directory
	white.Print("Snapshot directory: ")
	snapshotDir := cfg.SnapshotDir
	if snapshotDir == "" {
		snapshotDir = ".dataclean"
	}
	if _, err := os.Stat(snapshotDir); os.IsNotExist(err) {
		yellow.Printf("%s (will be created)\n", snapshotDir)
	} else {
		green.Println(snapshotDir)
	}
	fmt.Println()

	// Available operations hint
	cyan.Println("Available operations:")
	white.Println("  dataclean snapshot [name]  Create snapshot of all detected volumes")
	white.Println("  dataclean restore <name>   Restore from a saved snapshot")
	white.Println("  dataclean reset            Wipe all volumes to empty state")
	white.Println("  dataclean list             Show available snapshots")
}

func findComposeFile() string {
	candidates := []string{"compose.yaml", "compose.yml", "docker-compose.yaml", "docker-compose.yml"}
	for _, name := range candidates {
		if _, err := os.Stat(name); err == nil {
			return name
		}
	}
	return ""
}

// Helper to check if slice contains string
func containsString(slice []string, s string) bool {
	for _, item := range slice {
		if strings.EqualFold(item, s) {
			return true
		}
	}
	return false
}
