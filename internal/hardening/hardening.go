// internal/hardening/hardening.go
package hardening

import (
	"context"

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

func Run(ctx context.Context, exec executor.Executor, s *state.State, sshPort int) ([]StepResult, error) {
	steps := Steps(sshPort)
	var results []StepResult

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
