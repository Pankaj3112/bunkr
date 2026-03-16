package recipe

import (
	"bufio"
	"strings"
	"testing"
)

func TestPromptSelect_ValidChoice(t *testing.T) {
	reader := bufio.NewReader(strings.NewReader("2\n"))
	p := Prompt{
		Key:     "TZ",
		Label:   "Timezone",
		Options: []string{"UTC", "America/New_York", "Europe/London"},
	}

	value, err := promptSelect(reader, p)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if value != "America/New_York" {
		t.Fatalf("expected America/New_York, got %s", value)
	}
}

func TestPromptSelect_Default(t *testing.T) {
	reader := bufio.NewReader(strings.NewReader("\n"))
	p := Prompt{
		Key:     "TZ",
		Label:   "Timezone",
		Options: []string{"UTC", "America/New_York"},
		Default: "UTC",
	}

	value, err := promptSelect(reader, p)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if value != "UTC" {
		t.Fatalf("expected UTC, got %s", value)
	}
}

func TestPromptSelect_InvalidThenValid(t *testing.T) {
	reader := bufio.NewReader(strings.NewReader("abc\n99\n1\n"))
	p := Prompt{
		Key:     "TZ",
		Label:   "Timezone",
		Options: []string{"UTC", "America/New_York"},
	}

	value, err := promptSelect(reader, p)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if value != "UTC" {
		t.Fatalf("expected UTC, got %s", value)
	}
}
