// internal/caddy/caddy.go
package caddy

import (
	"context"
	"fmt"
	"strings"

	"github.com/pankajbeniwal/bunkr/internal/executor"
)

const CaddyfilePath = "/etc/caddy/Caddyfile"

// initCaddyfile replaces the default Caddyfile with an empty bunkr-managed one
func initCaddyfile(ctx context.Context, exec executor.Executor) error {
	data, err := exec.ReadFile(ctx, CaddyfilePath)
	if err != nil {
		// No file yet, write empty one
		return exec.WriteFile(ctx, CaddyfilePath, []byte("# Managed by bunkr\n"), 0644)
	}
	if !strings.Contains(string(data), "# Managed by bunkr") && !strings.Contains(string(data), "# bunkr:") {
		// Default Caddyfile from fresh install — replace it
		return exec.WriteFile(ctx, CaddyfilePath, []byte("# Managed by bunkr\n"), 0644)
	}
	return nil
}

func AddBlock(ctx context.Context, exec executor.Executor, name string, domain string, hostPort int) error {
	if err := initCaddyfile(ctx, exec); err != nil {
		return err
	}

	// Remove existing block for this recipe first (prevents duplicates on retry)
	RemoveBlock(ctx, exec, name)

	existing, err := exec.ReadFile(ctx, CaddyfilePath)
	if err != nil {
		existing = []byte{}
	}

	block := fmt.Sprintf(
		"\n# bunkr:%s\n%s {\n    reverse_proxy localhost:%d\n}\n# /bunkr:%s\n",
		name, domain, hostPort, name,
	)

	content := string(existing) + block
	return exec.WriteFile(ctx, CaddyfilePath, []byte(content), 0644)
}

func RemoveBlock(ctx context.Context, exec executor.Executor, name string) error {
	data, err := exec.ReadFile(ctx, CaddyfilePath)
	if err != nil {
		return fmt.Errorf("failed to read Caddyfile: %w", err)
	}

	startMarker := fmt.Sprintf("# bunkr:%s", name)
	endMarker := fmt.Sprintf("# /bunkr:%s", name)

	lines := strings.Split(string(data), "\n")
	var result []string
	inside := false
	for _, line := range lines {
		if strings.TrimSpace(line) == startMarker {
			inside = true
			continue
		}
		if strings.TrimSpace(line) == endMarker {
			inside = false
			continue
		}
		if !inside {
			result = append(result, line)
		}
	}

	return exec.WriteFile(ctx, CaddyfilePath, []byte(strings.Join(result, "\n")), 0644)
}

func Reload(ctx context.Context, exec executor.Executor) error {
	_, err := exec.Run(ctx, "systemctl reload caddy")
	return err
}

func waitForAptLock(ctx context.Context, exec executor.Executor) error {
	_, err := exec.Run(ctx, "fuser /var/lib/dpkg/lock-frontend >/dev/null 2>&1")
	if err != nil {
		return nil // lock not held
	}
	_, err = exec.Run(ctx, "for i in $(seq 1 60); do fuser /var/lib/dpkg/lock-frontend >/dev/null 2>&1 || exit 0; sleep 2; done; exit 1")
	if err != nil {
		return fmt.Errorf("timed out waiting for apt lock")
	}
	return nil
}

func EnsureInstalled(ctx context.Context, exec executor.Executor) error {
	_, err := exec.Run(ctx, "which caddy")
	if err == nil {
		return nil
	}

	if err := waitForAptLock(ctx, exec); err != nil {
		return err
	}

	commands := []string{
		"apt-get install -y debian-keyring debian-archive-keyring apt-transport-https curl",
		"curl -1sLf 'https://dl.cloudsmith.io/public/caddy/stable/gpg.key' | gpg --dearmor -o /usr/share/keyrings/caddy-stable-archive-keyring.gpg",
		"curl -1sLf 'https://dl.cloudsmith.io/public/caddy/stable/debian.deb.txt' | tee /etc/apt/sources.list.d/caddy-stable.list",
		"apt-get update",
		"apt-get install -y caddy",
	}

	for _, cmd := range commands {
		if _, err := exec.Run(ctx, cmd); err != nil {
			return fmt.Errorf("failed to install Caddy: %w", err)
		}
	}
	return nil
}
