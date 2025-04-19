package main

import (
	"os"
	"testing"
)

func RunMainApp(t *testing.T) {
	if _, err := os.Stat("./test"); os.IsNotExist(err) {
		// Create the directory if it doesn't exist
		err := os.Mkdir("./test", 0755)
		if err != nil {
			t.Fatalf("Failed to create test directory: %v", err)
		}
	}

	{
		currentDir, err := os.Getwd()
		if err != nil {
			t.Fatalf("Failed to get current directory: %v", err)
		}
		defer os.Chdir(currentDir) // Change back to the original directory after the test

		err = os.Chdir("./test")
		if err != nil {
			t.Fatalf("Failed to change directory: %v", err)
		}

		main()
	}
}

func TestMainApp(t *testing.T) {
	// Simulate command-line arguments
	os.Args = []string{"app", "--username", "McCzarny"}
	RunMainApp(t)
}

func TestStrictMode(t *testing.T) {
	// Simulate command-line arguments
	os.Args = []string{"app", "--username", "McCzarny", "--output", "strict-summary.svg", "--max-events", "100", "--mode", "strict"}
	// Check if ./test directory exists
	RunMainApp(t)
}
