package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
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

// countWordsInFile counts the number of words in a file
func countWordsInFile(filePath string) (int, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return 0, err
	}
	
	// Split content by whitespace and count non-empty words
	words := strings.Fields(string(content))
	return len(words), nil
}

// fileContainsPattern checks if a file contains the specified regex pattern
func fileContainsPattern(filePath string, pattern string) bool {
	if pattern == "" {
		return true // If no pattern specified, all files match
	}

	// Check if path exists and is not a directory
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			// Silently skip non-existent files
			return false
		}
		fmt.Fprintf(os.Stderr, "Error accessing file %s: %v\n", filePath, err)
		return false
	}
	
	// Skip directories
	if fileInfo.IsDir() {
		return false
	}

	content, err := os.ReadFile(filePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading file %s: %v\n", filePath, err)
		return false
	}

	matched, err := regexp.Match(pattern, content)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error with regex pattern '%s': %v\n", pattern, err)
		return false
	}
	return matched
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

// shouldInclude checks if a file should be included based on the include patterns
// Patterns can be grouped with semicolons (;) for AND logic - file must match at least
// one pattern from each group to be included
func shouldInclude(file string, includePattern string) bool {
	// If no include pattern specified, include all files
	if includePattern == "" {
		return true
	}
	
	// Split the pattern by semicolons to get AND groups
	groups := strings.Split(includePattern, ";")
	
	// File must match at least one pattern from each group
	for _, group := range groups {
		if group == "" {
			continue // Skip empty groups
		}
		
		// Split the group by commas to get OR patterns
		patterns := strings.Split(group, ",")
		
		matched := false
		for _, pattern := range patterns {
			if pattern == "" {
				continue // Skip empty patterns
			}
			
			// Check if file starts with the pattern (directory inclusion)
			if strings.HasPrefix(file, pattern) {
				matched = true
				break
			}
			
			// Check if file ends with the pattern (extension inclusion)
			if strings.HasSuffix(file, pattern) {
				matched = true
				break
			}
			
			// Check if pattern is exactly the file
			if file == pattern {
				matched = true
				break
			}
		}
		
		// If no pattern in this group matched, file should be excluded
		if !matched {
			return false
		}
	}
	
	// File matched at least one pattern from each group
	return true
}

func printFileContents(dir string, files []string, excludePatterns []string, includePatterns string, grepPattern string) {
	for _, file := range files {
		// Skip excluded files (exclude has priority)
		if shouldExclude(file, excludePatterns) {
			continue
		}
		
		// Skip files that don't match include patterns
		if !shouldInclude(file, includePatterns) {
			continue
		}
		
		// Construct full path
		fullPath := filepath.Join(dir, file)
		
		// Skip files that don't contain the pattern
		if !fileContainsPattern(fullPath, grepPattern) {
			continue
		}
		
		// Check if path exists and is not a directory
		fileInfo, err := os.Stat(fullPath)
		if err != nil {
			if os.IsNotExist(err) {
				// Silently skip non-existent files
				continue
			}
			fmt.Fprintf(os.Stderr, "Error accessing file %s: %v\n", fullPath, err)
			continue
		}
		
		// Skip directories
		if fileInfo.IsDir() {
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

func printFileList(files []string, excludePatterns []string, includeFlag string, grepPattern string) {
	totalWordCount := 0
	
	for _, file := range files {
		// Skip excluded files (exclude has priority)
		if shouldExclude(file, excludePatterns) {
			continue
		}
		
		// Skip files that don't match include patterns
		if !shouldInclude(file, includeFlag) {
			continue
		}
		
		// Construct full path for grep check
		fullPath := filepath.Join(".", file)
		
		// Skip files that don't contain the pattern
		if !fileContainsPattern(fullPath, grepPattern) {
			continue
		}
		
		// Count words in the file
		wordCount, err := countWordsInFile(fullPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error counting words in %s: %v\n", file, err)
			fmt.Printf("%s (word count error)\n", file)
			continue
		}
		
		// Add to total word count
		totalWordCount += wordCount
		
		// Print file with word count
		fmt.Printf("%s (%d words)\n", file, wordCount)
	}
	
	// Print total word count
	if totalWordCount > 0 {
		fmt.Printf("\nTotal: %d words\n", totalWordCount)
	}
}

func main() {
	// Define command line flags
	dirFlag := flag.String("dir", ".", "Directory to operate in")
	dryFlag := flag.Bool("dry", false, "Only list files without showing contents")
	excludeFlag := flag.String("exclude", "", "Comma-separated list of patterns to exclude (e.g. 'assets/,.png,.bin')")
	includeFlag := flag.String("include", "", "Patterns to include: comma-separated for OR logic, semicolon-separated groups for AND logic (e.g. '.go,.md' or 'src;.go,.cpp')")
	grepFlag := flag.String("grep", "", "Only include files matching this regex pattern (e.g. '(foo|bar).*')")

	// Custom usage function to show both commands and flags
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: ep [flags]\n\n")
		fmt.Fprintf(os.Stderr, "Flags:\n")
		flag.PrintDefaults()
	}

	// Parse flags
	flag.Parse()

	// Process exclude and include patterns
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
		printFileList(files, excludePatterns, *includeFlag, *grepFlag)
	} else {
		// Print each file's name and contents
		printFileContents(*dirFlag, files, excludePatterns, *includeFlag, *grepFlag)
	}
}
