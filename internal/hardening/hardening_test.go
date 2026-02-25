// internal/hardening/hardening_test.go
package hardening

import (
	"context"
	"fmt"
	"testing"

	"github.com/pankajbeniwal/bunkr/internal/executor"
	"github.com/pankajbeniwal/bunkr/internal/state"
)

func TestRunSteps_AllNew(t *testing.T) {
	mock := executor.NewMockExecutor()
	s := state.New()
	ctx := context.Background()

	// Make all checks return "not applied"
	mock.RunErrors["id bunkr"] = fmt.Errorf("no such user")
	mock.RunErrors["test -f /etc/ssh/sshd_config.d/99-bunkr.conf"] = fmt.Errorf("not found")
	mock.RunErrors["ufw status | grep -q 'Status: active'"] = fmt.Errorf("inactive")
	mock.RunErrors["systemctl is-active fail2ban"] = fmt.Errorf("inactive")
	mock.RunErrors["dpkg -l | grep -q unattended-upgrades"] = fmt.Errorf("not installed")
	mock.RunErrors["test -f /etc/sysctl.d/99-bunkr.conf"] = fmt.Errorf("not found")
	mock.RunErrors["swapon --show | grep -q /"] = fmt.Errorf("no swap")

	results, err := Run(ctx, mock, s, 2222)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 7 {
		t.Fatalf("expected 7 results, got %d", len(results))
	}
	for _, r := range results {
		if r.Skipped {
			t.Fatalf("expected step %s to not be skipped", r.Name)
		}
	}
}

func TestRunSteps_AllSkipped(t *testing.T) {
	mock := executor.NewMockExecutor()
	s := state.New()
	s.Hardening.Applied = true
	for _, step := range Steps(2222) {
		s.Hardening.Steps[step.Name] = true
	}

	results, err := Run(context.Background(), mock, s, 2222)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for _, r := range results {
		if !r.Skipped {
			t.Fatalf("expected step %s to be skipped", r.Name)
		}
	}
}
