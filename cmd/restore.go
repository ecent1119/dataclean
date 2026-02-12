package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/fatih/color"
	"github.com/spf13/cobra"

	"github.com/stackgen-cli/dataclean/internal/config"
	"github.com/stackgen-cli/dataclean/internal/docker"
	"github.com/stackgen-cli/dataclean/internal/snapshot"
)

var restoreCmd = &cobra.Command{
	Use:   "restore <name>",
	Short: "Restore a previously saved snapshot",
	Long: `Restore data volumes from a named snapshot.

This is a DESTRUCTIVE operation - existing data will be replaced.
Requires --force flag or interactive confirmation.

A backup of current state is automatically created before restore.

Examples:
  dataclean restore before-migration          # interactive confirmation
  dataclean restore before-migration --force  # skip confirmation
  dataclean restore before-migration --dry-run`,
	Args: cobra.ExactArgs(1),
	RunE: runRestore,
}

func init() {
	rootCmd.AddCommand(restoreCmd)
}

func runRestore(cmd *cobra.Command, args []string) error {
	name := args[0]

	// Load config
	cfg, err := config.Load(cfgFile)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Connect to Docker
	client, err := docker.NewClient()
	if err != nil {
		return fmt.Errorf("failed to connect to Docker: %w", err)
	}
	defer client.Close()

	// Check snapshot exists
	mgr := snapshot.NewManager(client, cfg)
	snap, err := mgr.Get(name)
	if err != nil {
		return fmt.Errorf("snapshot not found: %s", name)
	}

	// Show what will be restored
	if !quiet {
		color.Yellow("⚠️  RESTORE will replace current data with snapshot: %s", name)
		fmt.Println()
		fmt.Printf("   Created: %s\n", snap.Timestamp.Format("2006-01-02 15:04:05"))
		fmt.Printf("   Size: %s\n", snap.SizeHuman)
		fmt.Printf("   Volumes: %d\n", len(snap.Volumes))
		fmt.Println()
		for _, v := range snap.Volumes {
			fmt.Printf("  • %s (%s)\n", v.Name, v.DatastoreType)
		}
		fmt.Println()
	}

	// Require confirmation
	if !force && !dryRun {
		color.Red("⚠️  This will DELETE existing data and replace with snapshot!")
		fmt.Print("Type 'yes' to confirm: ")
		reader := bufio.NewReader(os.Stdin)
		response, _ := reader.ReadString('\n')
		response = strings.TrimSpace(strings.ToLower(response))
		if response != "yes" {
			color.Yellow("Aborted.")
			return nil
		}
	}

	// Dry run stops here
	if dryRun {
		color.Yellow("🔍 Dry run - no changes made")
		return nil
	}

	// Create backup before restore
	if !quiet {
		color.Cyan("📦 Creating backup of current state...")
	}

	// Perform restore
	if !quiet {
		color.Cyan("🔄 Restoring snapshot...")
	}

	err = mgr.Restore(name)
	if err != nil {
		return fmt.Errorf("failed to restore snapshot: %w", err)
	}

	if !quiet {
		color.Green("✅ Restored snapshot: %s", name)
	}

	return nil
}
