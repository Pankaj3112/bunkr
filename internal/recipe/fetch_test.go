// internal/recipe/fetch_test.go
package recipe

import (
	"testing"
)

func TestBuildRecipeURL(t *testing.T) {
	url := BuildRecipeURL("uptime-kuma", "")
	expected := "https://raw.githubusercontent.com/Pankaj3112/bunkr/main/recipes/uptime-kuma.yaml"
	if url != expected {
		t.Fatalf("expected %s, got %s", expected, url)
	}
}

func TestBuildRecipeURL_Override(t *testing.T) {
	url := BuildRecipeURL("ghost", "https://example.com/recipes")
	expected := "https://example.com/recipes/ghost.yaml"
	if url != expected {
		t.Fatalf("expected %s, got %s", expected, url)
	}
}

func TestBuildIndexURL(t *testing.T) {
	url := BuildIndexURL("")
	expected := "https://raw.githubusercontent.com/Pankaj3112/bunkr/main/recipes/index.yaml"
	if url != expected {
		t.Fatalf("expected %s, got %s", expected, url)
	}
}
