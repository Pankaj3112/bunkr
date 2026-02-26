// internal/tailscale/tailscale.go
package tailscale

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/pankajbeniwal/bunkr/internal/executor"
	"github.com/pankajbeniwal/bunkr/internal/ui"
)

// EnsureInstalled checks for tailscale and installs it if missing.
func EnsureInstalled(ctx context.Context, exec executor.Executor) error {
	_, err := exec.Run(ctx, "which tailscale")
	if err == nil {
		return nil
	}

	ui.Info("Installing Tailscale...")
	if _, err := exec.Run(ctx, "curl -fsSL https://tailscale.com/install.sh | sh"); err != nil {
		return fmt.Errorf("failed to install Tailscale: %w", err)
	}
	return nil
}

// IsConnected checks whether tailscale is connected to a tailnet.
func IsConnected(ctx context.Context, exec executor.Executor) (bool, error) {
	out, err := exec.Run(ctx, "tailscale status --json")
	if err != nil {
		return false, nil
	}
	var status struct {
		BackendState string `json:"BackendState"`
	}
	if err := json.Unmarshal([]byte(out), &status); err != nil {
		return false, nil
	}
	return status.BackendState == "Running", nil
}

// Hostname returns the MagicDNS hostname from tailscale status.
func Hostname(ctx context.Context, exec executor.Executor) (string, error) {
	out, err := exec.Run(ctx, "tailscale status --json")
	if err != nil {
		return "", fmt.Errorf("failed to get tailscale status: %w", err)
	}
	var status struct {
		Self struct {
			DNSName string `json:"DNSName"`
		} `json:"Self"`
	}
	if err := json.Unmarshal([]byte(out), &status); err != nil {
		return "", fmt.Errorf("failed to parse tailscale status: %w", err)
	}
	// DNSName has a trailing dot, remove it
	return strings.TrimSuffix(status.Self.DNSName, "."), nil
}

// Connect runs tailscale up in background, prints the auth URL for the user,
// and polls until the node is connected. Returns the MagicDNS hostname.
func Connect(ctx context.Context, exec executor.Executor) (string, error) {
	// Start tailscale up in background, capturing output to a temp file
	exec.Run(ctx, "tailscale up > /tmp/bunkr-ts-auth.log 2>&1 &")

	// Poll for the auth URL (appears almost immediately)
	ui.Info("Waiting for Tailscale auth URL...")
	var authURL string
	for i := 0; i < 30; i++ {
		time.Sleep(1 * time.Second)
		out, err := exec.Run(ctx, "cat /tmp/bunkr-ts-auth.log")
		if err == nil && strings.Contains(out, "https://login.tailscale.com") {
			for _, line := range strings.Split(out, "\n") {
				line = strings.TrimSpace(line)
				if strings.HasPrefix(line, "https://login.tailscale.com") {
					authURL = line
					break
				}
			}
			if authURL != "" {
				break
			}
		}
	}

	if authURL == "" {
		return "", fmt.Errorf("timed out waiting for Tailscale auth URL")
	}

	ui.Info("")
	ui.Info("Open this URL to authenticate Tailscale:")
	ui.Info(fmt.Sprintf("  %s", authURL))
	ui.Info("")
	ui.Info("Waiting for authentication...")

	// Poll for connection (up to 120s for user to authenticate)
	for i := 0; i < 120; i++ {
		time.Sleep(1 * time.Second)
		connected, _ := IsConnected(ctx, exec)
		if connected {
			hostname, err := Hostname(ctx, exec)
			if err != nil {
				return "", err
			}
			// Clean up temp file
			exec.Run(ctx, "rm -f /tmp/bunkr-ts-auth.log")
			return hostname, nil
		}
	}

	exec.Run(ctx, "rm -f /tmp/bunkr-ts-auth.log")
	return "", fmt.Errorf("timed out waiting for Tailscale authentication")
}

// Serve exposes a local port over HTTPS on the tailnet.
func Serve(ctx context.Context, exec executor.Executor, port int) error {
	cmd := fmt.Sprintf("tailscale serve --bg --https=443 http://localhost:%d", port)
	if _, err := exec.Run(ctx, cmd); err != nil {
		return fmt.Errorf("failed to configure tailscale serve: %w", err)
	}
	return nil
}

// RemoveServe stops serving a local port on the tailnet.
func RemoveServe(ctx context.Context, exec executor.Executor, port int) error {
	if _, err := exec.Run(ctx, "tailscale serve --https=443 off"); err != nil {
		return fmt.Errorf("failed to remove tailscale serve: %w", err)
	}
	return nil
}
