package compositor

import (
	"fmt"
	"strings"

	"github.com/badele/ansi-compositor/config"
	"github.com/badele/splitans/pkg/splitans"
)

func (c *Compositor) renderErrorBox(layer *config.Layer, boxError string) error {
	if boxError == "" || boxError == "none" {
		return nil
	}

	width, height := c.resolveLayerDimensions(layer)
	if width < 3 || height < 3 {
		return fmt.Errorf("boxerror requires width >= 3 and height >= 3 (got %dx%d)", width, height)
	}

	label := layer.Name
	if label == "" {
		label = "layer"
	}

	var lines []string
	switch boxError {
	case "rectangle":
		lines = drawBoxRectangle(width, height, label)
	case "fill":
		pattern := layer.GetBoxErrorPattern(c.config.Defaults)
		if pattern == "" {
			pattern = "#"
		}
		lines = drawBoxFill(width, height, label, pattern)
	default:
		return fmt.Errorf("unknown boxerror: %s", boxError)
	}
	content := strings.Join(lines, "\n")

	boxVT := splitans.NewVirtualTerminal(width, height, c.config.Term.Encoding, c.config.Term.UseVGAColors)
	tok := splitans.NewANSITokenizer([]byte(content))
	if err := boxVT.ApplyTokens(tok.Tokenize()); err != nil {
		return fmt.Errorf("apply box tokens: %w", err)
	}

	if err := c.workspace.Paste(boxVT, layer.X, layer.Y); err != nil {
		return fmt.Errorf("paste error box: %w", err)
	}

	return nil
}

func (c *Compositor) resolveLayerDimensions(layer *config.Layer) (int, int) {
	width := layer.Width
	height := layer.Height
	if width <= 0 {
		width = c.config.Term.Width - layer.X
	}
	if height <= 0 {
		height = c.config.Term.Height - layer.Y
	}
	return width, height
}

func drawBoxRectangle(width, height int, label string) []string {
	grid := newBoxGrid(width, height, ' ')
	drawBoxBorder(grid)
	drawBoxErrorText(grid, label)
	return boxGridToLines(grid)
}

func drawBoxFill(width, height int, label string, pattern string) []string {
	grid := newBoxGrid(width, height, ' ')
	fillPattern(grid, pattern, 1, 1, 1, 1)
	drawBoxBorder(grid)
	drawBoxErrorText(grid, label)
	return boxGridToLines(grid)
}

func newBoxGrid(width, height int, fill rune) [][]rune {
	grid := make([][]rune, height)
	for y := 0; y < height; y++ {
		row := make([]rune, width)
		for x := 0; x < width; x++ {
			row[x] = fill
		}
		grid[y] = row
	}
	return grid
}

func fillPattern(grid [][]rune, pattern string, padLeft, padTop, padRight, padBottom int) {
	lines := parsePatternLines(pattern)
	if len(lines) == 0 {
		return
	}
	maxWidth := 0
	for _, line := range lines {
		if len(line) > maxWidth {
			maxWidth = len(line)
		}
	}
	if maxWidth == 0 {
		return
	}

	width := len(grid[0])
	height := len(grid)
	startX := padLeft
	startY := padTop
	endX := width - padRight
	endY := height - padBottom
	if startX < 0 {
		startX = 0
	}
	if startY < 0 {
		startY = 0
	}
	if endX > width {
		endX = width
	}
	if endY > height {
		endY = height
	}
	if startX >= endX || startY >= endY {
		return
	}

	patternHeight := len(lines)
	for y := startY; y < endY; y++ {
		row := grid[y]
		line := lines[(y-startY)%patternHeight]
		for x := startX; x < endX; x++ {
			idx := (x - startX) % maxWidth
			if idx < len(line) {
				row[x] = line[idx]
			} else {
				row[x] = ' '
			}
		}
	}
}

