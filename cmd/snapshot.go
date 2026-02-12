package cmd

import (
	"fmt"
	"time"

	"github.com/fatih/color"
	"github.com/spf13/cobra"

	"github.com/stackgen-cli/dataclean/internal/config"
	"github.com/stackgen-cli/dataclean/internal/docker"
	"github.com/stackgen-cli/dataclean/internal/snapshot"
)

var (
	snapshotTags        []string
	snapshotDescription string
	snapshotMetadata    map[string]string
	snapshotInclude     []string
	snapshotExclude     []string
)

var snapshotCmd = &cobra.Command{
	Use:   "snapshot [name]",
	Short: "Create a named snapshot of current data state",
	Long: `Create a snapshot of all detected Docker Compose data volumes.

If no name is provided, a timestamp-based name will be generated.

Examples:
  dataclean snapshot                    # auto-named: snapshot-2024-01-15-143052
  dataclean snapshot before-migration   # named: before-migration
  dataclean snapshot --dry-run          # preview what would be snapshotted
  dataclean snapshot --tag release --tag v1.0
  dataclean snapshot --description "Pre-release snapshot"
  dataclean snapshot --include db_data --include cache_data
  dataclean snapshot --exclude temp_data`,
	Args: cobra.MaximumNArgs(1),
	RunE: runSnapshot,
}

func init() {
	rootCmd.AddCommand(snapshotCmd)

	snapshotCmd.Flags().StringSliceVarP(&snapshotTags, "tag", "t", nil, "Tags to add to snapshot")
	snapshotCmd.Flags().StringVarP(&snapshotDescription, "description", "d", "", "Description for snapshot")
	snapshotCmd.Flags().StringSliceVar(&snapshotInclude, "include", nil, "Only include these volumes")
	snapshotCmd.Flags().StringSliceVar(&snapshotExclude, "exclude", nil, "Exclude these volumes")
}

func runSnapshot(cmd *cobra.Command, args []string) error {
	// Determine snapshot name
	name := fmt.Sprintf("snapshot-%s", time.Now().Format("2006-01-02-150405"))
	if len(args) > 0 {
		name = args[0]
	}

	// Load config (auto-detect or from file)
	cfg, err := config.Load(cfgFile)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Apply include/exclude flags to config
	if len(snapshotInclude) > 0 {
		cfg.IncludeVolumes = snapshotInclude
	}
	if len(snapshotExclude) > 0 {
		cfg.ExcludeVolumes = append(cfg.ExcludeVolumes, snapshotExclude...)
	}

	// Detect volumes
	client, err := docker.NewClient()
	if err != nil {
		return fmt.Errorf("failed to connect to Docker: %w", err)
	}
	defer client.Close()

	volumes, err := client.DetectComposeVolumes(cfg)
	if err != nil {
		return fmt.Errorf("failed to detect volumes: %w", err)
	}

	if len(volumes) == 0 {
		color.Yellow("⚠️  No Docker Compose volumes detected in current directory")
		return nil
	}

	// Show what will be snapshotted
	if !quiet {
		color.Cyan("📸 Creating snapshot: %s", name)
		if snapshotDescription != "" {
			fmt.Printf("   Description: %s\n", snapshotDescription)
		}
		if len(snapshotTags) > 0 {
			fmt.Printf("   Tags: %v\n", snapshotTags)
		}
		fmt.Println()
		for _, v := range volumes {
			fmt.Printf("  • %s (%s)\n", v.Name, v.DatastoreType)
		}
		fmt.Println()
	}

	// Dry run stops here
	if dryRun {
		color.Yellow("🔍 Dry run - no changes made")
		return nil
	}

	// Create snapshot with options
	mgr := snapshot.NewManager(client, cfg)
	opts := snapshot.CreateOptions{
		Tags:        snapshotTags,
		Description: snapshotDescription,
	}
	result, err := mgr.CreateWithOptions(name, volumes, opts)
	if err != nil {
		return fmt.Errorf("failed to create snapshot: %w", err)
	}

	if !quiet {
		color.Green("✅ Snapshot created: %s", result.Name)
		fmt.Printf("   Size: %s\n", result.SizeHuman)
		fmt.Printf("   Path: %s\n", result.Path)
		if len(result.Tags) > 0 {
			fmt.Printf("   Tags: %v\n", result.Tags)
		}
	}

	return nil
}
