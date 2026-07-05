package ghsummary

import (
	"fmt"
	"html"
	"os"
	"strings"
	"time"
)

func GenerateSVGFile(text, outputPath string) error {
	svgContent, err := GenerateSVG(text, outputPath)
	if err != nil {
		return err
	}
	file, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.WriteString(svgContent)
	return err
}

func GenerateSVG(text, outputPath string) (string, error) {
	averageCharWidth := 10 // Width of a character in a monospaced font
	maxWidth := 480        // Maximum width in pixels
	maxCharsPerLine := maxWidth / averageCharWidth

	// Split text into lines
	var lines []string
	for len(text) > maxCharsPerLine {
		breakIndex := maxCharsPerLine
		// Find the last space within the maxCharsPerLine limit
		for i := maxCharsPerLine - 1; i >= 0; i-- {
			if text[i] == ' ' {
				breakIndex = i
				break
			}
		}
		lines = append(lines, text[:breakIndex])
		text = text[breakIndex:]
		// Trim leading spaces from the remaining text
		if len(text) > 0 && text[0] == ' ' {
			text = text[1:]
		}
	}
	lines = append(lines, text)

	// Generate SVG content with multiple lines
	svgText := ``
	y := 20
	for _, line := range lines {
		svgText += fmt.Sprintf(`<text x="10" y="%d" font-family="Courier" font-size="14" fill="gray">%s</text>`, y, renderMarkdownLine(line))
		y += 20 // Increment y position for the next line
	}

	// Add a generation timestamp at the bottom
	timestamp := svgTimestamp()
	svgText += fmt.Sprintf(`<text x="%d" y="%d" text-anchor="end" font-family="Courier" font-size="10" fill="gray" fill-opacity="50%%">Generated on: %s</text>`, maxWidth-10, y, html.EscapeString(timestamp))
	y += 20 // Increment y position for the timestamp

	svgContent := fmt.Sprintf(`<svg xmlns="http://www.w3.org/2000/svg" width="%d" height="%d">`, maxWidth, y)
	svgContent += svgText
	svgContent += `</svg>`
	return svgContent, nil
}

func svgTimestamp() string {
	if timestamp := os.Getenv("GHSUMMARY_SVG_TIMESTAMP"); timestamp != "" {
		return timestamp
	}
	return time.Now().Format(time.ANSIC)
}

func renderMarkdownLine(line string) string {
	var parts []string
	var current strings.Builder
	var bold, italic, code bool

	flush := func() {
		if current.Len() == 0 {
			return
		}
		text := current.String()
		current.Reset()
		if code {
			parts = append(parts, fmt.Sprintf(`<tspan font-family="monospace">%s</tspan>`, html.EscapeString(text)))
			return
		}
		attrs := []string{}
		if bold {
			attrs = append(attrs, `font-weight="bold"`)
		}
		if italic {
			attrs = append(attrs, `font-style="italic"`)
		}
		if len(attrs) == 0 {
			parts = append(parts, html.EscapeString(text))
			return
		}
		parts = append(parts, fmt.Sprintf(`<tspan %s>%s</tspan>`, strings.Join(attrs, " "), html.EscapeString(text)))
	}

	for i := 0; i < len(line); i++ {
		if code {
			if line[i] == '`' {
				flush()
				code = false
				continue
			}
			current.WriteByte(line[i])
			continue
		}

		if strings.HasPrefix(line[i:], "**") {
			flush()
			bold = !bold
			i++
			continue
		}
		if line[i] == '*' {
			flush()
			italic = !italic
			continue
		}
		if line[i] == '`' {
			flush()
			code = true
			continue
		}
		current.WriteByte(line[i])
	}

	flush()
	return strings.Join(parts, "")
}
