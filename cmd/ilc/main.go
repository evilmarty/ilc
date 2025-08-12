package main

import (
	"flag"
	"fmt"
	"os"
)

var (
	version = "dev"
	debug   = false
)

func main() {
	var showVersion, showHelp bool
	flag.BoolVar(&showVersion, "version", false, "Print version information")
	flag.BoolVar(&showHelp, "help", false, "Show help")
	flag.BoolVar(&debug, "debug", false, "Enable debug logging")

	// 1. Load configuration
	cfg, err := loadConfig(os.Args[1:])
	if err != nil {
		if showHelp {
			printHelp(nil, nil)
			os.Exit(0)
		}
		if debug {
			fmt.Fprintf(os.Stderr, "Error: %+v\n", err)
		} else {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		}
		os.Exit(1)
	}

	if showVersion {
		fmt.Println(version)
		os.Exit(0)
	}

	// 2. Select command
	cmd, cmdArgs, err := selectCommand(cfg)
	if err != nil {
		if debug {
			fmt.Fprintf(os.Stderr, "Error: %+v\n", err)
		} else {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		}
		os.Exit(1)
	}

	// Check for help flag after command selection
	args := flag.Args()
	for _, arg := range args {
		if arg == "--help" || arg == "-h" {
			showHelp = true
			break
		}
	}

	if showHelp {
		printHelp(cfg, cmd)
		os.Exit(0)
	}

	// 3. Collect inputs
	inputs, err := collectInputs(cmd, cmdArgs)
	if err != nil {
		if debug {
			fmt.Fprintf(os.Stderr, "Error: %+v\n", err)
		} else {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		}
		os.Exit(1)
	}

	// 4. Execute command
	if err := executeCommand(cfg, cmd, inputs); err != nil {
		if debug {
			fmt.Fprintf(os.Stderr, "Error: %+v\n", err)
		} else {
			fmt.Fprintf(os.Stderr, "Error executing command: %v\n", err)
		}
		os.Exit(1)
	}
}













