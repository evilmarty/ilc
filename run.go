package main

import (
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
	Config Config
	Args   []string
	Debug  bool
	Env    []string
	Stdin  *os.File
	Stdout *os.File
	Stderr *os.File
}

func (r *Runner) Run() error {
	cs, err := NewCommandSet(r.Config, r.Args)
	if err != nil {
		return fmt.Errorf("failed to select command: %v", err)
	}
	values := make(map[string]any)
	cs.ParseEnv(&values, r.Env)
	if err = cs.ParseArgs(&values); err != nil {
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
	choices := make([]*selection.Choice, commandsLength)

	for i, subcommand := range command.Commands {
		choices[i] = &selection.Choice{String: subcommand.Name, Value: subcommand}
	}

	sp := selection.New("Choose command", choices)

	if commandsLength <= minChoiceFiltering {
		sp.Filter = nil
	}

	if choice, err := sp.RunPrompt(); err != nil {
		return command, err
	} else {
		value := choice.Value.(ConfigCommand)
		return value, nil
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
	choices := make([]*selection.Choice, 0, input.Options.Len())
	for _, option := range input.Options {
		choices = append(choices, &selection.Choice{String: option.Label, Value: option.Value})
	}
	sp := selection.New(prompt, choices)

	if len(choices) <= minChoiceFiltering {
		sp.Filter = nil
	}

	if choice, err := sp.RunPrompt(); err == nil {
		return choice.Value.(string), err
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
