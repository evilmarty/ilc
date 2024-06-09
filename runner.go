package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/erikgeiser/promptkit/selection"
	"github.com/erikgeiser/promptkit/textinput"
	// "os/exec"
)

const (
	minChoiceFiltering = 5
)

type Runner struct {
	Config     Config
	Args       []string
	Env        []string
	Stdin      *os.File
	Stdout     *os.File
	Stderr     *os.File
	Entrypoint []string
}

func (r *Runner) printUsage(cs CommandSet) {
	entrypoint := append([]string{}, r.Entrypoint...)
	entrypoint = append(entrypoint, cs.String())
	u := NewUsage(entrypoint, "ILC", cs.Description()).ImportCommandSet(cs)
	fmt.Fprint(r.Stderr, u.String())
	os.Exit(0)
}

func (r *Runner) Run() error {
	cs, err := NewCommandSet(r.Config, r.Args)
	if err != nil {
		return fmt.Errorf("failed to select command: %v", err)
	}
	values := make(map[string]any)
	cs.ParseEnv(&values, r.Env)
	if err = cs.ParseArgs(&values); err == flag.ErrHelp {
		r.printUsage(cs)
	} else if err != nil {
		return fmt.Errorf("failed parsing arguments: %v", err)
	}
	if err = cs.AskInputs(&values); err != nil {
		return fmt.Errorf("failed getting input values: %v", err)
	}
	cmd, err := cs.Cmd(values, r.Env)
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
		Env:    os.Environ(),
		Stdin:  os.Stdin,
		Stdout: os.Stdout,
		Stderr: os.Stderr,
	}
	return &r
}

func selectCommand(command ConfigCommand) (ConfigCommand, error) {
	commandsLength := len(command.Commands)

	sp := selection.New("Choose command", command.Commands)

	if commandsLength <= minChoiceFiltering {
		sp.Filter = nil
	}

	if choice, err := sp.RunPrompt(); err != nil {
		return command, err
	} else {
		return choice, nil
	}
}

func askInput(input ConfigInput) (string, error) {
	if input.Selectable() {
		return selectInput(input)
	} else {
		return getInput(input)
	}
}

func selectInput(input ConfigInput) (string, error) {
	prompt := fmt.Sprintf("Choose a %s", input.Name)
	sp := selection.New(prompt, input.Options)

	if len(input.Options) <= minChoiceFiltering {
		sp.Filter = nil
	}

	if option, err := sp.RunPrompt(); err == nil {
		return option.Value, err
	} else {
		return "", err
	}
}

func getInput(input ConfigInput) (string, error) {
	prompt := fmt.Sprintf("Please specify a %s", input.Name)
	ti := textinput.New(prompt)
	ti.InitialValue = input.DefaultValue
	ti.Validate = func(value string) error {
		if input.Valid(value) {
			return nil
		} else {
			return fmt.Errorf("invalid value")
		}
	}
	return ti.RunPrompt()
}
