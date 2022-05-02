package main

import (
	"flag"
	"fmt"
	"os"
)

const (
	defaultConfigFile = "ilc.yml"
)

var (
	Version   = "No version provided"
	BuildDate = "Unknown build date"
)

func main() {
	configFile := flag.String("f", defaultConfigFile, "Config file to load")
	version := flag.Bool("v", false, "Print version")
	flag.Parse()

	if *version {
		fmt.Printf("ILC - %s\nVersion: %s\n", BuildDate, Version)
		os.Exit(0)
	}

	m, err := newModel(*configFile)
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
