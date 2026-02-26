// internal/docker/docker.go
package docker

import (
	"context"
	"fmt"
	"strings"

	"github.com/pankajbeniwal/bunkr/internal/executor"
)

const basePath = "/opt/bunkr"

func composePath(recipe string) string {
	return fmt.Sprintf("%s/%s/docker-compose.yml", basePath, recipe)
}

func EnsureInstalled(ctx context.Context, exec executor.Executor) error {
	_, err := exec.Run(ctx, "docker --version")
	if err == nil {
		return nil
	}
	// Wait for apt locks to be released (fresh VPS often has unattended-upgrades running)
	waitCmd := `for i in $(seq 1 60); do fuser /var/lib/apt/lists/lock /var/lib/dpkg/lock-frontend /var/cache/apt/archives/lock >/dev/null 2>&1 || break; sleep 2; done`
	exec.Run(ctx, waitCmd)
	_, err = exec.Run(ctx, "curl -fsSL https://get.docker.com | sh")
	if err != nil {
		return fmt.Errorf("failed to install Docker: %w", err)
	}
	return nil
}

func ComposeUp(ctx context.Context, exec executor.Executor, recipe string) error {
	// Pull images first as a separate step so the up command doesn't block
	// on a long download with no feedback.
	pullCmd := fmt.Sprintf("docker compose -f %s pull 2>&1", composePath(recipe))
	if _, err := exec.Run(ctx, pullCmd); err != nil {
		return fmt.Errorf("failed to pull images: %w", err)
	}

	cmd := fmt.Sprintf("docker compose -f %s up -d", composePath(recipe))
	_, err := exec.Run(ctx, cmd)
	return err
}

// RunInit runs a one-off command using the recipe's image and volumes via
// docker compose run. Used for init commands like "openclaw setup" that
// must run before the main service starts.
func RunInit(ctx context.Context, exec executor.Executor, recipe string, initCmd string) error {
	cmd := fmt.Sprintf("docker compose -f %s run --rm --no-deps %s %s 2>&1",
		composePath(recipe), recipe, initCmd)
	_, err := exec.Run(ctx, cmd)
	return err
}

// RunPostInit writes the post_init commands to a shell script on the host,
// bind-mounts it into the container, and executes it via docker compose run.
// This avoids shell quoting issues from double-escaping through SSH+sudo.
func RunPostInit(ctx context.Context, exec executor.Executor, recipe string, commands []string) error {
	dir := fmt.Sprintf("%s/%s", basePath, recipe)
	scriptPath := dir + "/post-init.sh"

	// Build script content
	script := "#!/bin/sh\nset -e\n"
	for _, cmd := range commands {
		script += cmd + "\n"
	}

	// Write script to host
	if err := exec.WriteFile(ctx, scriptPath, []byte(script), 0755); err != nil {
		return fmt.Errorf("failed to write post-init script: %w", err)
	}

	// Run script inside container via bind-mount (--entrypoint overrides
	// the image entrypoint so "sh" is executed directly)
	cmd := fmt.Sprintf(
		"docker compose -f %s run --rm --no-deps --entrypoint sh -v %s:/tmp/bunkr-post-init.sh:ro %s /tmp/bunkr-post-init.sh 2>&1",
		composePath(recipe), scriptPath, recipe,
	)
	_, err := exec.Run(ctx, cmd)
	if err != nil {
		return fmt.Errorf("post-init failed: %w", err)
	}
	return nil
}

func ComposeDown(ctx context.Context, exec executor.Executor, recipe string, purge bool) error {
	cmd := fmt.Sprintf("docker compose -f %s down", composePath(recipe))
	if purge {
		cmd += " -v"
	}
	_, err := exec.Run(ctx, cmd)
	return err
}

func ComposePull(ctx context.Context, exec executor.Executor, recipe string) error {
	cmd := fmt.Sprintf("docker compose -f %s pull", composePath(recipe))
	_, err := exec.Run(ctx, cmd)
	return err
}

type ServiceStatus struct {
	Name   string
	Status string // "running" or "exited"
}

func ComposeStatus(ctx context.Context, exec executor.Executor, recipe string) ([]ServiceStatus, error) {
	cmd := fmt.Sprintf("docker compose -f %s ps --format '{{.Name}} {{.State}}'", composePath(recipe))
	out, err := exec.Run(ctx, cmd)
	if err != nil {
		return nil, err
	}

	var statuses []ServiceStatus
	for _, line := range strings.Split(strings.TrimSpace(out), "\n") {
		parts := strings.Fields(line)
		if len(parts) >= 2 {
			statuses = append(statuses, ServiceStatus{Name: parts[0], Status: parts[1]})
		}
	}
	return statuses, nil
}

func HealthCheck(ctx context.Context, exec executor.Executor, url string, timeout, interval int) error {
	cmd := fmt.Sprintf(
		"for i in $(seq 1 %d); do curl -sf %s > /dev/null 2>&1 && exit 0; sleep %d; done; exit 1",
		timeout/interval, url, interval,
	)
	_, err := exec.Run(ctx, cmd)
	if err != nil {
		return fmt.Errorf("health check failed after %ds: %w", timeout, err)
	}
	return nil
}
