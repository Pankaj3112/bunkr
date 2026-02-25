// internal/hardening/fail2ban.go
package hardening

import (
	"context"

	"github.com/pankajbeniwal/bunkr/internal/executor"
)

func Fail2banStep() Step {
	return Step{
		Name:  "fail2ban",
		Label: "Fail2ban installed",
		Check: func(ctx context.Context, exec executor.Executor) (bool, error) {
			_, err := exec.Run(ctx, "systemctl is-active fail2ban")
			return err == nil, err
		},
		Apply: func(ctx context.Context, exec executor.Executor) error {
			cmds := []string{
				"apt-get install -y fail2ban",
				"systemctl enable fail2ban",
				"systemctl start fail2ban",
			}
			for _, cmd := range cmds {
				if _, err := exec.Run(ctx, cmd); err != nil {
					return err
				}
			}
			return nil
		},
	}
}
