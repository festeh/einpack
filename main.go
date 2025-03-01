package main

import (
	"fmt"
	"os"
)

func main() {
	// Check if any arguments were provided
	if len(os.Args) < 2 {
		fmt.Println("Usage: ep <command>")
		os.Exit(1)
	}

	// Get the command from arguments
	command := os.Args[1]

	// Handle different commands
	switch command {
	case "hello":
		fmt.Println("Hello, world!")
	case "version":
		fmt.Println("v0.1.0")
	default:
		fmt.Printf("Unknown command: %s\n", command)
		os.Exit(1)
	}
}
