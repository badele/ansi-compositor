// Package compositor implements the ANSI art composition engine.
package compositor

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/badele/ansi-compositor/config"
	"github.com/badele/splitans/pkg/splitans"
)

// Compositor composes multiple layers onto a workspace.
type Compositor struct {
	config    *config.Config
	workspace *splitans.VirtualTerminal
}

// New creates a new Compositor from a configuration.
func New(cfg *config.Config) *Compositor {
	return &Compositor{
		config: cfg,
	}
}

// Compose processes all layers and creates the final composition.
func (c *Compositor) Compose() error {
	// Create workspace
	c.workspace = splitans.NewVirtualTerminal(
		c.config.Term.Width,
		c.config.Term.Height,
		c.config.Term.Encoding,
		c.config.Term.UseVGAColors,
	)

	// Apply background fill if specified
	if c.config.Term.Fill != "" {
		if err := c.applyFill(); err != nil {
			return fmt.Errorf("fill error: %w", err)
		}
	}

	// Process each layer
	for _, layer := range c.config.Layers {
		if !layer.IsEnabled() {
			continue
		}

		if err := c.processLayer(&layer); err != nil {
			return fmt.Errorf("layer %q: %w", layer.Name, err)
		}
	}

	return nil
}

// applyFill fills the workspace with the specified character and style.
func (c *Compositor) applyFill() error {
	fill := c.config.Term.Fill

	// Parse fill format: "char" or "char:SGR"
	char := ' '
	var sgrStr string

	if idx := strings.Index(fill, ":"); idx != -1 {
		if idx > 0 {
			runes := []rune(fill[:idx])
			if len(runes) > 0 {
				char = runes[0]
			}
		}
		sgrStr = fill[idx+1:]
	} else if len(fill) > 0 {
		char = []rune(fill)[0]
	}

	// Create SGR from neotex codes if specified
	sgr := splitans.NewSGR()
	if sgrStr != "" {
		// Parse neotex-style SGR codes (e.g., "Bb" for blue background)
		// This is simplified - for full support we'd need to parse neotex codes
		sgr = parseNeotexSGR(sgrStr)
	}

	// Fill the workspace
	c.workspace.Fill(char, sgr)
	return nil
}

// parseNeotexSGR parses simple neotex SGR codes.
// Supports: Fk-Fw (fg 0-7), FK-FW (fg 8-15), Bk-Bw (bg 0-7), BK-BW (bg 8-15)
func parseNeotexSGR(codes string) *splitans.SGR {
	sgr := splitans.NewSGR()

	// Split by comma if multiple codes
	parts := strings.Split(codes, ",")
	for _, code := range parts {
		code = strings.TrimSpace(code)
		if len(code) < 2 {
			continue
		}

		prefix := code[0]
		colorChar := code[1]

		var colorIndex uint8
		var isBright bool

		// Lowercase = standard colors (0-7), Uppercase = bright colors (8-15)
		switch {
		case colorChar >= 'k' && colorChar <= 'w':
			colorIndex = uint8(colorChar - 'k')
		case colorChar >= 'K' && colorChar <= 'W':
			colorIndex = uint8(colorChar-'K') + 8
			isBright = true
		default:
			continue
		}

		_ = isBright // Could be used for extended handling

		switch prefix {
		case 'F': // Foreground
			sgr.FgColor.Type = splitans.ColorStandard
			sgr.FgColor.Index = colorIndex
		case 'B': // Background
			sgr.BgColor.Type = splitans.ColorStandard
			sgr.BgColor.Index = colorIndex
		}
	}

	return sgr
}

func parseNeotexLineCount(data []byte) int {
	for i := 0; i+2 < len(data); i++ {
		if data[i] != '!' || data[i+1] != 'N' || data[i+2] != 'L' {
			continue
		}
		value := 0
		j := i + 3
		for j < len(data) && data[j] >= '0' && data[j] <= '9' {
			value = value*10 + int(data[j]-'0')
			j++
		}
		if j > i+3 {
			return value
		}
	}
	return 0
}

