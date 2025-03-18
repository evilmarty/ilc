package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"strings"
)

var (
	EnvVarPrefix         = "ILC_INPUT_"
	ErrConfigFileMissing = errors.New("configuration file not provided")
	ErrMissingArguments  = errors.New("no arguments given")
	ErrInvalidCommand    = errors.New("invalid command")
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

func (r *Runner) usage() Usage {
	fs := r.flagSet()
	u := NewUsage(os.Stderr)
	u.Title = r.Name
	u.Entrypoint = r.Entrypoint
	u.ImportFlags(fs)
	return u
}

func (r *Runner) flagSet() *flag.FlagSet {
	fs := flag.NewFlagSet(r.Name, flag.ExitOnError)
	fs.Usage = func() {
		r.usage().Print()
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

func (r *Runner) getInputValuesFromEnv(inputs Inputs) map[string]string {
	values := make(map[string]string, len(inputs))
	inputEnv := r.Env.FilterPrefix(EnvVarPrefix).TrimPrefix(EnvVarPrefix)
	for _, input := range inputs {
		if value, found := inputEnv[input.EnvName()]; found {
			values[input.Name] = value
		}
	}
	return values
}

func (r *Runner) run() error {
	var err error
	if r.NonInteractive {
		logger.Println("Running in non-interactive mode")
	}
	selected, args := r.Config.Select(r.Args)
	var inputs Inputs
	var values map[string]any
	usageFunc := func() {
		u := r.usage()
		u.ImportSelectedCommands(selected).Print()
		os.Exit(0)
	}
	for {
		inputs = selected.Inputs()
		fs := inputs.FlagSet()
		fs.Init(r.Name, flag.ExitOnError)
		fs.Usage = usageFunc
		// Set the input values on the flag set to determine what inputs are outstanding
		for name, value := range r.getInputValuesFromEnv(inputs) {
			fs.Set(name, value)
		}
		if err := fs.Parse(args); err != nil {
			return err
		} else if selected.Runnable() {
			values = make(map[string]any, len(inputs))
			fs.Visit(func(f *flag.Flag) {
				if v, ok := f.Value.(flag.Getter); ok {
					values[f.Name] = v.Get()
				}
			})
			break
		} else if r.NonInteractive {
			return ErrInvalidCommand
		} else if selected, err = askCommands(selected); err != nil {
			return err
		}
	}

	var missingInputs Inputs
	for _, input := range inputs {
		if _, found := values[input.Name]; !found {
			missingInputs = append(missingInputs, input)
		}
	}

	if len(missingInputs) > 0 {
		if r.NonInteractive {
			var inputNames []string
			for _, input := range missingInputs {
				inputNames = append(inputNames, input.Name)
			}
			return fmt.Errorf("missing inputs: %s", strings.Join(inputNames, ", "))
		}
		if err := askInputs(missingInputs); err != nil {
			return err
		}
		for name, value := range missingInputs.GetAll() {
			values[name] = value
		}
	}

	data := NewTemplateData(values, r.Env)
	cmd, err := selected.Cmd(data, r.Env)
	if err != nil {
		return fmt.Errorf("failed generating script: %v", err)
	}
	cmd.Env = append(cmd.Env, inputs.ToEnvMap().Prefix(EnvVarPrefix).ToList()...)
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
