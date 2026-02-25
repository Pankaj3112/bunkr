// internal/executor/remote_test.go
package executor

import (
	"testing"
)

func TestParseTarget(t *testing.T) {
	tests := []struct {
		input string
		user  string
		host  string
	}{
		{"root@167.71.50.23", "root", "167.71.50.23"},
		{"bunkr@example.com", "bunkr", "example.com"},
		{"myserver", "root", "myserver"},
	}
	for _, tt := range tests {
		user, host := parseTarget(tt.input)
		if user != tt.user || host != tt.host {
			t.Errorf("parseTarget(%q) = %q, %q; want %q, %q", tt.input, user, host, tt.user, tt.host)
		}
	}
}
