// internal/hardening/swap.go
package hardening

import (
	"context"

	"github.com/pankajbeniwal/bunkr/internal/executor"
)

func SwapStep() Step {
	return Step{
		Name:  "swap",
		Label: "Swap configured",
		Check: func(ctx context.Context, exec executor.Executor) (bool, error) {
			_, err := exec.Run(ctx, "swapon --show | grep -q /")
			return err == nil, err
		},
		Apply: func(ctx context.Context, exec executor.Executor) error {
			cmds := []string{
				"fallocate -l 1G /swapfile",
				"chmod 600 /swapfile",
				"mkswap /swapfile",
				"swapon /swapfile",
				"echo '/swapfile none swap sw 0 0' >> /etc/fstab",
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
