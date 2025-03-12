package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
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

// listGitFiles lists all files tracked by git in the specified directory
func listGitFiles(dir string) ([]string, error) {
	// Get absolute path
	absPath, err := filepath.Abs(dir)
	if err != nil {
		return nil, fmt.Errorf("error resolving path: %v", err)
	}

	// Change to the directory
	currentDir, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("error getting current directory: %v", err)
	}
	defer os.Chdir(currentDir) // Ensure we change back to original directory

	err = os.Chdir(absPath)
	if err != nil {
		return nil, fmt.Errorf("error changing to directory %s: %v", absPath, err)
	}

	// Run git ls-files to get all tracked files
	cmd := exec.Command("git", "ls-files")
	var out bytes.Buffer
	cmd.Stdout = &out
	err = cmd.Run()
	if err != nil {
		return nil, fmt.Errorf("error running git ls-files: %v", err)
	}

	// Split output by newlines to get file list
	var files []string
	scanner := bufio.NewScanner(&out)
	for scanner.Scan() {
		file := scanner.Text()
		if file != "" {
			files = append(files, file)
		}
	}

	return files, nil
}

// fileContainsPattern checks if a file contains the specified pattern
func fileContainsPattern(filePath string, pattern string) bool {
	if pattern == "" {
		return true // If no pattern specified, all files match
	}

	content, err := os.ReadFile(filePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading file %s: %v\n", filePath, err)
		return false
	}

	return strings.Contains(string(content), pattern)
}

// shouldExclude checks if a file should be excluded based on the exclude patterns
func shouldExclude(file string, excludePatterns []string) bool {
	for _, pattern := range excludePatterns {
		// Check if file starts with the pattern (directory exclusion)
		if strings.HasPrefix(file, pattern) {
			return true
		}
		
		// Check if file ends with the pattern (extension exclusion)
		if strings.HasSuffix(file, pattern) {
			return true
		}
		
		// Check if pattern is exactly the file
		if file == pattern {
			return true
		}
	}
	return false
}

func printFileContents(dir string, files []string, excludePatterns []string, grepPattern string) {
	for _, file := range files {
		// Skip excluded files
		if shouldExclude(file, excludePatterns) {
			continue
		}
		
		// Construct full path
		fullPath := filepath.Join(dir, file)
		
		// Skip files that don't contain the pattern
		if !fileContainsPattern(fullPath, grepPattern) {
			continue
		}
		
		// Print file name
		fmt.Printf("\n=== %s ===\n\n", file)
		
		content, err := os.ReadFile(fullPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading file %s: %v\n", fullPath, err)
			continue
		}
		
		fmt.Println(string(content))
	}
}

func printFileList(files []string, excludePatterns []string, grepPattern string) {
	for _, file := range files {
		// Skip excluded files
		if shouldExclude(file, excludePatterns) {
			continue
		}
		
		// Construct full path for grep check
		fullPath := filepath.Join(".", file)
		
		// Skip files that don't contain the pattern
		if !fileContainsPattern(fullPath, grepPattern) {
			continue
		}
		
		fmt.Println(file)
	}
}

func main() {
	// Define command line flags
	dirFlag := flag.String("dir", ".", "Directory to operate in")
	dryFlag := flag.Bool("dry", false, "Only list files without showing contents")
	excludeFlag := flag.String("exclude", "", "Comma-separated list of patterns to exclude (e.g. 'assets/,.png,.bin')")
	grepFlag := flag.String("grep", "", "Only include files containing this pattern")

	// Custom usage function to show both commands and flags
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: ep [flags]\n\n")
		fmt.Fprintf(os.Stderr, "Flags:\n")
		flag.PrintDefaults()
	}

	// Parse flags
	flag.Parse()

	// Process exclude patterns
	var excludePatterns []string
	if *excludeFlag != "" {
		excludePatterns = strings.Split(*excludeFlag, ",")
	}

	// Check if directory is in a git repository
	if !isGitRepo(*dirFlag) {
		fmt.Fprintf(os.Stderr, "Error: %s is not in a git repository\n", *dirFlag)
		os.Exit(1)
	}

	// Get list of git files
	files, err := listGitFiles(*dirFlag)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	if *dryFlag {
		// Only print the list of files
		printFileList(files, excludePatterns, *grepFlag)
	} else {
		// Print each file's name and contents
		printFileContents(*dirFlag, files, excludePatterns, *grepFlag)
	}
}
