package ghsummary

import (
	"fmt"
	"os"
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
		svgText += fmt.Sprintf(`<text x="10" y="%d" font-family="Courier" font-size="14" fill="gray">%s</text>`, y, line)
		y += 20 // Increment y position for the next line
	}
	svgContent := fmt.Sprintf(`<svg xmlns="http://www.w3.org/2000/svg" width="%d" height="%d">`, maxWidth, y)
	svgContent += svgText
	svgContent += `</svg>`
	return svgContent, nil
}
