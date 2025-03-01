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
	}

	// Parse flags but keep the command and its args for later processing
	flag.Parse()
	args := flag.Args()

	// Check if any command was provided
	if len(args) < 1 {
		flag.Usage()
		os.Exit(1)
	}

}
