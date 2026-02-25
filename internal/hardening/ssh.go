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
			_, err := exec.Run(ctx, "test -f /etc/ssh/sshd_config.d/99-bunkr.conf")
			return err == nil, err
		},
		Apply: func(ctx context.Context, exec executor.Executor) error {
			// Back up original config
			if _, err := exec.Run(ctx, "cp -n /etc/ssh/sshd_config /etc/ssh/sshd_config.bak"); err != nil {
				return err
			}

			// Apply settings via sshd_config.d drop-in (included by default on modern Ubuntu)
			config := fmt.Sprintf(`Port %d
PermitRootLogin no
PasswordAuthentication no
PubkeyAuthentication yes
X11Forwarding no
MaxAuthTries 3
AllowUsers bunkr
# bunkr-managed`, port)

			if err := exec.WriteFile(ctx, "/etc/ssh/sshd_config.d/99-bunkr.conf", []byte(config), 0644); err != nil {
				return err
			}

			// Validate config before restarting
			if _, err := exec.Run(ctx, "sshd -t"); err != nil {
				exec.Run(ctx, "rm -f /etc/ssh/sshd_config.d/99-bunkr.conf")
				return fmt.Errorf("invalid SSH config: %w", err)
			}

			// Handle systemd socket activation (Ubuntu 24.04+)
			// If ssh.socket exists, we must override it to change the port
			if _, err := exec.Run(ctx, "test -f /lib/systemd/system/ssh.socket"); err == nil {
				override := fmt.Sprintf(`[Socket]
ListenStream=
ListenStream=0.0.0.0:%d
ListenStream=[::]:%d`, port, port)
				if _, err := exec.Run(ctx, "mkdir -p /etc/systemd/system/ssh.socket.d"); err != nil {
					return err
				}
				if err := exec.WriteFile(ctx, "/etc/systemd/system/ssh.socket.d/override.conf", []byte(override), 0644); err != nil {
					return err
				}
				if _, err := exec.Run(ctx, "systemctl daemon-reload && systemctl restart ssh.socket && systemctl restart ssh"); err != nil {
					// Clean up on failure
					exec.Run(ctx, "rm -rf /etc/systemd/system/ssh.socket.d")
					exec.Run(ctx, "rm -f /etc/ssh/sshd_config.d/99-bunkr.conf")
					exec.Run(ctx, "systemctl daemon-reload && systemctl restart ssh.socket && systemctl restart ssh")
					return err
				}
			} else {
				// No socket activation, just restart the service
				if _, err := exec.Run(ctx, "systemctl restart sshd 2>/dev/null || systemctl restart ssh"); err != nil {
					return err
				}
			}

			// Verify SSH is listening on the new port
			if _, err := exec.Run(ctx, fmt.Sprintf("ss -tlnp | grep ':%d '", port)); err != nil {
				// Restore on failure
				exec.Run(ctx, "rm -f /etc/ssh/sshd_config.d/99-bunkr.conf")
				exec.Run(ctx, "rm -rf /etc/systemd/system/ssh.socket.d")
				exec.Run(ctx, "systemctl daemon-reload && (systemctl restart ssh.socket 2>/dev/null; systemctl restart sshd 2>/dev/null || systemctl restart ssh)")
				return fmt.Errorf("SSH not listening on port %d after restart, restored backup", port)
			}

			return nil
		},
	}
}
