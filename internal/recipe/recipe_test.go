package recipe

import (
	"strings"
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

func TestParseRecipe_Display(t *testing.T) {
	yaml := `
name: test-app
version: "1.0.0"
description: Test app with display
image: test:latest

ports:
  - 8080

environment:
  SECRET_KEY: "auto_generate_64"

display:
  - key: SECRET_KEY
    label: "Secret Key"
`
	r, err := Parse([]byte(yaml))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(r.Display) != 1 {
		t.Fatalf("expected 1 display var, got %d", len(r.Display))
	}
	if r.Display[0].Key != "SECRET_KEY" {
		t.Fatalf("expected key SECRET_KEY, got %s", r.Display[0].Key)
	}
	if r.Display[0].Label != "Secret Key" {
		t.Fatalf("expected label 'Secret Key', got %s", r.Display[0].Label)
	}
}

func TestParseRecipe_SelectPrompt(t *testing.T) {
	yaml := `
name: test-select
version: "1.0.0"
description: Test select prompt
image: test:latest

prompts:
  - key: TIMEZONE
    label: "Timezone"
    options:
      - UTC
      - America/New_York
      - Europe/London
    default: "UTC"

ports:
  - 8080
`
	r, err := Parse([]byte(yaml))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(r.Prompts) != 1 {
		t.Fatalf("expected 1 prompt, got %d", len(r.Prompts))
	}
	p := r.Prompts[0]
	if len(p.Options) != 3 {
		t.Fatalf("expected 3 options, got %d", len(p.Options))
	}
	if p.Options[0] != "UTC" {
		t.Fatalf("expected first option UTC, got %s", p.Options[0])
	}
	if p.Default != "UTC" {
		t.Fatalf("expected default UTC, got %s", p.Default)
	}
}

func TestParseRecipe_N8N(t *testing.T) {
	yamlContent := `
name: n8n
version: "2.12.2"
description: Workflow automation platform
image: docker.n8n.io/n8nio/n8n:2.12.2

prompts:
  - key: DOMAIN
    label: "Domain for n8n"
    required: true
  - key: GENERIC_TIMEZONE
    label: "Timezone"
    options:
      - UTC
      - America/New_York
    default: "UTC"

ports:
  - 5678

volumes:
  - n8n_data:/home/node/.n8n

services:
  - name: n8n_db
    image: postgres:16-alpine
    environment:
      POSTGRES_USER: "n8n"
      POSTGRES_PASSWORD: "${N8N_DB_PASSWORD}"
      POSTGRES_DB: "n8n"
    volumes:
      - n8n_db:/var/lib/postgresql/data

environment:
  DB_TYPE: "postgresdb"
  N8N_DB_PASSWORD: "auto_generate_32"
  N8N_ENCRYPTION_KEY: "auto_generate_64"
  DB_POSTGRESDB_PASSWORD: "${N8N_DB_PASSWORD}"
  N8N_HOST: "${DOMAIN}"
  GENERIC_TIMEZONE: "${GENERIC_TIMEZONE}"

display:
  - key: N8N_ENCRYPTION_KEY
    label: "Encryption Key"

health_check:
  url: "http://localhost:5678/healthz"
  timeout: 60
  interval: 3
`
	r, err := Parse([]byte(yamlContent))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if r.Name != "n8n" {
		t.Fatalf("expected name n8n, got %s", r.Name)
	}
	if len(r.Prompts) != 2 {
		t.Fatalf("expected 2 prompts, got %d", len(r.Prompts))
	}
	if len(r.Prompts[1].Options) != 2 {
		t.Fatalf("expected 2 options on timezone prompt, got %d", len(r.Prompts[1].Options))
	}
	if len(r.Display) != 1 || r.Display[0].Key != "N8N_ENCRYPTION_KEY" {
		t.Fatalf("unexpected display: %+v", r.Display)
	}
	if len(r.Services) != 1 || r.Services[0].Name != "n8n_db" {
		t.Fatalf("unexpected services: %+v", r.Services)
	}

	// Test full expansion flow
	r.Environment = ExpandAutoGenerate(r.Environment)
	if len(r.Environment["N8N_DB_PASSWORD"]) != 32 {
		t.Fatalf("expected 32-char password, got %d", len(r.Environment["N8N_DB_PASSWORD"]))
	}
	if len(r.Environment["N8N_ENCRYPTION_KEY"]) != 64 {
		t.Fatalf("expected 64-char key, got %d", len(r.Environment["N8N_ENCRYPTION_KEY"]))
	}

	// Test merged values expansion in compose
	promptValues := map[string]string{
		"DOMAIN":           "n8n.example.com",
		"GENERIC_TIMEZONE": "UTC",
	}
	merged := MergeValues(promptValues, r.Environment)

	out, err := GenerateCompose(r, merged, 5678)
	if err != nil {
		t.Fatalf("compose error: %v", err)
	}
	s := string(out)

	if strings.Contains(s, "${N8N_DB_PASSWORD}") {
		t.Fatal("found unexpanded ${N8N_DB_PASSWORD} in compose output")
	}
	if strings.Contains(s, "${DOMAIN}") {
		t.Fatal("found unexpanded ${DOMAIN} in compose output")
	}
	if !strings.Contains(s, "n8n.example.com") {
		t.Fatal("expected domain expansion in compose output")
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
