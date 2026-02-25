// cmd/bunkr/uninstall.go
package main

import (
	"context"
	"fmt"

	"github.com/pankajbeniwal/bunkr/internal/caddy"
	"github.com/pankajbeniwal/bunkr/internal/docker"
	"github.com/pankajbeniwal/bunkr/internal/state"
	"github.com/pankajbeniwal/bunkr/internal/ui"
	"github.com/spf13/cobra"
)

var purgeFlag bool

var uninstallCmd = &cobra.Command{
	Use:   "uninstall <recipe>",
	Short: "Remove an installed app",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := requireRemote(); err != nil {
			return err
		}

		ctx := context.Background()
		name := args[0]

		exec, err := newExecutor()
		if err != nil {
			return err
		}

		s, err := state.Load(ctx, exec)
		if err != nil {
			return err
		}

		if _, ok := s.Recipes[name]; !ok {
			return fmt.Errorf("recipe %s is not installed", name)
		}

		ui.Header(fmt.Sprintf("Uninstalling %s...", name))

		// Stop containers
		if err := docker.ComposeDown(ctx, exec, name, purgeFlag); err != nil {
			ui.Warn("Failed to stop containers: " + err.Error())
		}
		ui.Success("Containers stopped")

		// Remove Caddy block
		if err := caddy.RemoveBlock(ctx, exec, name); err != nil {
			ui.Warn("Failed to remove Caddy config: " + err.Error())
		}
		ui.Success("Caddy config removed")

		// Reload Caddy
		if err := caddy.Reload(ctx, exec); err != nil {
			ui.Warn("Caddy reload failed")
		}

		// Remove directory
		dir := fmt.Sprintf("/opt/bunkr/%s", name)
		if _, err := exec.Run(ctx, fmt.Sprintf("rm -rf %s", dir)); err != nil {
			ui.Warn("Failed to remove directory: " + err.Error())
		}
		ui.Success("Files removed")

		// Update state
		delete(s.Recipes, name)
		if err := state.Save(ctx, exec, s); err != nil {
			return err
		}

		ui.Result(fmt.Sprintf("%s has been uninstalled", name))
		return nil
	},
}

func init() {
	uninstallCmd.Flags().BoolVar(&purgeFlag, "purge", false, "also remove volumes (data)")
	rootCmd.AddCommand(uninstallCmd)
}
