// internal/recipe/compose_test.go
package recipe

import (
	"strings"
	"testing"
)

func TestGenerateCompose_Simple(t *testing.T) {
	r := &Recipe{
		Name:    "uptime-kuma",
		Image:   "louislam/uptime-kuma:1.23.11",
		Ports:   []int{3001},
		Volumes: []string{"uptime_kuma_data:/app/data"},
	}
	values := map[string]string{"DOMAIN": "status.example.com"}

	out, err := GenerateCompose(r, values, 3001)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	s := string(out)

	if !strings.Contains(s, "uptime-kuma:") {
		t.Fatal("expected service name uptime-kuma")
	}
	if !strings.Contains(s, "louislam/uptime-kuma:1.23.11") {
		t.Fatal("expected image in output")
	}
	if !strings.Contains(s, "127.0.0.1:3001:3001") {
		t.Fatal("expected port mapping 127.0.0.1:3001:3001")
	}
	if !strings.Contains(s, "uptime_kuma_data:/app/data") {
		t.Fatal("expected volume mount")
	}
}

func TestGenerateCompose_MultiService(t *testing.T) {
	r := &Recipe{
		Name:  "plausible",
		Image: "ghcr.io/plausible/community-edition:v2.0.0",
		Ports: []int{8000},
		Services: []Service{
			{
				Name:        "plausible_db",
				Image:       "postgres:16-alpine",
				Environment: map[string]string{"POSTGRES_PASSWORD": "postgres"},
				Volumes:     []string{"plausible_db:/var/lib/postgresql/data"},
			},
		},
		Environment: map[string]string{
			"BASE_URL": "https://${DOMAIN}",
		},
	}
	values := map[string]string{"DOMAIN": "analytics.example.com"}

	out, err := GenerateCompose(r, values, 8000)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	s := string(out)

	if !strings.Contains(s, "plausible_db:") {
		t.Fatal("expected plausible_db service")
	}
	if !strings.Contains(s, "postgres:16-alpine") {
		t.Fatal("expected postgres image")
	}
}

func TestGenerateCompose_ServiceEnvExpansion(t *testing.T) {
	r := &Recipe{
		Name:  "n8n",
		Image: "n8n:latest",
		Ports: []int{5678},
		Services: []Service{
			{
				Name:  "n8n_db",
				Image: "postgres:16-alpine",
				Environment: map[string]string{
					"POSTGRES_PASSWORD": "${DB_PASS}",
					"POSTGRES_USER":     "n8n",
				},
				Volumes: []string{"n8n_db:/var/lib/postgresql/data"},
			},
		},
		Environment: map[string]string{
			"DB_PASS": "secret123",
		},
	}
	merged := map[string]string{
		"DOMAIN":  "n8n.example.com",
		"DB_PASS": "secret123",
	}

	out, err := GenerateCompose(r, merged, 5678)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	s := string(out)

	if strings.Contains(s, "${DB_PASS}") {
		t.Fatal("expected service env to be expanded, but found literal ${DB_PASS}")
	}
	if !strings.Contains(s, "secret123") {
		t.Fatal("expected expanded password 'secret123' in service env")
	}
}

func TestGenerateCompose_DifferentHostPort(t *testing.T) {
	r := &Recipe{
		Name:  "app",
		Image: "app:latest",
		Ports: []int{3000},
	}

	out, err := GenerateCompose(r, nil, 3005)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(string(out), "127.0.0.1:3005:3000") {
		t.Fatal("expected port mapping 127.0.0.1:3005:3000")
	}
}
