// internal/hardening/firewall.go
package hardening

import (
	"context"
	"fmt"

	"github.com/pankajbeniwal/bunkr/internal/executor"
)

func FirewallStep(sshPort int) Step {
	return Step{
		Name:  "firewall",
		Label: "Firewall configured",
		Check: func(ctx context.Context, exec executor.Executor) (bool, error) {
			_, err := exec.Run(ctx, "ufw status | grep -q 'Status: active'")
			return err == nil, err
		},
		Apply: func(ctx context.Context, exec executor.Executor) error {
			cmds := []string{
				"apt-get install -y ufw",
				"ufw default deny incoming",
				"ufw default allow outgoing",
				"ufw allow 22/tcp",
				fmt.Sprintf("ufw allow %d/tcp", sshPort),
				"ufw allow 80/tcp",
				"ufw allow 443/tcp",
				"ufw --force enable",
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
