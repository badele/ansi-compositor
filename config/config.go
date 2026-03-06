// Package config defines the YAML configuration schema for ansi-compositor.
package config

import (
	"fmt"
	"time"
)

// Config represents the complete YAML configuration file.
type Config struct {
	// Term defines the workspace (main VirtualTerminal)
	Term Term `yaml:"term"`

	// Defaults are applied to all layers unless overridden
	Defaults *LayerDefaults `yaml:"defaults,omitempty"`

	// Layers define content to be composited onto the workspace
	Layers []Layer `yaml:"layers"`

	// Sauce defines optional SAUCE metadata to export
	Sauce *SauceConfig `yaml:"sauce,omitempty"`

	// Output configuration
	Output Output `yaml:"output,omitempty"`
}

// Term defines the main workspace properties.
type Term struct {
	// Width of the workspace in columns
	Width int `yaml:"width"`

	// Height of the workspace in rows
	Height int `yaml:"height"`

	// Fill character and style for the background (optional)
	// Format: "char" or "char:SGR" where SGR is neotex style (e.g., "BK" for black bg)
	// Examples: " ", " :Bb" (space with blue background), "░:Fk,Bb"
	Fill string `yaml:"fill,omitempty"`

	// Encoding for the workspace: utf8, cp437, cp850, iso-8859-1
	Encoding string `yaml:"encoding,omitempty"`

	// UseVGAColors uses exact VGA hardware colors
	UseVGAColors bool `yaml:"vgaColors,omitempty"`
}

// LayerDefaults contains default values applied to all layers.
type LayerDefaults struct {
	// InputFormat: auto, ansi, neotex, plaintext
	InputFormat string `yaml:"inputFormat,omitempty"`

	// InputEncoding: utf8, cp437, cp850, iso-8859-1
	InputEncoding string `yaml:"inputEncoding,omitempty"`
}

// Layer defines a single content layer to be composited.
type Layer struct {
	// Name is an identifier for the layer (for debugging/logging)
	Name string `yaml:"name"`

	// Position on the workspace (1-indexed in config, converted to 0-indexed internally)
	X int `yaml:"x"`
	Y int `yaml:"y"`

	// Size of the layer's VirtualTerminal (optional)
	// If not specified, uses the source's natural size
	Width  int `yaml:"width,omitempty"`
	Height int `yaml:"height,omitempty"`

	// Content source (exactly one must be specified)
	// File path to load content from
	File string `yaml:"file,omitempty"`

	// Command to execute (stdout will be captured)
	// Can be a string or list of strings
	Cmd interface{} `yaml:"cmd,omitempty"`

	// Inline content (useful for small text)
	Content string `yaml:"content,omitempty"`

	// Input format: auto, ansi, neotex, plaintext (overrides defaults)
	InputFormat string `yaml:"inputFormat,omitempty"`

	// Input encoding: utf8, cp437, cp850, iso-8859-1 (overrides defaults)
	InputEncoding string `yaml:"inputEncoding,omitempty"`

	// AlignH: horizontal alignment - left, center, right (default: left)
	AlignH string `yaml:"alignH,omitempty"`

	// AlignV: vertical alignment - top, middle, bottom (default: top)
	AlignV string `yaml:"alignV,omitempty"`

	// Crop the source content before pasting (optional)
	// Format: "x,y:x1,y1" (1-indexed, same as splitans)
	Crop string `yaml:"crop,omitempty"`

	// Enabled allows disabling a layer without removing it
	Enabled *bool `yaml:"enabled,omitempty"`
}

// Output defines how to export the final composition.
type Output struct {
	// Format: ansi, neotex, plaintext
	Format string `yaml:"format,omitempty"`

	// Encoding: utf8, cp437, cp850, iso-8859-1
	Encoding string `yaml:"encoding,omitempty"`

	// Inline outputs everything on a single line
	Inline bool `yaml:"inline,omitempty"`

	// KeepTrailingLines preserves trailing empty lines in ansi/neotex output
	KeepTrailingLines bool `yaml:"keepTrailingLines,omitempty"`

	// File to write output (stdout if not specified)
	File string `yaml:"file,omitempty"`
}

// SauceConfig describes optional SAUCE metadata for export.
// If the block is present, SAUCE export is enabled unless explicitly disabled via enabled:false.
type SauceConfig struct {
	Enabled   *bool  `yaml:"enabled,omitempty"`
	Title     string `yaml:"title,omitempty"`
	Author    string `yaml:"author,omitempty"`
	Group     string `yaml:"group,omitempty"`
	Date      string `yaml:"date,omitempty"` // YYYYMMDD or YYYY-MM-DD
	Font      string `yaml:"font,omitempty"` // TInfoS
	ICEColors bool   `yaml:"iceColors,omitempty"`
	Comments  *uint8 `yaml:"comments,omitempty"`
	DataType  *uint8 `yaml:"dataType,omitempty"`
	FileType  *uint8 `yaml:"fileType,omitempty"`
}

// GetCmd returns the command as a string slice.
func (l *Layer) GetCmd() []string {
	if l.Cmd == nil {
		return nil
	}

	switch v := l.Cmd.(type) {
	case string:
		return []string{"sh", "-c", v}
	case []interface{}:
		result := make([]string, len(v))
		for i, item := range v {
			if s, ok := item.(string); ok {
				result[i] = s
			}
		}
		return result
	}
	return nil
}

// IsEnabled returns whether the layer is enabled (default true).
func (l *Layer) IsEnabled() bool {
	if l.Enabled == nil {
		return true
	}
	return *l.Enabled
}

// GetInputFormat returns the input format with fallback to default.
func (l *Layer) GetInputFormat(defaults *LayerDefaults) string {
	if l.InputFormat != "" {
		return l.InputFormat
	}
	if defaults != nil && defaults.InputFormat != "" {
		return defaults.InputFormat
	}
	return "auto"
}

// GetInputEncoding returns the input encoding with fallback to default.
func (l *Layer) GetInputEncoding(defaults *LayerDefaults) string {
	if l.InputEncoding != "" {
		return l.InputEncoding
	}
	if defaults != nil && defaults.InputEncoding != "" {
		return defaults.InputEncoding
	}
	return "utf8"
}

// Validate checks a sauce configuration when enabled or unspecified (default enabled).
func (s *SauceConfig) Validate() error {
	if s.Enabled != nil && !*s.Enabled {
		return nil
	}

	if len(s.Title) > 35 {
		return fmt.Errorf("sauce.title exceeds 35 characters")
	}
	if len(s.Author) > 20 {
		return fmt.Errorf("sauce.author exceeds 20 characters")
	}
	if len(s.Group) > 20 {
		return fmt.Errorf("sauce.group exceeds 20 characters")
	}
	if len(s.Font) > 22 {
		return fmt.Errorf("sauce.font exceeds 22 characters")
	}

	if s.Date != "" {
		if _, err := parseSauceDate(s.Date); err != nil {
			return fmt.Errorf("sauce.date invalid: %w", err)
		}
	}

	return nil
}

// parseSauceDate accepts YYYYMMDD or YYYY-MM-DD and returns time.Time.
func parseSauceDate(value string) (time.Time, error) {
	layouts := []string{"20060102", "2006-01-02"}
	for _, layout := range layouts {
		if t, err := time.Parse(layout, value); err == nil {
			return t, nil
		}
	}
	return time.Time{}, fmt.Errorf("expected YYYYMMDD or YYYY-MM-DD")
}
