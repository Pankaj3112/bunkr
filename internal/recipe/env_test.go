// internal/recipe/env_test.go
package recipe

import (
	"strings"
	"testing"
)

func TestGenerateEnv(t *testing.T) {
	values := map[string]string{
		"DOMAIN":      "example.com",
		"ADMIN_EMAIL": "admin@example.com",
	}

	out := GenerateEnv(values)
	s := string(out)

	if !strings.Contains(s, "DOMAIN=example.com") {
		t.Fatal("expected DOMAIN=example.com")
	}
	if !strings.Contains(s, "ADMIN_EMAIL=admin@example.com") {
		t.Fatal("expected ADMIN_EMAIL=admin@example.com")
	}
}

func TestExpandAutoGenerate(t *testing.T) {
	env := map[string]string{
		"KEY":    "auto_generate_32",
		"SECRET": "auto_generate_64",
		"NORMAL": "hello",
	}

	expanded := ExpandAutoGenerate(env)

	if len(expanded["KEY"]) != 32 {
		t.Fatalf("expected 32-char key, got %d", len(expanded["KEY"]))
	}
	if len(expanded["SECRET"]) != 64 {
		t.Fatalf("expected 64-char secret, got %d", len(expanded["SECRET"]))
	}
	if expanded["NORMAL"] != "hello" {
		t.Fatalf("expected 'hello', got %s", expanded["NORMAL"])
	}
}
