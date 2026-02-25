// internal/hardening/hardening.go
package hardening

import (
	"context"
	"fmt"

	"github.com/pankajbeniwal/bunkr/internal/executor"
	"github.com/pankajbeniwal/bunkr/internal/state"
	"github.com/pankajbeniwal/bunkr/internal/ui"
)

type Step struct {
	Name  string
	Label string
	Check func(ctx context.Context, exec executor.Executor) (bool, error)
	Apply func(ctx context.Context, exec executor.Executor) error
}

type StepResult struct {
	Name    string
	Skipped bool
	Error   error
}

func Steps(sshPort int) []Step {
	return []Step{
		UserStep(),
		SSHStep(sshPort),
		FirewallStep(sshPort),
		Fail2banStep(),
		UpgradesStep(),
		SysctlStep(),
		SwapStep(),
	}
}

func waitForAptLock(ctx context.Context, exec executor.Executor) error {
	// Fresh VPS often has unattended-upgrades running apt on first boot.
	// Check if lock is held, and if so, wait up to 120 seconds.
	_, err := exec.Run(ctx, "fuser /var/lib/dpkg/lock-frontend >/dev/null 2>&1")
	if err != nil {
		// Lock is not held, no need to wait
		return nil
	}
	ui.Info("Waiting for package manager to finish...")
	_, err = exec.Run(ctx, "for i in $(seq 1 60); do fuser /var/lib/dpkg/lock-frontend >/dev/null 2>&1 || exit 0; sleep 2; done; exit 1")
	if err != nil {
		return fmt.Errorf("timed out waiting for apt lock (another package manager is running)")
	}
	return nil
}

func Run(ctx context.Context, exec executor.Executor, s *state.State, sshPort int) ([]StepResult, error) {
	steps := Steps(sshPort)
	var results []StepResult

	// Wait for any existing apt operations to finish (common on fresh VPS)
	if err := waitForAptLock(ctx, exec); err != nil {
		return nil, err
	}

	for _, step := range steps {
		if s.Hardening.Steps[step.Name] {
			ui.Skip(step.Label + " already configured")
			results = append(results, StepResult{Name: step.Name, Skipped: true})
			continue
		}

		applied, err := step.Check(ctx, exec)
		if err == nil && applied {
			ui.Skip(step.Label + " already configured")
			s.Hardening.Steps[step.Name] = true
			results = append(results, StepResult{Name: step.Name, Skipped: true})
			continue
		}

		if err := step.Apply(ctx, exec); err != nil {
			ui.Error(step.Label + " failed: " + err.Error())
			results = append(results, StepResult{Name: step.Name, Error: err})
			return results, err
		}

		ui.Success(step.Label)
		s.Hardening.Steps[step.Name] = true
		results = append(results, StepResult{Name: step.Name})
	}

	s.Hardening.Applied = true
	return results, nil
}
