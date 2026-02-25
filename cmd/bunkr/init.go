// cmd/bunkr/init.go
package main

import (
	"context"
	"time"

	"github.com/pankajbeniwal/bunkr/internal/executor"
	"github.com/pankajbeniwal/bunkr/internal/hardening"
	"github.com/pankajbeniwal/bunkr/internal/state"
	"github.com/pankajbeniwal/bunkr/internal/ui"
	"github.com/spf13/cobra"
)

var sshPortFlag int

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Harden the server (no app install)",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		exec, err := newExecutor()
		if err != nil {
			return err
		}

		ui.Header("Hardening VPS...")

		s, err := state.Load(ctx, exec)
		if err != nil {
			return err
		}

		_, err = hardening.Run(ctx, exec, s, sshPortFlag)
		if err != nil {
			return err
		}

		s.Hardening.AppliedAt = time.Now()
		s.Hardening.SSHPort = sshPortFlag

		if err := state.Save(ctx, exec, s); err != nil {
			return err
		}

		ui.Result("Server hardened successfully!")
		return nil
	},
}

func init() {
	initCmd.Flags().IntVar(&sshPortFlag, "ssh-port", 2222, "SSH port to configure")
	rootCmd.AddCommand(initCmd)
}

func newExecutor() (executor.Executor, error) {
	if onFlag != "" {
		return executor.NewRemoteExecutor(onFlag)
	}
	return executor.NewLocalExecutor(), nil
}
