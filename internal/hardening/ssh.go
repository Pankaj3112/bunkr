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
			_, err := exec.Run(ctx, "grep -q 'PermitRootLogin no' /etc/ssh/sshd_config")
			return err == nil, err
		},
		Apply: func(ctx context.Context, exec executor.Executor) error {
			config := fmt.Sprintf(`Port %d
PermitRootLogin no
PasswordAuthentication no
PubkeyAuthentication yes
X11Forwarding no
MaxAuthTries 3
AllowUsers bunkr`, port)

			if err := exec.WriteFile(ctx, "/etc/ssh/sshd_config.d/99-bunkr.conf", []byte(config), 0644); err != nil {
				return err
			}
			if _, err := exec.Run(ctx, "systemctl restart sshd 2>/dev/null || systemctl restart ssh"); err != nil {
				return err
			}
			return nil
		},
	}
}
