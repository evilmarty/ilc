package main

import (
	"flag"
	"fmt"
	"os"
)

const (
	defaultConfigFile = "ilc.yml"
)

func main() {
	var configFile = flag.String("f", defaultConfigFile, "Config file to load")
	flag.Parse()

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