func parsePatternLines(pattern string) [][]rune {
	if pattern == "" {
		return nil
	}
	pattern = strings.ReplaceAll(pattern, "\r\n", "\n")
	parts := strings.Split(pattern, "\n")
	if len(parts) > 0 && parts[len(parts)-1] == "" {
		parts = parts[:len(parts)-1]
	}
	if len(parts) == 0 {
		return nil
	}

	lines := make([][]rune, 0, len(parts))
	for _, part := range parts {
		lines = append(lines, []rune(part))
	}
	return lines
}

func drawBoxBorder(grid [][]rune) {
	if len(grid) == 0 || len(grid[0]) == 0 {
		return
	}
	width := len(grid[0])
	height := len(grid)
	if width == 1 {
		grid[0][0] = '+'
		return
	}

	grid[0][0] = '+'
	grid[0][width-1] = '+'
	grid[height-1][0] = '+'
	grid[height-1][width-1] = '+'
	for x := 1; x < width-1; x++ {
		grid[0][x] = '-'
		grid[height-1][x] = '-'
	}
	for y := 1; y < height-1; y++ {
		grid[y][0] = '|'
		grid[y][width-1] = '|'
	}
}

func drawBoxErrorText(grid [][]rune, label string) {
	if len(grid) == 0 || len(grid[0]) == 0 {
		return
	}
	width := len(grid[0])
	innerWidth := width - 2
	if innerWidth <= 0 {
		return
	}
	innerHeight := len(grid) - 2
	if innerHeight <= 0 {
		return
	}

	lines := buildErrorLines(label, innerWidth)
	if len(lines) == 0 {
		return
	}

	renderLines := lines
	if innerHeight < len(lines) {
		switch innerHeight {
		case 1:
			renderLines = [][]rune{lines[1]}
		case 2:
			renderLines = [][]rune{lines[1], lines[2]}
		case 3:
			renderLines = [][]rune{lines[0], lines[1], lines[2]}
		default:
			renderLines = lines
		}
	}

	blockHeight := len(renderLines)
	rowStart := 1
	if innerHeight >= blockHeight {
		rowStart = 1 + (innerHeight-blockHeight)/2
	} else {
		rowStart = len(grid) / 2
	}
	colStart := 1 + (innerWidth-len(renderLines[0]))/2
	if colStart < 1 {
		colStart = 1
	}

	for rowIdx := 0; rowIdx < blockHeight; rowIdx++ {
		targetRow := rowStart + rowIdx
		if targetRow < 0 || targetRow >= len(grid) {
			continue
		}
		line := renderLines[rowIdx]
		for i, r := range line {
			grid[targetRow][colStart+i] = r
		}
	}
}

func buildErrorLines(label string, innerWidth int) [][]rune {
	if innerWidth <= 0 {
		return nil
	}

	labelRunes := []rune(label)
	maxLabel := innerWidth - 4
	if maxLabel < 0 {
		maxLabel = 0
	}
	if len(labelRunes) > maxLabel {
		labelRunes = labelRunes[:maxLabel]
	}

	lineWidth := len(labelRunes) + 4
	if lineWidth > innerWidth {
		lineWidth = innerWidth
	}
	if lineWidth <= 0 {
		return nil
	}

	blank := make([]rune, lineWidth)
	for i := range blank {
		blank[i] = ' '
	}
	line1 := append([]rune(nil), blank...)
	startLabel := 2
	if startLabel >= lineWidth {
		startLabel = 0
	}
	for i, r := range labelRunes {
		if startLabel+i >= lineWidth {
			break
		}
		line1[startLabel+i] = r
	}

	line2 := append([]rune(nil), blank...)
	errorRunes := []rune("error")
	if len(errorRunes) > lineWidth {
		errorRunes = errorRunes[:lineWidth]
	}
	startError := (lineWidth - len(errorRunes)) / 2
	for i, r := range errorRunes {
		if startError+i >= lineWidth {
			break
		}
		line2[startError+i] = r
	}

	return [][]rune{blank, line1, line2, blank}
}

func boxGridToLines(grid [][]rune) []string {
	lines := make([]string, len(grid))
	for i, row := range grid {
		lines[i] = string(row)
	}
	return lines
}
