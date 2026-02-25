// internal/hardening/ssh.go
package hardening

import (
	"context"
	"fmt"

	"github.com/pankajbeniwal/bunkr/internal/executor"
)

func SSHStep(port int) Step {
	return Step{
		Name:  "ssh_hardening",
		Label: "SSH hardened",
		Check: func(ctx context.Context, exec executor.Executor) (bool, error) {
			_, err := exec.Run(ctx, "grep -q '# bunkr-managed' /etc/ssh/sshd_config")
			return err == nil, err
		},
		Apply: func(ctx context.Context, exec executor.Executor) error {
			// Back up original config
			if _, err := exec.Run(ctx, "cp -n /etc/ssh/sshd_config /etc/ssh/sshd_config.bak"); err != nil {
				return err
			}

			// Apply settings directly to sshd_config (sshd_config.d is not always included)
			cmds := []string{
				fmt.Sprintf("sed -i 's/^#*Port .*/Port %d/' /etc/ssh/sshd_config", port),
				"sed -i 's/^#*PermitRootLogin .*/PermitRootLogin no/' /etc/ssh/sshd_config",
				"sed -i 's/^#*PasswordAuthentication .*/PasswordAuthentication no/' /etc/ssh/sshd_config",
				"sed -i 's/^#*PubkeyAuthentication .*/PubkeyAuthentication yes/' /etc/ssh/sshd_config",
				"sed -i 's/^#*X11Forwarding .*/X11Forwarding no/' /etc/ssh/sshd_config",
				"sed -i 's/^#*MaxAuthTries .*/MaxAuthTries 3/' /etc/ssh/sshd_config",
				// Add AllowUsers if not present
				"grep -q '^AllowUsers' /etc/ssh/sshd_config || echo 'AllowUsers bunkr' >> /etc/ssh/sshd_config",
				// Mark as managed
				"grep -q '# bunkr-managed' /etc/ssh/sshd_config || echo '# bunkr-managed' >> /etc/ssh/sshd_config",
			}
			for _, cmd := range cmds {
				if _, err := exec.Run(ctx, cmd); err != nil {
					return err
				}
			}

			// Validate config before restarting
			if _, err := exec.Run(ctx, "sshd -t"); err != nil {
				// Restore backup if config is invalid
				exec.Run(ctx, "cp /etc/ssh/sshd_config.bak /etc/ssh/sshd_config")
				return fmt.Errorf("invalid SSH config, restored backup: %w", err)
			}

			if _, err := exec.Run(ctx, "systemctl restart sshd 2>/dev/null || systemctl restart ssh"); err != nil {
				return err
			}

			// Verify SSH is listening on the new port
			if _, err := exec.Run(ctx, fmt.Sprintf("ss -tlnp | grep ':%d '", port)); err != nil {
				// Restore backup if SSH didn't come up on new port
				exec.Run(ctx, "cp /etc/ssh/sshd_config.bak /etc/ssh/sshd_config")
				exec.Run(ctx, "systemctl restart sshd 2>/dev/null || systemctl restart ssh")
				return fmt.Errorf("SSH not listening on port %d after restart, restored backup", port)
			}

			return nil
		},
	}
}
