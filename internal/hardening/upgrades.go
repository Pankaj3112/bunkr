// internal/hardening/upgrades.go
package hardening

import (
	"context"

	"github.com/pankajbeniwal/bunkr/internal/executor"
)

func UpgradesStep() Step {
	return Step{
		Name:  "unattended_upgrades",
		Label: "Unattended upgrades enabled",
		Check: func(ctx context.Context, exec executor.Executor) (bool, error) {
			_, err := exec.Run(ctx, "dpkg -l | grep -q unattended-upgrades")
			return err == nil, err
		},
		Apply: func(ctx context.Context, exec executor.Executor) error {
			cmds := []string{
				"apt-get -o DPkg::Lock::Timeout=120 install -y unattended-upgrades",
				"dpkg-reconfigure -f noninteractive unattended-upgrades",
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
