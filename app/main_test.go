package main

import (
	"os"
	"testing"
)

func TestMainApp(t *testing.T) {
	// Simulate command-line arguments
	os.Args = []string{"app", "--username", "McCzarny"}

	// Check if ./test directory exists
	if _, err := os.Stat("./test"); os.IsNotExist(err) {
		// Create the directory if it doesn't exist
		err := os.Mkdir("./test", 0755)
		if err != nil {
			t.Fatalf("Failed to create test directory: %v", err)
		}
	}
	// Move the current working directory to ./test
	err := os.Chdir("./test")
	if err != nil {
		t.Fatalf("Failed to change directory: %v", err)
	}

	// Run the main function
	main()
}
