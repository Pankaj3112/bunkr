// internal/executor/local_test.go
package executor

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestLocalExecutor_Run(t *testing.T) {
	exec := NewLocalExecutor()
	out, err := exec.Run(context.Background(), "echo hello")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out != "hello\n" {
		t.Fatalf("expected 'hello\\n', got %q", out)
	}
}

func TestLocalExecutor_Run_Error(t *testing.T) {
	exec := NewLocalExecutor()
	_, err := exec.Run(context.Background(), "false")
	if err == nil {
		t.Fatal("expected error from 'false' command")
	}
}

func TestLocalExecutor_WriteReadFile(t *testing.T) {
	exec := NewLocalExecutor()
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "test.txt")

	err := exec.WriteFile(ctx, path, []byte("hello world"), 0644)
	if err != nil {
		t.Fatalf("unexpected write error: %v", err)
	}

	data, err := exec.ReadFile(ctx, path)
	if err != nil {
		t.Fatalf("unexpected read error: %v", err)
	}
	if string(data) != "hello world" {
		t.Fatalf("expected 'hello world', got %q", string(data))
	}

	info, _ := os.Stat(path)
	if info.Mode().Perm() != 0644 {
		t.Fatalf("expected mode 0644, got %v", info.Mode().Perm())
	}
}
