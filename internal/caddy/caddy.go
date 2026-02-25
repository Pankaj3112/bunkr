// internal/caddy/caddy.go
package caddy

import (
	"context"
	"fmt"
	"strings"

	"github.com/pankajbeniwal/bunkr/internal/executor"
)

const CaddyfilePath = "/etc/caddy/Caddyfile"

func AddBlock(ctx context.Context, exec executor.Executor, name string, domain string, hostPort int) error {
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

func EnsureInstalled(ctx context.Context, exec executor.Executor) error {
	_, err := exec.Run(ctx, "which caddy")
	if err == nil {
		return nil
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
