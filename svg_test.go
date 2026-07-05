package ghsummary

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGenerateSVG(t *testing.T) {
	// Create a temporary test directory
	testDir := t.TempDir()
	outputPath := filepath.Join(testDir, "test_output.svg")

	err := GenerateSVGFile("This is a test text that should be split into multiple lines if it exceeds the maximum width.", outputPath)
	if err != nil {
		t.Fatalf("GenerateSVG failed: %v", err)
	}

	// Check if the file was created
	if _, err := os.Stat(outputPath); os.IsNotExist(err) {
		t.Fatalf("Output file was not created")
	}

	// Optionally, you can read the file and verify its contents
	content, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("Failed to read output file: %v", err)
	}

	if len(content) == 0 {
		t.Fatalf("Output file is empty")
	}
}

func TestGenerateSVGSupportsBasicMarkdownFormatting(t *testing.T) {
	svg, err := GenerateSVG("Use **bold** and *italic* plus `code`", "")
	if err != nil {
		t.Fatalf("GenerateSVG failed: %v", err)
	}

	if !strings.Contains(svg, `font-weight="bold"`) {
		t.Fatalf("expected bold formatting to be rendered, got: %s", svg)
	}

	if !strings.Contains(svg, `font-style="italic"`) {
		t.Fatalf("expected italic formatting to be rendered, got: %s", svg)
	}

	if !strings.Contains(svg, `font-family="monospace"`) {
		t.Fatalf("expected inline code formatting to be rendered, got: %s", svg)
	}
}

func TestGenerateSVGMatchesCommittedExample(t *testing.T) {
	t.Setenv("GHSUMMARY_SVG_TIMESTAMP", "Sun Jul  5 14:16:36 2026")

	outputPath := filepath.Join(t.TempDir(), "summary.svg")
	input := "*McCzarny* recently focused on his `ghsummary` project, adding features " +
		"like **strict mode** and improving its action workflow. He actively developed the " +
		"`upload-image` GitHub action, implementing **Cloudinary** support, a delete image " +
		"function, and fixing test issues."

	if err := GenerateSVGFile(input, outputPath); err != nil {
		t.Fatalf("GenerateSVGFile failed: %v", err)
	}

	got, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("failed to read generated SVG: %v", err)
	}

	want, err := os.ReadFile(filepath.Join("doc", "summary.svg"))
	if err != nil {
		t.Fatalf("failed to read fixture SVG: %v", err)
	}

	if string(got) != string(want) {
		t.Fatalf("generated SVG does not match committed example")
	}
}
