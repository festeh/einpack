package main

import (
	"flag"
	"fmt"
	"os"
)

func main() {
	// Define command line flags
	dirFlag := flag.String("dir", ".", "Directory to operate in")
	
	// Custom usage function to show both commands and flags
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: ep [flags] <command>\n\n")
		fmt.Fprintf(os.Stderr, "Flags:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nCommands:\n")
		fmt.Fprintf(os.Stderr, "  hello     Print a greeting\n")
		fmt.Fprintf(os.Stderr, "  version   Print version information\n")
	}

	// Parse flags but keep the command and its args for later processing
	flag.Parse()
	args := flag.Args()

	// Check if any command was provided
	if len(args) < 1 {
		flag.Usage()
		os.Exit(1)
	}

	// Get the command from arguments
	command := args[0]

	// Handle different commands
	switch command {
	case "hello":
		fmt.Printf("Hello, world! (Using directory: %s)\n", *dirFlag)
	case "version":
		fmt.Println("v0.1.0")
	default:
		fmt.Printf("Unknown command: %s\n", command)
		os.Exit(1)
	}
}
