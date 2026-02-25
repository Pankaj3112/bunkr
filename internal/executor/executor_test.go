// internal/executor/executor_test.go
package executor

import (
	"context"
	"testing"
)

func TestMockExecutor_Run(t *testing.T) {
	m := NewMockExecutor()
	m.RunOutputs["echo hello"] = "hello\n"

	out, err := m.Run(context.Background(), "echo hello")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out != "hello\n" {
		t.Fatalf("expected 'hello\\n', got %q", out)
	}
	if len(m.Calls) != 1 || m.Calls[0].Method != "Run" {
		t.Fatalf("expected 1 Run call, got %v", m.Calls)
	}
}

func TestMockExecutor_WriteReadFile(t *testing.T) {
	m := NewMockExecutor()
	ctx := context.Background()

	err := m.WriteFile(ctx, "/tmp/test", []byte("data"), 0644)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, err := m.ReadFile(ctx, "/tmp/test")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(data) != "data" {
		t.Fatalf("expected 'data', got %q", string(data))
	}
}

func TestMockExecutor_ReadFile_NotFound(t *testing.T) {
	m := NewMockExecutor()
	_, err := m.ReadFile(context.Background(), "/nonexistent")
	if err == nil {
		t.Fatal("expected error for missing file")
	}
}
