// cmd/bunkr/root.go
package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

var onFlag string

var rootCmd = &cobra.Command{
	Use:   "bunkr",
	Short: "Harden a VPS and deploy self-hosted apps in one command",
	Long:  "Bunkr takes a fresh VPS and turns it into a hardened server running any self-hosted app.",
}

func init() {
	rootCmd.PersistentFlags().StringVar(&onFlag, "on", "", "remote server to execute on (e.g., root@167.71.50.23)")
	rootCmd.AddCommand(versionCmd)
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print bunkr version",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("bunkr", version)
	},
}
