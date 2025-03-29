package utils

import (
	"fmt"
	"testing"
)

func TestSanitizeInputs(t *testing.T) {
	// Test cases for input sanitization
	tests := []struct {
		username   string
		outputFile string
		expected   bool
	}{
		{"validUser", "output.svg", true},
		{"validUser", "dir/output.svg", true},
		{"validUser", "./dir/output.svg", true},
		{"invalid/user", "output.svg", false},
		{"validUser", "../output.svg", false},
		{"validUser", "", false},
		{"userWithVeryLongNameThatExceedsThirtyNineCharacters", "output.svg", false},
		{"validUser", "/absolute/path.svg", false},
		{"validUser", "../parent/path.svg", false},
		{"validUser", "dir/../../parent/path.svg", false},
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("username=%s,outputFile=%s", test.username, test.outputFile), func(t *testing.T) {
			result := SanitizeInputs(test.username, test.outputFile)
			if result != test.expected {
				t.Errorf("SanitizeInputs(%q, %q) = %v; want %v", test.username, test.outputFile, result, test.expected)
			}
		})
	}
}
