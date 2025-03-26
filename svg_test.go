package ghsummary

import (
	"os"
	"path/filepath"
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
