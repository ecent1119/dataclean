package cmd

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/fatih/color"
	"github.com/spf13/cobra"

	"github.com/stackgen-cli/dataclean/internal/config"
	"github.com/stackgen-cli/dataclean/internal/docker"
	"github.com/stackgen-cli/dataclean/internal/snapshot"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "Show available snapshots",
	Long: `List all available snapshots for the current project.

Examples:
  dataclean list`,
	Aliases: []string{"ls"},
	RunE:    runList,
}

func init() {
	rootCmd.AddCommand(listCmd)
}

func runList(cmd *cobra.Command, args []string) error {
	// Load config
	cfg, err := config.Load(cfgFile)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Connect to Docker (needed for manager)
	client, err := docker.NewClient()
	if err != nil {
		return fmt.Errorf("failed to connect to Docker: %w", err)
	}
	defer client.Close()

	// List snapshots
	mgr := snapshot.NewManager(client, cfg)
	snapshots, err := mgr.List()
	if err != nil {
		return fmt.Errorf("failed to list snapshots: %w", err)
	}

	if len(snapshots) == 0 {
		color.Yellow("No snapshots found.")
		fmt.Println()
		fmt.Println("Create one with: dataclean snapshot [name]")
		return nil
	}

	// Print table
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "NAME\tCREATED\tSIZE\tVOLUMES")
	fmt.Fprintln(w, "----\t-------\t----\t-------")

	for _, snap := range snapshots {
		fmt.Fprintf(w, "%s\t%s\t%s\t%d\n",
			snap.Name,
			snap.Timestamp.Format("2006-01-02 15:04"),
			snap.SizeHuman,
			len(snap.Volumes),
		)
	}
	w.Flush()

	return nil
}
