package tailscale

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/pankajbeniwal/bunkr/internal/executor"
)

func TestEnsureInstalled_AlreadyPresent(t *testing.T) {
	mock := executor.NewMockExecutor()
	mock.RunOutputs["which tailscale"] = "/usr/bin/tailscale"

	err := EnsureInstalled(context.Background(), mock)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(mock.Calls) != 1 {
		t.Fatalf("expected 1 call (which), got %d", len(mock.Calls))
	}
}

func TestEnsureInstalled_InstallsWhenMissing(t *testing.T) {
	mock := executor.NewMockExecutor()
	mock.RunErrors["which tailscale"] = fmt.Errorf("not found")

	err := EnsureInstalled(context.Background(), mock)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(mock.Calls) != 2 {
		t.Fatalf("expected 2 calls (which + install), got %d", len(mock.Calls))
	}
	installCmd := mock.Calls[1].Args[0].(string)
	if installCmd != "curl -fsSL https://tailscale.com/install.sh | sh" {
		t.Fatalf("unexpected install command: %s", installCmd)
	}
}

func TestEnsureInstalled_InstallFails(t *testing.T) {
	mock := executor.NewMockExecutor()
	mock.RunErrors["which tailscale"] = fmt.Errorf("not found")
	mock.RunErrors["curl -fsSL https://tailscale.com/install.sh | sh"] = fmt.Errorf("curl failed")

	err := EnsureInstalled(context.Background(), mock)
	if err == nil {
		t.Fatal("expected error")
	}
	if err.Error() != "failed to install Tailscale: curl failed" {
		t.Fatalf("unexpected error message: %s", err.Error())
	}
}

func TestIsConnected_Running(t *testing.T) {
	mock := executor.NewMockExecutor()
	mock.RunOutputs["tailscale status --json 2>/dev/null || true"] = `{"BackendState":"Running"}`

	connected, err := IsConnected(context.Background(), mock)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !connected {
		t.Fatal("expected connected=true")
	}
}

func TestIsConnected_NotRunning(t *testing.T) {
	mock := executor.NewMockExecutor()
	mock.RunOutputs["tailscale status --json 2>/dev/null || true"] = `{"BackendState":"Stopped"}`

	connected, err := IsConnected(context.Background(), mock)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if connected {
		t.Fatal("expected connected=false")
	}
}

func TestIsConnected_EmptyOutput(t *testing.T) {
	mock := executor.NewMockExecutor()
	mock.RunOutputs["tailscale status --json 2>/dev/null || true"] = ""

	connected, err := IsConnected(context.Background(), mock)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if connected {
		t.Fatal("expected connected=false for empty output")
	}
}

func TestIsConnected_InvalidJSON(t *testing.T) {
	mock := executor.NewMockExecutor()
	mock.RunOutputs["tailscale status --json 2>/dev/null || true"] = "not json"

	connected, err := IsConnected(context.Background(), mock)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if connected {
		t.Fatal("expected connected=false for invalid JSON")
	}
}

func TestIsConnected_RunError(t *testing.T) {
	mock := executor.NewMockExecutor()
	mock.RunErrors["tailscale status --json 2>/dev/null || true"] = fmt.Errorf("exec failed")

	connected, err := IsConnected(context.Background(), mock)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if connected {
		t.Fatal("expected connected=false on exec error")
	}
}

func TestHostname(t *testing.T) {
	mock := executor.NewMockExecutor()
	mock.RunOutputs["tailscale status --json 2>/dev/null || true"] = `{"Self":{"DNSName":"myhost.tail1234.ts.net."}}`

	hostname, err := Hostname(context.Background(), mock)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if hostname != "myhost.tail1234.ts.net" {
		t.Fatalf("expected myhost.tail1234.ts.net, got %s", hostname)
	}
}

func TestHostname_NoDot(t *testing.T) {
	mock := executor.NewMockExecutor()
	mock.RunOutputs["tailscale status --json 2>/dev/null || true"] = `{"Self":{"DNSName":"myhost.tail1234.ts.net"}}`

	hostname, err := Hostname(context.Background(), mock)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if hostname != "myhost.tail1234.ts.net" {
		t.Fatalf("expected myhost.tail1234.ts.net, got %s", hostname)
	}
}

func TestHostname_EmptyOutput(t *testing.T) {
	mock := executor.NewMockExecutor()
	mock.RunOutputs["tailscale status --json 2>/dev/null || true"] = ""

	_, err := Hostname(context.Background(), mock)
	if err == nil {
		t.Fatal("expected error for empty output")
	}
	if !strings.Contains(err.Error(), "empty output") {
		t.Fatalf("unexpected error message: %s", err.Error())
	}
}

func TestHostname_InvalidJSON(t *testing.T) {
	mock := executor.NewMockExecutor()
	mock.RunOutputs["tailscale status --json 2>/dev/null || true"] = "not json"

	_, err := Hostname(context.Background(), mock)
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
}

func TestServe_Success(t *testing.T) {
	mock := executor.NewMockExecutor()
	cmd := "tailscale serve --bg --https=443 http://localhost:8080 2>&1"
	mock.RunOutputs[cmd] = "Available within your tailnet"

	err := Serve(context.Background(), mock, 8080)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(mock.Calls) != 1 {
		t.Fatalf("expected 1 call, got %d", len(mock.Calls))
	}
}

func TestServe_GenericError(t *testing.T) {
	mock := executor.NewMockExecutor()
	cmd := "tailscale serve --bg --https=443 http://localhost:8080 2>&1"
	mock.RunErrors[cmd] = fmt.Errorf("something went wrong")

	err := Serve(context.Background(), mock, 8080)
	if err == nil {
		t.Fatal("expected error")
	}
	if err.Error() != "failed to configure tailscale serve: something went wrong" {
		t.Fatalf("unexpected error message: %s", err.Error())
	}
}

func TestRemoveServe_Success(t *testing.T) {
	mock := executor.NewMockExecutor()

	err := RemoveServe(context.Background(), mock, 8080)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(mock.Calls) != 1 {
		t.Fatalf("expected 1 call, got %d", len(mock.Calls))
	}
	cmd := mock.Calls[0].Args[0].(string)
	if cmd != "tailscale serve --https=443 off" {
		t.Fatalf("unexpected command: %s", cmd)
	}
}

func TestRemoveServe_Error(t *testing.T) {
	mock := executor.NewMockExecutor()
	mock.RunErrors["tailscale serve --https=443 off"] = fmt.Errorf("not running")

	err := RemoveServe(context.Background(), mock, 8080)
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "not running") {
		t.Fatalf("expected wrapped error, got: %s", err.Error())
	}
}
