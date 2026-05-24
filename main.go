package main

import (
	"fmt"
	"os"

	"github.com/evilmarty/ilc/internal/ilc"
)

var (
	Name      = "ILC"
	Version   = "No version provided"
	BuildDate = "Unknown build date"
	Commit    = ""
)

func main() {
	defer func() {
		if err := recover(); err != nil {
			// Ensure terminal cursor is visible and formatting is reset before re-panicking
			fmt.Fprintf(os.Stderr, "\033[?25h\033[0m\n[ilc panic recovery]\n")
			panic(err)
		}
	}()

	r := ilc.Runner{
		Name:      Name,
		Version:   Version,
		BuildDate: BuildDate,
		Commit:    Commit,
		Env:       ilc.NewEnvMap(os.Environ()),
		Output:    os.Stderr,
	}
	if err := r.Parse(os.Args); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		if err == ilc.ErrConfigFileMissing {
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

