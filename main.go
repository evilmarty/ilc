package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
)

var (
	Version     = "No version provided"
	BuildDate   = "Unknown build date"
	mainFlagSet = flag.NewFlagSet("ILC", flag.ExitOnError)
	logger      = log.New(io.Discard, "DEBUG: ", log.Lshortfile)
)

func main() {
	mainFlagSet.Usage = func() {
		u := NewUsage(os.Args[0:1], "ILC", "")
		fmt.Printf("%s", u.String())
		os.Exit(0)
	}
	mainFlagSet.BoolFunc("version", "Displays the version", func(_ string) error {
		fmt.Printf("ILC - %s\nVersion: %s\n", BuildDate, Version)
		os.Exit(0)
		return nil
	})
	debug := mainFlagSet.Bool("debug", false, "Print debug information")
	nonInteractive := mainFlagSet.Bool("non-interactive", false, "Disable interactivity")
	mainFlagSet.Parse(os.Args[1:])
	args := mainFlagSet.Args()
	if len(args) == 0 {
		fmt.Fprintf(mainFlagSet.Output(), "configuration file not provided\n")
		os.Exit(2)
	}

	if *debug {
		logger.SetOutput(os.Stderr)
	}

	config, err := LoadConfig(args[0])
	if err != nil {
		fmt.Fprintf(mainFlagSet.Output(), "error loading configuration: %v\n", err)
		os.Exit(2)
	}

	runner := NewRunner(config, args[1:])
	runner.NonInteractive = *nonInteractive
	runner.Entrypoint = getEntrypoint(args[0])
	err = runner.Run()
	if err != nil {
		fmt.Printf("%v\n", err)
		os.Exit(1)
	}
}

func getEntrypoint(configPath string) []string {
	underscore := ""
	for _, item := range os.Environ() {
		if strings.HasPrefix(item, "_=") {
			underscore = strings.TrimPrefix(item, "_=")
			break
		}
	}
	// Check if config was invoked
	if underscore == configPath {
		logger.Println("Detected entrypoint to be config file")
		return []string{underscore}
	} else {
		logger.Println("Assuming direct execution of binary")
		return []string{os.Args[0], configPath}
	}
}
