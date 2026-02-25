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
	// Wait for apt lock before running Docker install script (it uses apt internally)
	exec.Run(ctx, "for i in $(seq 1 60); do fuser /var/lib/dpkg/lock-frontend >/dev/null 2>&1 || exit 0; sleep 2; done")
	_, err = exec.Run(ctx, "curl -fsSL https://get.docker.com | sh")
	if err != nil {
		return fmt.Errorf("failed to install Docker: %w", err)
	}
	return nil
}

func ComposeUp(ctx context.Context, exec executor.Executor, recipe string) error {
	cmd := fmt.Sprintf("docker compose -f %s up -d", composePath(recipe))
	_, err := exec.Run(ctx, cmd)
	return err
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
