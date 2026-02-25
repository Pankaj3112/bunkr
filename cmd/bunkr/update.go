// cmd/bunkr/update.go
package main

import (
	"context"
	"fmt"

	"github.com/pankajbeniwal/bunkr/internal/docker"
	"github.com/pankajbeniwal/bunkr/internal/recipe"
	"github.com/pankajbeniwal/bunkr/internal/state"
	"github.com/pankajbeniwal/bunkr/internal/ui"
	"github.com/spf13/cobra"
)

var updateCmd = &cobra.Command{
	Use:   "update <recipe>",
	Short: "Update an installed app to the latest version",
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

		current, ok := s.Recipes[name]
		if !ok {
			return fmt.Errorf("recipe %s is not installed", name)
		}

		ui.Header(fmt.Sprintf("Checking for updates to %s...", name))

		latest, err := recipe.Fetch(name)
		if err != nil {
			return err
		}

		if latest.Version == current.Version {
			ui.Info(fmt.Sprintf("%s is already at version %s", name, current.Version))
			return nil
		}

		ui.Info(fmt.Sprintf("Updating %s: %s → %s", name, current.Version, latest.Version))

		// Pull new images
		if err := docker.ComposePull(ctx, exec, name); err != nil {
			return fmt.Errorf("failed to pull images: %w", err)
		}
		ui.Success("Images pulled")

		// Regenerate compose file with new image
		composeData, err := recipe.GenerateCompose(latest, map[string]string{"DOMAIN": current.Domain}, current.Port)
		if err != nil {
			return err
		}

		dir := fmt.Sprintf("/opt/bunkr/%s", name)
		if err := exec.WriteFile(ctx, dir+"/docker-compose.yml", composeData, 0644); err != nil {
			return err
		}

		// Restart
		if err := docker.ComposeDown(ctx, exec, name, false); err != nil {
			return err
		}
		if err := docker.ComposeUp(ctx, exec, name); err != nil {
			return err
		}
		ui.Success("Containers restarted")

		// Update state
		current.Version = latest.Version
		s.Recipes[name] = current
		if err := state.Save(ctx, exec, s); err != nil {
			return err
		}

		ui.Result(fmt.Sprintf("%s updated to %s", name, latest.Version))
		return nil
	},
}

func init() {
	rootCmd.AddCommand(updateCmd)
}
