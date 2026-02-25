// cmd/bunkr/init.go
package main

import (
	"context"
	"fmt"
	"runtime"
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
		if err := requireRemote(); err != nil {
			return err
		}

		ctx := context.Background()
		exec, err := newExecutor()
		if err != nil {
			return err
		}

		// Set system-wide apt lock timeout (fresh VPS often has apt running)
		exec.Run(ctx, `echo 'DPkg::Lock::Timeout "120";' > /etc/apt/apt.conf.d/99-bunkr-lock-wait`)

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

func requireRemote() error {
	if onFlag == "" && runtime.GOOS != "linux" {
		return fmt.Errorf("--on flag is required on %s (e.g., --on root@167.71.50.23)\n\nbunkr server commands run on Linux. Use --on to target a remote server,\nor run bunkr directly on a Linux machine.", runtime.GOOS)
	}
	return nil
}
