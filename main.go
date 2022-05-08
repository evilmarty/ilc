package main

import (
	"flag"
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
	configFile := flag.String("f", defaultConfigFile, "Config file to load")
	showVersion := flag.Bool("version", false, "Print version")
	showHelp := flag.Bool("help", false, "Show this help screen")
	flag.Parse()

	if *showVersion {
		fmt.Printf("ILC - %s\nVersion: %s\n", BuildDate, Version)
		os.Exit(0)
	}

	m, err := newModel(*configFile)

	if *showHelp {
		printHelp(m)
		os.Exit(0)
	}

	if err != nil {
		fmt.Println("Error loading config:", err)
		os.Exit(1)
	}

	err = m.Run(flag.Args())

	if err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
}
