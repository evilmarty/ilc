package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
)

var (
	ErrConfigFileMissing = errors.New("configuration file not provided")
	ErrMissingArguments  = errors.New("no arguments given")
)

type Runner struct {
	Name           string
	Version        string
	BuildDate      string
	Commit         string
	Env            EnvMap
	Output         *os.File
	Args           []string
	Entrypoint     []string
	NonInteractive bool
	ValidateConfig bool
	ConfigPath     string
	Config         *Config
	Debug          bool
	parsed         bool
}

func (r *Runner) Parsed() bool {
	return r.parsed
}

func (r *Runner) Printf(format string, a ...any) {
	fmt.Fprintf(r.Output, format, a...)
}

func (r *Runner) printVersion() {
	if r.Name != "" {
		r.Printf("%s", r.Name)
		if r.BuildDate != "" {
			r.Printf(" - %s", r.BuildDate)
		}
		r.Printf("\n")
	}
	if r.Version != "" {
		r.Printf("Version: %s\n", r.Version)
	}
	if r.Commit != "" {
		r.Printf("Commit: %s\n", r.Commit)
	}
	os.Exit(0)
}

func (r *Runner) flagSet() *flag.FlagSet {
	fs := flag.NewFlagSet(r.Name, flag.ExitOnError)
	fs.Usage = func() {
		u := NewUsage(os.Stderr)
		u.Title = r.Name
		u.Entrypoint = r.Entrypoint
		u.ImportFlags(fs).Print()
		os.Exit(0)
	}
	fs.BoolFunc("version", "Displays the version", func(_ string) error {
		r.printVersion()
		return nil
	})
	fs.BoolVar(&r.Debug, "debug", false, "Print debug information")
	fs.BoolVar(&r.NonInteractive, "non-interactive", false, "Disable interactivity")
	fs.BoolVar(&r.ValidateConfig, "validate", false, "Validate configuration")
	return fs
}

func (r *Runner) Parse(args []string) error {
	r.parsed = true
	r.Entrypoint = args[0:1]
	r.Args = args[1:]
	fs := r.flagSet()
	if err := fs.Parse(r.Args); err != nil {
		return err
	}
	if r.Debug {
		logger.SetOutput(os.Stderr)
	}
	args = fs.Args()
	if len(args) == 0 {
		return ErrConfigFileMissing
	}
	r.ConfigPath = args[0]
	r.Args = args[1:]
	if config, err := LoadConfig(r.ConfigPath); err != nil {
		return err
	} else {
		r.Config = &config
	}
	if underscore, found := r.Env["_"]; found && underscore == r.ConfigPath {
		r.Entrypoint = []string{r.ConfigPath}
	} else {
		r.Entrypoint = append(r.Entrypoint, r.ConfigPath)
	}
	return nil
}

func (r *Runner) printUsage(cs CommandSet) {
	fs := r.flagSet()
	u := NewUsage(r.Output).ImportCommandSet(cs).ImportFlags(fs)
	u.Entrypoint = append([]string{}, r.Entrypoint...)
	if s := cs.String(); s != "" {
		u.Entrypoint = append(u.Entrypoint, s)
	}
	u.Description = cs.Description()
	u.Print()
	os.Exit(0)
}

func (r *Runner) run() error {
	if r.NonInteractive {
		logger.Println("Running in non-interactive mode")
	}
	cs, err := NewCommandSet(*r.Config, r.Args)
	if err == flag.ErrHelp {
		r.printUsage(cs)
	}
	if err != nil {
		return fmt.Errorf("failed to select command: %v", err)
	}
	if !r.NonInteractive {
		cs, err = cs.AskCommands()
		if err != nil {
			return fmt.Errorf("failed to select command: %v", err)
		}
	}
	if !cs.Runnable() {
		return fmt.Errorf("no command specified")
	}
	values := make(map[string]any)
	cs.ParseEnv(&values, r.Env)
	if err = cs.ParseArgs(&values); err == flag.ErrHelp {
		r.printUsage(cs)
	} else if err != nil {
		return fmt.Errorf("failed parsing arguments: %v", err)
	}
	if !r.NonInteractive {
		if err = cs.AskInputs(&values); err != nil {
			return fmt.Errorf("failed getting input values: %v", err)
		}
	}
	if err = cs.Validate(values); err != nil {
		return err
	}
	data := NewTemplateData(values, r.Env)
	cmd, err := cs.Cmd(data, r.Env)
	if err != nil {
		return fmt.Errorf("failed generating script: %v", err)
	}
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func (r *Runner) Run() error {
	if !r.parsed {
		return ErrMissingArguments
	}
	if r.ValidateConfig {
		r.Printf("configuration is valid\n")
		return nil
	}
	return r.run()
}
