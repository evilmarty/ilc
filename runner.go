package main

import (
	"flag"
	"fmt"
	"os"
	// "os/exec"
)

type Runner struct {
	Config         Config
	Args           []string
	Env            EnvMap
	Stdin          *os.File
	Stdout         *os.File
	Stderr         *os.File
	Entrypoint     []string
	NonInteractive bool
}

func (r *Runner) printUsage(cs CommandSet) {
	u := NewUsage(mainFlagSet.Output()).ImportCommandSet(cs).ImportFlags(mainFlagSet)
	u.Entrypoint = append([]string{}, r.Entrypoint...)
	if s := cs.String(); s != "" {
		u.Entrypoint = append(u.Entrypoint, s)
	}
	u.Description = cs.Description()
	u.Print()
	os.Exit(0)
}

func (r *Runner) Run() error {
	if r.NonInteractive {
		logger.Println("Running in non-interactive mode")
	}
	cs, err := NewCommandSet(r.Config, r.Args)
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
	cmd.Stdin = r.Stdin
	cmd.Stdout = r.Stdout
	cmd.Stderr = r.Stderr
	return cmd.Run()
}

func NewRunner(config Config, args []string) *Runner {
	r := Runner{
		Config: config,
		Args:   args,
		Env:    NewEnvMap(os.Environ()),
		Stdin:  os.Stdin,
		Stdout: os.Stdout,
		Stderr: os.Stderr,
	}
	return &r
}
