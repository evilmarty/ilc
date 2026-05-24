package ilc

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

const (
	EnvVarPrefix = "ILC_INPUT_"
	EnvHistFile  = "ILC_HISTFILE"
	ReplayPrefix = "!"
)

var (
	ErrConfigFileMissing = errors.New("configuration file not provided")
	ErrMissingArguments  = errors.New("no arguments given")
	ErrInvalidCommand    = errors.New("invalid command")
	ErrInvalidReplay     = errors.New("invalid replay command")
)

var exitFunc = os.Exit

type Runner struct {
	Name           string
	NameVersion    string
	BuildDate      string
	Commit         string
	Version        string
	Env            EnvMap
	Output         io.Writer
	Stdout         io.Writer
	Stderr         io.Writer
	Stdin          io.Reader
	Args           []string
	Entrypoint     []string
	NonInteractive bool
	ValidateConfig bool
	ConfigPath     string
	Config         *Config
	HistoryFile    string
	HistoryStore   HistoryStore
	Debug          bool
	parsed         bool
}

func (r *Runner) Parsed() bool {
	return r.parsed
}

func (r *Runner) Printf(format string, a ...any) {
	output := r.Output
	if output == nil {
		output = os.Stderr
	}
	fmt.Fprintf(output, format, a...)
}

func (r *Runner) printVersion() {
	r.Printf("%s\n", r.Name)
	if r.Version != "" {
		r.Printf("Version: %s\n", r.Version)
	}
	if r.BuildDate != "" {
		r.Printf("BuildDate: %s\n", r.BuildDate)
	}
	if r.Commit != "" {
		r.Printf("Commit: %s\n", r.Commit)
	}
	exitFunc(0)
}

func (r *Runner) usage() Usage {
	fs := r.flagSet()
	output := r.Output
	if output == nil {
		output = os.Stderr
	}
	u := NewUsage(output)
	u.Title = r.Name
	u.Entrypoint = r.Entrypoint
	u.ImportFlags(fs)
	return u
}

func (r *Runner) flagSet() *flag.FlagSet {
	fs := flag.NewFlagSet(r.Name, flag.ExitOnError)
	fs.Usage = func() {
		r.usage().Print()
		exitFunc(0)
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
		config.Name = filepath.Base(r.ConfigPath)
		r.Config = &config
	}
	if underscore, found := r.Env["_"]; found && underscore == r.ConfigPath {
		r.Entrypoint = []string{r.ConfigPath}
	} else {
		r.Entrypoint = append(r.Entrypoint, r.ConfigPath)
	}
	return nil
}

func (r *Runner) run() error {
	var err error
	logger.Printf("Running with arguments: %s\n", strings.Join(r.Args, " "))
	selection := r.Config.Select(r.Args)
	inps := selection.Inputs()
	missing, err := inps.ParseEnvAndArgs(selection.Args, r.Env)
	if err != nil {
		return err
	}

	if !selection.Runnable() || len(missing) > 0 {
		if r.NonInteractive {
			if !selection.Runnable() {
				return ErrInvalidCommand
			}
			var missingNames []string
			for _, input := range missing {
				missingNames = append(missingNames, input.Name)
			}
			return fmt.Errorf("missing inputs: %s", strings.Join(missingNames, ", "))
		}
		selection, err = askCommands(selection, r.Env)
		if err != nil {
			return err
		}
		inps = selection.Inputs()
	}

	values := inps.Values()
	data := NewTemplateData(values, r.Env)
	cmd, err := selection.Cmd(data, r.Env)
	if err != nil {
		return fmt.Errorf("failed generating script: %v", err)
	}

	// Clean up temporary script file after the runner finishes
	if len(cmd.Args) > 0 {
		scriptFile := cmd.Args[len(cmd.Args)-1]
		defer os.Remove(scriptFile)
	}

	cmd.Env = append(cmd.Env, EnvMap(inps.ToEnvMap()).ToList()...)
	stdin := r.Stdin
	if stdin == nil {
		stdin = os.Stdin
	}
	stdout := r.Stdout
	if stdout == nil {
		stdout = os.Stdout
	}
	stderr := r.Stderr
	if stderr == nil {
		stderr = os.Stderr
	}

	cmd.Stdin = stdin
	cmd.Stdout = stdout
	cmd.Stderr = stderr
	if err := cmd.Run(); err != nil {
		return err
	} else {
		r.recordToHistory(selection.ToArgs())
		return nil
	}
}

func (r *Runner) Run() error {
	if !r.parsed {
		return ErrMissingArguments
	}
	if r.ValidateConfig {
		if err := r.Config.Validate(); err != nil {
			return fmt.Errorf("configuration validation failed: %w", err)
		}
		r.Printf("configuration is valid\n")
		return nil
	}
	if r.NonInteractive {
		logger.Println("Running in non-interactive mode")
	}
	if r.HistoryFile == "" {
		if histFile, found := r.Env[EnvHistFile]; found {
			r.HistoryFile = histFile
		}
	}
	var err error
	if r.isReplay() {
		err = r.replay()
	} else {
		err = r.run()
	}

	if errors.Is(err, flag.ErrHelp) {
		selection := r.Config.Select(r.Args)
		output := r.Output
		if output == nil {
			output = os.Stderr
		}
		u := NewUsage(output)
		u.Title = r.Name
		u.Entrypoint = r.Entrypoint
		u.ImportFlags(r.flagSet())
		u.ImportSelection(selection)
		u.Print()
		return nil
	}

	return err
}

func (r *Runner) replay() error {
	store := r.getHistoryStore()
	history, err := store.Load(r.HistoryFile)
	if err != nil {
		return err
	}
	r.Args[0] = strings.TrimPrefix(r.Args[0], ReplayPrefix)
	if r.Args[0] == "" {
		r.Args = r.Args[1:]
	}
	if args, found := history.Lookup(r.ConfigPath, r.Args); found {
		r.Args = args
		logger.Printf("Replaying using arguments: %s\n", strings.Join(r.Args, " "))
		return r.run()
	} else {
		return ErrInvalidReplay
	}
}

func (r *Runner) getHistoryStore() HistoryStore {
	if r.HistoryStore == nil {
		return FileHistoryStore{}
	}
	return r.HistoryStore
}

func (r *Runner) isReplay() bool {
	if len(r.Args) == 0 {
		return false
	}
	return strings.HasPrefix(r.Args[0], ReplayPrefix)
}

func (R *Runner) recordToHistory(args []string) {
	store := R.getHistoryStore()
	history, err := store.Load(R.HistoryFile)
	if err != nil {
		logger.Printf("Failed to load history for recording: %v\n", err)
		// Try to record using a fresh history object if file load fails (standard behavior)
		fresh := History{
			Path:    R.HistoryFile,
			Records: make(map[string][][]string),
		}
		history = &fresh
	}
	history.Append(R.ConfigPath, args)
	if err := store.Save(history); err != nil {
		logger.Printf("Failed to save to history: %v\n", err)
	}
}
