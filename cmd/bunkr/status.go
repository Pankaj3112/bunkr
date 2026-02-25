// cmd/bunkr/status.go
package main

import (
	"context"
	"fmt"

	"github.com/pankajbeniwal/bunkr/internal/docker"
	"github.com/pankajbeniwal/bunkr/internal/state"
	"github.com/pankajbeniwal/bunkr/internal/ui"
	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show status of installed apps",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := requireRemote(); err != nil {
			return err
		}

		ctx := context.Background()

		exec, err := newExecutor()
		if err != nil {
			return err
		}

		s, err := state.Load(ctx, exec)
		if err != nil {
			return err
		}

		if len(s.Recipes) == 0 {
			ui.Info("No apps installed")
			return nil
		}

		ui.Header("Installed apps")
		fmt.Printf("\n  %-20s %-10s %-30s %-10s %s\n", "NAME", "VERSION", "DOMAIN", "PORT", "STATUS")
		fmt.Printf("  %-20s %-10s %-30s %-10s %s\n", "----", "-------", "------", "----", "------")

		for name, r := range s.Recipes {
			status := "unknown"
			statuses, err := docker.ComposeStatus(ctx, exec, name)
			if err == nil && len(statuses) > 0 {
				status = statuses[0].Status
			}
			fmt.Printf("  %-20s %-10s %-30s %-10d %s\n", name, r.Version, r.Domain, r.Port, status)
		}
		fmt.Println()

		if s.Hardening.Applied {
			ui.Success(fmt.Sprintf("Server hardened (SSH port: %d)", s.Hardening.SSHPort))
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(statusCmd)
}