// processLayer loads and pastes a single layer onto the workspace.
func (c *Compositor) processLayer(layer *config.Layer) error {
	// Get content data
	data, err := c.getLayerContent(layer)
	if err != nil {
		return err
	}

	// Determine format
	format := layer.GetInputFormat(c.config.Defaults)
	if format == "auto" {
		format = detectFormat(layer.File, data)
	}

	// Convert encoding if needed
	encoding := layer.GetInputEncoding(c.config.Defaults)
	if encoding != "utf8" {
		data, err = splitans.ConvertToUTF8(data, encoding)
		if err != nil {
			return fmt.Errorf("encoding conversion: %w", err)
		}
	}

	// Tokenize
	var tokens []splitans.Token
	neotexWidth := 0
	neotexHeight := 0

	switch format {
	case "neotex":
		inputWidth := c.config.Term.Width
		if layer.Width > 0 {
			inputWidth = layer.Width
		}
		parsedWidth, tok := splitans.NewNeotexTokenizer(data, inputWidth)
		tokens = tok.Tokenize()
		neotexWidth = parsedWidth
		neotexHeight = parseNeotexLineCount(data)
	case "ansi":
		tok := splitans.NewANSITokenizer(data)
		tokens = tok.Tokenize()
	case "plaintext":
		// Wrap plaintext in a simple token
		tokens = []splitans.Token{{Type: splitans.TokenText, Value: string(data)}}
	default:
		return fmt.Errorf("unknown format: %s", format)
	}

	// Determine layer VT dimensions
	vtWidth := layer.Width
	vtHeight := layer.Height
	if vtWidth <= 0 {
		vtWidth = c.config.Term.Width - layer.X
	}
	if vtHeight <= 0 {
		vtHeight = c.config.Term.Height - layer.Y
	}

	contentWidth := vtWidth
	contentHeight := vtHeight
	if format == "neotex" {
		if neotexWidth > 0 {
			contentWidth = neotexWidth
		}
		if neotexHeight > 0 {
			contentHeight = neotexHeight
		}
	}
	if contentWidth > vtWidth {
		contentWidth = vtWidth
	}
	if contentHeight > vtHeight {
		contentHeight = vtHeight
	}

	// Create content VT and apply tokens
	contentVT := splitans.NewVirtualTerminal(contentWidth, contentHeight, c.config.Term.Encoding, c.config.Term.UseVGAColors)
	if err := contentVT.ApplyTokens(tokens); err != nil {
		return fmt.Errorf("apply tokens: %w", err)
	}

	// Create layer VT and apply alignment by moving content within it (before crop)
	layerVT := splitans.NewVirtualTerminal(vtWidth, vtHeight, c.config.Term.Encoding, c.config.Term.UseVGAColors)
	offsetX, offsetY := 0, 0
	if layer.AlignH != "" || layer.AlignV != "" {
		bounds := contentVT.GetContentBounds()
		if !bounds.Empty {
			switch layer.AlignH {
			case "center":
				offsetX = (vtWidth-bounds.Width)/2 - bounds.MinX
			case "right":
				offsetX = vtWidth - bounds.Width - bounds.MinX
				// "left" or "" = no offset
			}

			switch layer.AlignV {
			case "middle":
				offsetY = (vtHeight-bounds.Height)/2 - bounds.MinY
			case "bottom":
				offsetY = vtHeight - bounds.Height - bounds.MinY
				// "top" or "" = no offset
			}
		}
	}

	if err := layerVT.Paste(contentVT, offsetX, offsetY); err != nil {
		return fmt.Errorf("paste layer content: %w", err)
	}

	// Apply crop if specified
	if layer.Crop != "" {
		cropRegion, err := splitans.ParseCropRegion(layer.Crop)
		if err != nil {
			return fmt.Errorf("invalid crop: %w", err)
		}
		layerVT = layerVT.Crop(cropRegion.X, cropRegion.Y, cropRegion.Width, cropRegion.Height)
		if layerVT == nil {
			return fmt.Errorf("crop resulted in nil VT")
		}
	}

	// Paste onto workspace
	if err := c.workspace.Paste(layerVT, layer.X, layer.Y); err != nil {
		return fmt.Errorf("paste: %w", err)
	}

	return nil
}

// getLayerContent retrieves the raw content for a layer.
func (c *Compositor) getLayerContent(layer *config.Layer) ([]byte, error) {
	// File source
	if layer.File != "" {
		data, err := os.ReadFile(layer.File)
		if err != nil {
			return nil, fmt.Errorf("read file %q: %w", layer.File, err)
		}
		return data, nil
	}

	// Command source
	if cmd := layer.GetCmd(); cmd != nil {
		var stdout bytes.Buffer
		var stderr bytes.Buffer

		command := exec.Command(cmd[0], cmd[1:]...)
		command.Stdout = &stdout
		command.Stderr = &stderr

		if err := command.Run(); err != nil {
			return nil, fmt.Errorf("command failed: %w\nstderr: %s", err, stderr.String())
		}

		return stdout.Bytes(), nil
	}

	// Inline content
	if layer.Content != "" {
		return []byte(layer.Content), nil
	}

	return nil, fmt.Errorf("no content source")
}

// detectFormat auto-detects the format based on file extension or content.
func detectFormat(filename string, data []byte) string {
	// Check extension first
	ext := strings.ToLower(filepath.Ext(filename))
	switch ext {
	case ".neo", ".neotex":
		return "neotex"
	case ".ans", ".ansi":
		return "ansi"
	case ".txt":
		return "plaintext"
	}

	// Check content for neotex marker
	if bytes.Contains(data, []byte(" | ")) {
		return "neotex"
	}

	// Check for ANSI escape sequences
	if bytes.Contains(data, []byte("\x1b[")) {
		return "ansi"
	}

	// Default to plaintext
	return "plaintext"
}

