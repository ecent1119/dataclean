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

var resetCmd = &cobra.Command{
	Use:   "reset",
	Short: "Wipe all data volumes to empty state",
	Long: `Reset all detected Docker Compose data volumes to empty state.

This is a DESTRUCTIVE operation - all data will be permanently deleted.
Requires --force flag or interactive confirmation.

A backup of current state is automatically created before reset.

Examples:
  dataclean reset          # interactive confirmation
  dataclean reset --force  # skip confirmation
  dataclean reset --dry-run`,
	RunE: runReset,
}

func init() {
	rootCmd.AddCommand(resetCmd)
}

func runReset(cmd *cobra.Command, args []string) error {
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

	// Detect volumes
	volumes, err := client.DetectComposeVolumes(cfg)
	if err != nil {
		return fmt.Errorf("failed to detect volumes: %w", err)
	}

	if len(volumes) == 0 {
		color.Yellow("⚠️  No Docker Compose volumes detected in current directory")
		return nil
	}

	// Show what will be reset
	if !quiet {
		color.Red("🗑️  RESET will DELETE all data in the following volumes:")
		fmt.Println()
		for _, v := range volumes {
			fmt.Printf("  • %s (%s)\n", v.Name, v.DatastoreType)
		}
		fmt.Println()
	}

	// Require confirmation
	if !force && !dryRun {
		color.Red("⚠️  This will PERMANENTLY DELETE all data!")
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

	// Create backup before reset
	if !quiet {
		color.Cyan("📦 Creating backup of current state...")
	}

	// Perform reset
	mgr := snapshot.NewManager(client, cfg)

	if !quiet {
		color.Cyan("🗑️  Resetting volumes...")
	}

	err = mgr.Reset(volumes)
	if err != nil {
		return fmt.Errorf("failed to reset volumes: %w", err)
	}

	if !quiet {
		color.Green("✅ All volumes reset to empty state")
	}

	return nil
}
