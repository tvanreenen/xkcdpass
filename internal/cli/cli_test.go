package cli

import (
	"bytes"
	"testing"
)

func TestParseDefaults(t *testing.T) {
	var stderr bytes.Buffer

	config, showVersion, err := Parse(nil, &stderr)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	if showVersion {
		t.Fatal("showVersion = true, want false")
	}

	if config.Words != 4 {
		t.Fatalf("default words = %d, want 4", config.Words)
	}

	if config.Separator != "" {
		t.Fatalf("default separator = %q, want %q", config.Separator, "")
	}
}

func TestParseVersionFlag(t *testing.T) {
	var stderr bytes.Buffer

	_, showVersion, err := Parse([]string{"--version"}, &stderr)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	if !showVersion {
		t.Fatal("showVersion = false, want true")
	}
}