// Export returns the final composition in the configured format.
func (c *Compositor) Export() (string, error) {
	if c.workspace == nil {
		return "", fmt.Errorf("must call Compose() first")
	}

	output := c.config.Output
	sauce, err := c.buildSauce()
	if err != nil {
		return "", err
	}

	switch output.Format {
	case "ansi":
		var ansiOut string
		if output.Inline {
			ansiOut = c.workspace.ExportFlattenedANSIInline()
		} else {
			ansiOut = c.workspace.ExportFlattenedANSI()
		}

		if sauce == nil {
			return ansiOut, nil
		}

		combined := append([]byte(ansiOut), sauce.ToBytes()...)
		return string(combined), nil

	case "neotex":
		var text, sequences string
		if output.Inline {
			text, sequences = splitans.ExportToInlineNeotex(c.workspace)
		} else {
			text, sequences = splitans.ExportToNeotex(c.workspace)
		}

		if sauce == nil {
			return concatenateTextAndSequence(text, sequences, c.workspace.GetWidth(), " | "), nil
		}

		labels, err := sauceToNeotexLabels(sauce)
		if err != nil {
			return "", err
		}

		textWithSauce := fmt.Sprintf("%s\n%s", text, strings.Repeat(" ", c.config.Term.Width))
		sequencesWithSauce := fmt.Sprintf("%s\n%s", sequences, strings.Join(labels, ";"))
		return concatenateTextAndSequence(textWithSauce, sequencesWithSauce, c.config.Term.Width, " | "), nil

	case "plaintext", "text":
		if output.Inline {
			return c.workspace.ExportPlainTextInline(), nil
		}
		return c.workspace.ExportPlainText(), nil

	default:
		return "", fmt.Errorf("unknown output format: %s", output.Format)
	}
}

// GetWorkspace returns the workspace VT for direct manipulation.
func (c *Compositor) GetWorkspace() *splitans.VirtualTerminal {
	return c.workspace
}

// concatenateTextAndSequence combines text and sequence lines with a separator.
func concatenateTextAndSequence(leftText, rightText string, leftWidth int, separator string) string {
	leftLines := strings.Split(leftText, "\n")
	rightLines := strings.Split(rightText, "\n")

	result := []string{}
	numLines := len(leftLines)

	for i := 0; i < numLines; i++ {
		if i < len(leftLines) {
			leftLine := leftLines[i]
			rightLine := ""
			if i < len(rightLines) {
				rightLine = rightLines[i]
			}

			if len(leftLine) < leftWidth {
				break
			}

			result = append(result, fmt.Sprintf("%s%s%s", leftLine, separator, rightLine))
		}
	}

	return strings.Join(result, "\n")
}

// buildSauce constructs a SAUCE metadata record from configuration, or returns nil if disabled/absent.
func (c *Compositor) buildSauce() (*splitans.Sauce, error) {
	sc := c.config.Sauce
	if sc == nil || (sc.Enabled != nil && !*sc.Enabled) {
		return nil, nil
	}

	sauce := splitans.NewSauce(c.config.Term.Width, c.config.Term.Height)
	sauce.Title = sc.Title
	sauce.Author = sc.Author
	sauce.Group = sc.Group
	sauce.TInfoS = sc.Font
	sauce.SetICEColors(sc.ICEColors)

	if sc.Comments != nil {
		sauce.Comments = *sc.Comments
	}
	if sc.DataType != nil {
		sauce.DataType = *sc.DataType
	}
	if sc.FileType != nil {
		sauce.FileType = *sc.FileType
	}
	if sc.Date != "" {
		dt, err := parseSauceDate(sc.Date)
		if err != nil {
			return nil, err
		}
		sauce.Date = dt
	}

	return sauce, nil
}

func sauceToNeotexLabels(sauce *splitans.Sauce) ([]string, error) {
	if sauce == nil {
		return nil, nil
	}

	labels := []string{}

	if sauce.Title != "" {
		label, err := formatNeotexLabel("ST", sauce.Title)
		if err != nil {
			return nil, err
		}
		labels = append(labels, label)
	}
	if sauce.Author != "" {
		label, err := formatNeotexLabel("SA", sauce.Author)
		if err != nil {
			return nil, err
		}
		labels = append(labels, label)
	}
	if sauce.Group != "" {
		label, err := formatNeotexLabel("SG", sauce.Group)
		if err != nil {
			return nil, err
		}
		labels = append(labels, label)
	}
	if !sauce.Date.IsZero() {
		label, err := formatNeotexLabel("SD", sauce.Date.Format("20060102"))
		if err != nil {
			return nil, err
		}
		labels = append(labels, label)
	}
	if sauce.TInfoS != "" {
		label, err := formatNeotexLabel("SF", sauce.TInfoS)
		if err != nil {
			return nil, err
		}
		labels = append(labels, label)
	}
	if sauce.HasICEColors() {
		labels = append(labels, "!SI")
	}

	return labels, nil
}

func formatNeotexLabel(key, value string) (string, error) {
	if strings.ContainsAny(value, "<>") {
		return "", fmt.Errorf("neotex label %s contains angle brackets", key)
	}

	if strings.ContainsAny(value, " ;,:<>") {
		return fmt.Sprintf("!%s<%s>", key, value), nil
	}

	return fmt.Sprintf("!%s%s", key, value), nil
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
