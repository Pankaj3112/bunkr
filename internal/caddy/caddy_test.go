// internal/caddy/caddy_test.go
package caddy

import (
	"context"
	"strings"
	"testing"

	"github.com/pankajbeniwal/bunkr/internal/executor"
)

func TestAddBlock(t *testing.T) {
	mock := executor.NewMockExecutor()
	mock.Files[CaddyfilePath] = []byte("# Existing config\n")

	err := AddBlock(context.Background(), mock, "uptime-kuma", "status.example.com", 3001)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	content := string(mock.Files[CaddyfilePath])
	if !strings.Contains(content, "# bunkr:uptime-kuma") {
		t.Fatal("expected bunkr marker")
	}
	if !strings.Contains(content, "status.example.com") {
		t.Fatal("expected domain")
	}
	if !strings.Contains(content, "reverse_proxy localhost:3001") {
		t.Fatal("expected reverse_proxy directive")
	}
	if !strings.Contains(content, "# /bunkr:uptime-kuma") {
		t.Fatal("expected closing marker")
	}
}

func TestRemoveBlock(t *testing.T) {
	mock := executor.NewMockExecutor()
	mock.Files[CaddyfilePath] = []byte(`# Existing
# bunkr:ghost
blog.example.com {
    reverse_proxy localhost:2368
}
# /bunkr:ghost
# Other config
`)

	err := RemoveBlock(context.Background(), mock, "ghost")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	content := string(mock.Files[CaddyfilePath])
	if strings.Contains(content, "bunkr:ghost") {
		t.Fatal("expected ghost block to be removed")
	}
	if !strings.Contains(content, "# Existing") {
		t.Fatal("expected other config to remain")
	}
	if !strings.Contains(content, "# Other config") {
		t.Fatal("expected trailing config to remain")
	}
}

func TestAddBlock_NoCaddyfile(t *testing.T) {
	mock := executor.NewMockExecutor()
	// No existing Caddyfile — ReadFile will return error

	err := AddBlock(context.Background(), mock, "ghost", "blog.example.com", 2368)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	content := string(mock.Files[CaddyfilePath])
	if !strings.Contains(content, "blog.example.com") {
		t.Fatal("expected domain in new Caddyfile")
	}
}
