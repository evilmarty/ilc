package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
)

var (
	Version     = "No version provided"
	BuildDate   = "Unknown build date"
	mainFlagSet = flag.NewFlagSet("ILC", flag.ExitOnError)
)

func main() {
	mainFlagSet.Usage = func() {
		fmt.Fprintf(mainFlagSet.Output(), "Usage of ILC:\n")
		mainFlagSet.PrintDefaults()
		os.Exit(0)
	}
	mainFlagSet.BoolFunc("version", "Displays the version", func(_ string) error {
		fmt.Printf("ILC - %s\nVersion: %s\n", BuildDate, Version)
		os.Exit(0)
		return nil
	})
	debug := mainFlagSet.Bool("debug", false, "Print debug information")
	mainFlagSet.Parse(os.Args[1:])
	args := mainFlagSet.Args()
	if len(args) == 0 {
		fmt.Fprintf(mainFlagSet.Output(), "configuration file not provided\n")
		os.Exit(2)
	}

	config, err := LoadConfig(args[0])
	if err != nil {
		fmt.Fprintf(mainFlagSet.Output(), "error loading configuration: %v\n", err)
		os.Exit(2)
	}

	runner := NewRunner(config, args[1:])
	runner.Entrypoint = getEntrypoint(args[0])
	runner.Debug = *debug
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
		return []string{underscore}
	} else {
		return []string{os.Args[0], configPath}
	}
}
