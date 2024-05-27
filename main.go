package main

import (
	"fmt"
	"os"
)

const (
	defaultConfigFile = "ilc.yml"
	helpWidth         = 80
)

var (
	Version   = "No version provided"
	BuildDate = "Unknown build date"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	switch os.Args[1] {
	case "help":
		printUsage()
	case "--help":
		printUsage()
	case "--version":
		printVersion()
	default:
		if err := loadAndRun(os.Args[1], os.Args[2:]); err != nil {
			fmt.Printf("Error: %v", err)
			os.Exit(1)
		}
	}
}

func printUsage() {
	fmt.Printf(`
ILC
---

Usage:

  ilc <--version|--help|CONFIG> ...

Arguments:

  --version         Display the version information.
  --help            Display this message.

  CONFIG            The path to a ILC config file.
`)
}

func printVersion() {
	fmt.Printf("ILC - %s\nVersion: %s\n", BuildDate, Version)
}

func loadAndRun(configPath string, args []string) error {
	config, err := LoadConfig(configPath)
	if err != nil {
		return fmt.Errorf("failed to load config '%s' for the following reason: %s", configPath, err)
	}
	return run(config, args)
}
