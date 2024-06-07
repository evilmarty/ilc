package main

import (
	"flag"
	"fmt"
	"os"
)

var (
	Version   = "No version provided"
	BuildDate = "Unknown build date"
)

func main() {
	fs := flag.NewFlagSet("ILC", flag.ExitOnError)
	fs.Usage = func() {
		fmt.Fprintf(fs.Output(), "Usage of ILC:\n")
		fs.PrintDefaults()
		os.Exit(0)
	}
	fs.BoolFunc("version", "Displays the version", func(_ string) error {
		fmt.Printf("ILC - %s\nVersion: %s\n", BuildDate, Version)
		os.Exit(0)
		return nil
	})
	debug := fs.Bool("debug", false, "Print debug information")
	fs.Parse(os.Args[1:])
	args := fs.Args()
	if len(args) == 0 {
		fmt.Fprintf(fs.Output(), "configuration file not provided\n")
		os.Exit(2)
	}

	config, err := LoadConfig(args[0])
	if err != nil {
		fmt.Fprintf(fs.Output(), "error loading configuration: %v\n", err)
		os.Exit(2)
	}

	runner := NewRunner(config, args[1:])
	runner.Debug = *debug
	err = runner.Run()
	if err != nil {
		fmt.Printf("%v\n", err)
		os.Exit(1)
	}
}
