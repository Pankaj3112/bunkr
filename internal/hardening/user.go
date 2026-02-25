// internal/hardening/user.go
package hardening

import (
	"context"

	"github.com/pankajbeniwal/bunkr/internal/executor"
)

func UserStep() Step {
	return Step{
		Name:  "sudo_user",
		Label: "Sudo user created",
		Check: func(ctx context.Context, exec executor.Executor) (bool, error) {
			_, err := exec.Run(ctx, "id bunkr")
			return err == nil, err
		},
		Apply: func(ctx context.Context, exec executor.Executor) error {
			cmds := []string{
				"adduser --disabled-password --gecos '' bunkr",
				"usermod -aG sudo bunkr",
				"echo 'bunkr ALL=(ALL) NOPASSWD:ALL' > /etc/sudoers.d/bunkr",
				"mkdir -p /home/bunkr/.ssh",
				"cp /root/.ssh/authorized_keys /home/bunkr/.ssh/authorized_keys",
				"chown -R bunkr:bunkr /home/bunkr/.ssh",
				"chmod 700 /home/bunkr/.ssh",
				"chmod 600 /home/bunkr/.ssh/authorized_keys",
			}
			for _, cmd := range cmds {
				if _, err := exec.Run(ctx, cmd); err != nil {
					return err
				}
			}
			if _, err := exec.Run(ctx, "su - bunkr -c 'whoami'"); err != nil {
				return err
			}
			return nil
		},
	}
}
