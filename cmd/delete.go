package cmd

import (
	"fmt"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/stackgen-cli/dataclean/internal/config"
	"github.com/stackgen-cli/dataclean/internal/docker"
	"github.com/stackgen-cli/dataclean/internal/snapshot"
	"github.com/stackgen-cli/dataclean/internal/tui"
)

var deleteCmd = &cobra.Command{
	Use:   "delete [snapshot-name]",
	Short: "Delete a saved snapshot",
	Long: `Delete a previously saved snapshot to free up disk space.

If no snapshot name is provided, an interactive selector will be shown.

Examples:
  dataclean delete my-snapshot
  dataclean delete                    # Interactive selection
  dataclean delete my-snapshot -f     # Skip confirmation`,
	Args: cobra.MaximumNArgs(1),
	RunE: runDelete,
}

func init() {
	rootCmd.AddCommand(deleteCmd)
}

func runDelete(cmd *cobra.Command, args []string) error {
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

	// Create snapshot manager
	mgr := snapshot.NewManager(client, cfg)

	// Get snapshot name
	var snapshotName string
	if len(args) > 0 {
		snapshotName = args[0]
	} else {
		// Interactive selection
		snapshots, err := mgr.List()
		if err != nil {
			return fmt.Errorf("failed to list snapshots: %w", err)
		}

		if len(snapshots) == 0 {
			return fmt.Errorf("no snapshots found")
		}

		selected, err := tui.RunSnapshotSelector(snapshots)
		if err != nil {
			return err
		}
		snapshotName = selected.Name
	}

	// Verify snapshot exists
	snap, err := mgr.Get(snapshotName)
	if err != nil {
		return fmt.Errorf("snapshot '%s' not found", snapshotName)
	}

	// Show what will be deleted
	if !quiet {
		color.Cyan("Snapshot to delete:\n")
		fmt.Printf("  Name:      %s\n", snap.Name)
		fmt.Printf("  Created:   %s\n", snap.Timestamp.Format("2006-01-02 15:04:05"))
		fmt.Printf("  Size:      %s\n", snap.SizeHuman)
		fmt.Printf("  Volumes:   %d\n", len(snap.Volumes))
		fmt.Println()
	}

	// Dry run check
	if dryRun {
		color.Yellow("Dry run: would delete snapshot '%s'", snapshotName)
		return nil
	}

	// Confirmation
	if !force {
		color.Yellow("⚠️  This action cannot be undone!")
		fmt.Println()
		confirmed, err := tui.ConfirmDestructive(fmt.Sprintf("Delete snapshot '%s'?", snapshotName))
		if err != nil {
			return err
		}
		if !confirmed {
			fmt.Println("Cancelled.")
			return nil
		}
	}

	// Delete the snapshot
	if !quiet {
		fmt.Printf("Deleting snapshot '%s'...\n", snapshotName)
	}

	if err := mgr.Delete(snapshotName); err != nil {
		return fmt.Errorf("failed to delete snapshot: %w", err)
	}

	if !quiet {
		color.Green("✅ Snapshot '%s' deleted successfully", snapshotName)
	}

	return nil
}
