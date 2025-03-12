package main

import (
	"fmt"
	"io"
	"log"
	"os"
)

var (
	Name      = "ILC"
	Version   = "No version provided"
	BuildDate = "Unknown build date"
	Commit    = ""
	logger    = log.New(io.Discard, "DEBUG: ", log.Lshortfile)
)

func main() {
	r := Runner{
		Name:      Name,
		Version:   Version,
		BuildDate: BuildDate,
		Commit:    Commit,
		Env:       NewEnvMap(os.Environ()),
		Output:    os.Stderr,
	}
	if err := r.Parse(os.Args); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		if err == ErrConfigFileMissing {
			os.Exit(2)
		} else {
			os.Exit(1)
		}
	}

	if err := r.Run(); err != nil {
		r.Printf("%v\n", err)
		os.Exit(1)
	}
}
