package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

// isGitRepo checks if the provided directory is within a git repository
func isGitRepo(dir string) bool {
	// Get absolute path
	absPath, err := filepath.Abs(dir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error resolving path: %v\n", err)
		return false
	}

	// Change to the directory to check
	currentDir, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error getting current directory: %v\n", err)
		return false
	}
	defer os.Chdir(currentDir) // Ensure we change back to original directory

	err = os.Chdir(absPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error changing to directory %s: %v\n", absPath, err)
		return false
	}

	// Run git rev-parse to check if we're in a git repo
	cmd := exec.Command("git", "rev-parse", "--is-inside-work-tree")
	err = cmd.Run()
	return err == nil
}

func main() {
	// Define command line flags
	dirFlag := flag.String("dir", ".", "Directory to operate in")

	// Custom usage function to show both commands and flags
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: ep [flags] \n\n")
		fmt.Fprintf(os.Stderr, "Flags:\n")
		flag.PrintDefaults()
	}

	// Parse flags but keep the command and its args for later processing
	flag.Parse()

	// Check if directory is in a git repository
	if !isGitRepo(*dirFlag) {
		fmt.Fprintf(os.Stderr, "Error: %s is not in a git repository\n", *dirFlag)
		os.Exit(1)
	}
}
