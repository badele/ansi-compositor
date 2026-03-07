package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Load reads and parses a YAML configuration file.
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	return Parse(data, filepath.Dir(path))
}

// Parse parses YAML data into a Config struct.
// basePath is used to resolve relative file paths in the config.
func Parse(data []byte, basePath string) (*Config, error) {
	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse YAML: %w", err)
	}

	// Apply defaults
	if cfg.Term.Encoding == "" {
		cfg.Term.Encoding = "utf8"
	}
	if cfg.Output.Format == "" {
		cfg.Output.Format = "ansi"
	}
	if cfg.Output.Encoding == "" {
		cfg.Output.Encoding = "utf8"
	}
	if cfg.Sauce != nil && cfg.Sauce.Enabled == nil {
		val := true
		cfg.Sauce.Enabled = &val
	}

	// Resolve relative paths
	for i := range cfg.Layers {
		if cfg.Layers[i].File != "" && !filepath.IsAbs(cfg.Layers[i].File) {
			cfg.Layers[i].File = filepath.Join(basePath, cfg.Layers[i].File)
		}
	}
	if cfg.Output.File != "" && !filepath.IsAbs(cfg.Output.File) {
		cfg.Output.File = filepath.Join(basePath, cfg.Output.File)
	}

	// Validate
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	// Convert 1-indexed coordinates to 0-indexed for internal use
	for i := range cfg.Layers {
		cfg.Layers[i].X--
		cfg.Layers[i].Y--
	}

	return &cfg, nil
}

// Validate checks the configuration for errors.
func (c *Config) Validate() error {
	if c.Term.Width <= 0 {
		return fmt.Errorf("term.width must be positive")
	}
	if c.Term.Height <= 0 {
		return fmt.Errorf("term.height must be positive")
	}

	if c.Defaults != nil && c.Defaults.BoxError != "" {
		boxError := normalizeBoxError(c.Defaults.BoxError)
		if !isValidBoxError(boxError) {
			return fmt.Errorf("defaults.boxError must be none, rectangle, or fill")
		}
	}

	for i, layer := range c.Layers {
		if err := layer.Validate(i, c.Defaults); err != nil {
			return err
		}
	}

	if c.Sauce != nil {
		if err := c.Sauce.Validate(); err != nil {
			return err
		}
	}

	return nil
}

// Validate checks a layer configuration.
func (l *Layer) Validate(index int, defaults *LayerDefaults) error {
	name := l.Name
	if name == "" {
		name = fmt.Sprintf("layer[%d]", index)
	}

	// Must have exactly one content source
	sources := 0
	if l.File != "" {
		sources++
	}
	if l.Cmd != nil {
		sources++
	}
	if l.Content != "" {
		sources++
	}

	if sources == 0 {
		return fmt.Errorf("%s: must specify file, cmd, or content", name)
	}
	if sources > 1 {
		return fmt.Errorf("%s: specify only one of file, cmd, or content", name)
	}

	// Position validation (1-indexed)
	if l.X < 1 {
		return fmt.Errorf("%s: x must be >= 1 (1-indexed)", name)
	}
	if l.Y < 1 {
		return fmt.Errorf("%s: y must be >= 1 (1-indexed)", name)
	}

	// Size validation (if specified)
	if l.Width < 0 {
		return fmt.Errorf("%s: width must be non-negative", name)
	}
	if l.Height < 0 {
		return fmt.Errorf("%s: height must be non-negative", name)
	}

	// Alignment validation
	if l.AlignH != "" && l.AlignH != "left" && l.AlignH != "center" && l.AlignH != "right" {
		return fmt.Errorf("%s: alignH must be left, center, or right", name)
	}
	if l.AlignV != "" && l.AlignV != "top" && l.AlignV != "middle" && l.AlignV != "bottom" {
		return fmt.Errorf("%s: alignV must be top, middle, or bottom", name)
	}

	if l.BoxError != "" {
		boxError := normalizeBoxError(l.BoxError)
		if !isValidBoxError(boxError) {
			return fmt.Errorf("%s: boxError must be none, rectangle, or fill", name)
		}
	}
	if l.Cmd != nil && l.GetBoxError(defaults) != "none" {
		if l.Width > 0 && l.Width < 3 {
			return fmt.Errorf("%s: boxError requires width >= 3", name)
		}
		if l.Height > 0 && l.Height < 3 {
			return fmt.Errorf("%s: boxError requires height >= 3", name)
		}
	}

	return nil
}
