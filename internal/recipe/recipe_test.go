// internal/recipe/recipe_test.go
package recipe

import (
	"testing"
)

const testRecipeYAML = `
name: uptime-kuma
version: "1.23.11"
description: Self-hosted uptime monitoring
image: louislam/uptime-kuma:1.23.11

prompts:
  - key: DOMAIN
    label: "Domain for Uptime Kuma"
    required: true

ports:
  - 3001

volumes:
  - uptime_kuma_data:/app/data

health_check:
  url: "http://localhost:3001"
  timeout: 30
  interval: 2
`

func TestParseRecipe(t *testing.T) {
	r, err := Parse([]byte(testRecipeYAML))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if r.Name != "uptime-kuma" {
		t.Fatalf("expected name uptime-kuma, got %s", r.Name)
	}
	if r.Version != "1.23.11" {
		t.Fatalf("expected version 1.23.11, got %s", r.Version)
	}
	if r.Image != "louislam/uptime-kuma:1.23.11" {
		t.Fatalf("expected image louislam/uptime-kuma:1.23.11, got %s", r.Image)
	}
	if len(r.Prompts) != 1 {
		t.Fatalf("expected 1 prompt, got %d", len(r.Prompts))
	}
	if r.Prompts[0].Key != "DOMAIN" || !r.Prompts[0].Required {
		t.Fatalf("unexpected prompt: %+v", r.Prompts[0])
	}
	if len(r.Ports) != 1 || r.Ports[0] != 3001 {
		t.Fatalf("expected ports [3001], got %v", r.Ports)
	}
	if r.HealthCheck == nil || r.HealthCheck.Timeout != 30 {
		t.Fatalf("unexpected health check: %+v", r.HealthCheck)
	}
}

const testMultiServiceYAML = `
name: plausible
version: "2.0.0"
description: Privacy-friendly web analytics
image: ghcr.io/plausible/community-edition:v2.0.0

prompts:
  - key: DOMAIN
    label: "Domain for Plausible"
    required: true
  - key: ADMIN_PASSWORD
    label: "Admin password"
    required: true
    secret: true

ports:
  - 8000

services:
  - name: plausible_db
    image: postgres:16-alpine
    environment:
      POSTGRES_PASSWORD: "postgres"
    volumes:
      - plausible_db:/var/lib/postgresql/data

environment:
  BASE_URL: "https://${DOMAIN}"
  SECRET_KEY_BASE: "auto_generate_64"
`

func TestParseRecipe_MultiService(t *testing.T) {
	r, err := Parse([]byte(testMultiServiceYAML))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(r.Services) != 1 {
		t.Fatalf("expected 1 service, got %d", len(r.Services))
	}
	if r.Services[0].Name != "plausible_db" {
		t.Fatalf("expected service name plausible_db, got %s", r.Services[0].Name)
	}
	if r.Environment["SECRET_KEY_BASE"] != "auto_generate_64" {
		t.Fatalf("expected auto_generate_64, got %s", r.Environment["SECRET_KEY_BASE"])
	}
}

func TestValidateRecipe(t *testing.T) {
	r := &Recipe{Name: "", Version: "1.0", Image: "img", Ports: []int{80}}
	if err := r.Validate(); err == nil {
		t.Fatal("expected validation error for empty name")
	}

	r.Name = "test"
	r.Image = ""
	if err := r.Validate(); err == nil {
		t.Fatal("expected validation error for empty image")
	}
}
