// internal/hardening/sysctl.go
package hardening

import (
	"context"

	"github.com/pankajbeniwal/bunkr/internal/executor"
)

func SysctlStep() Step {
	return Step{
		Name:  "sysctl",
		Label: "Kernel parameters hardened",
		Check: func(ctx context.Context, exec executor.Executor) (bool, error) {
			_, err := exec.Run(ctx, "test -f /etc/sysctl.d/99-bunkr.conf")
			return err == nil, err
		},
		Apply: func(ctx context.Context, exec executor.Executor) error {
			config := `# Bunkr kernel hardening
net.ipv4.conf.all.rp_filter = 1
net.ipv4.conf.default.rp_filter = 1
net.ipv4.icmp_echo_ignore_broadcasts = 1
net.ipv4.conf.all.accept_redirects = 0
net.ipv4.conf.default.accept_redirects = 0
net.ipv4.conf.all.send_redirects = 0
net.ipv4.conf.default.send_redirects = 0
net.ipv4.conf.all.accept_source_route = 0
net.ipv4.conf.default.accept_source_route = 0
net.ipv4.tcp_syncookies = 1
`
			if err := exec.WriteFile(ctx, "/etc/sysctl.d/99-bunkr.conf", []byte(config), 0644); err != nil {
				return err
			}
			if _, err := exec.Run(ctx, "sysctl --system"); err != nil {
				return err
			}
			return nil
		},
	}
}
