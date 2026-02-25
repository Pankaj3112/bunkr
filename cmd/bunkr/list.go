// cmd/bunkr/list.go
package main

import (
	"fmt"

	"github.com/pankajbeniwal/bunkr/internal/recipe"
	"github.com/pankajbeniwal/bunkr/internal/ui"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List available recipes",
	RunE: func(cmd *cobra.Command, args []string) error {
		ui.Header("Available recipes")

		index, err := recipe.FetchIndex()
		if err != nil {
			return err
		}

		fmt.Printf("\n  %-20s %-10s %s\n", "NAME", "VERSION", "DESCRIPTION")
		fmt.Printf("  %-20s %-10s %s\n", "----", "-------", "-----------")
		for _, entry := range index {
			fmt.Printf("  %-20s %-10s %s\n", entry.Name, entry.Version, entry.Description)
		}
		fmt.Println()

		return nil
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
}
