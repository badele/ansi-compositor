package config

import (
	"strings"
	"testing"
)

func boolPtr(v bool) *bool { return &v }

func TestSauceValidationLengthError(t *testing.T) {
	s := &SauceConfig{Title: strings.Repeat("a", 36)}
	if err := s.Validate(); err == nil {
		t.Fatalf("expected error for too long title")
	}
}

func TestSauceValidationDisabledSkips(t *testing.T) {
	s := &SauceConfig{Enabled: boolPtr(false), Title: strings.Repeat("a", 40)}
	if err := s.Validate(); err != nil {
		t.Fatalf("disabled sauce should skip validation, got %v", err)
	}
}

func TestParseSetsSauceEnabledDefault(t *testing.T) {
	yamlData := `
term:
  width: 80
  height: 25
layers:
  - name: base
    x: 1
    y: 1
    content: "hi"
sauce:
  title: "Test"
`

	cfg, err := Parse([]byte(yamlData), ".")
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}

	if cfg.Sauce == nil || cfg.Sauce.Enabled == nil || !*cfg.Sauce.Enabled {
		t.Fatalf("expected sauce enabled by default when block present")
	}
}
