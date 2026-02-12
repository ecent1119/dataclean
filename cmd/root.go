package cmd

import (
	"fmt"
	"os"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var (
	version   = "1.0.0"
	cfgFile   string
	dryRun    bool
	force     bool
	quiet     bool
)

var rootCmd = &cobra.Command{
	Use:   "dataclean",
	Short: "Snapshot and reset Docker Compose data volumes",
	Long: color.New(color.FgCyan).Sprint(`
dataclean - Local Dev Data Reset & Snapshot Tool

Create, restore, and manage snapshots of Docker Compose data volumes
for rapid development iteration and testing workflows.

Commands:
  snapshot [name]  Create a named snapshot of current data state
  restore <name>   Restore a previously saved snapshot
  reset            Wipe all data volumes to empty state
  list             Show available snapshots

`) + color.New(color.FgYellow).Sprint("For local development and testing only.") + `
Destructive operations require --force or interactive confirmation.`,
	Version: version,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is ./.dataclean.yaml)")
	rootCmd.PersistentFlags().BoolVar(&dryRun, "dry-run", false, "preview changes without executing")
	rootCmd.PersistentFlags().BoolVarP(&force, "force", "f", false, "skip confirmation prompts for destructive operations")
	rootCmd.PersistentFlags().BoolVarP(&quiet, "quiet", "q", false, "minimal output (for CI/scripts)")
}
