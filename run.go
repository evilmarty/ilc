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
	values, err := cs.Values()
	if err != nil {
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

func (r *Runner) selectCommands() (CommandChain, []string, error) {
	var cursor ConfigCommand
	rootCommand := ConfigCommand{
		Name:        "",
		Description: r.Config.Description,
		Run:         r.Config.Run,
		Shell:       r.Config.Shell,
		Env:         r.Config.Env,
		Pure:        r.Config.Pure,
		Inputs:      r.Config.Inputs,
		Commands:    r.Config.Commands,
	}
	cc := CommandChain{rootCommand}
	args := r.Args

	for len(args) > 0 {
		cursor = cc[len(cc)-1]
		if cursor.Run != "" || len(cursor.Commands) == 0 {
			return cc, args, nil
		}
		next := cursor.Commands.Get(args[0])
		if next == nil {
			return cc, args, fmt.Errorf("invalid subcommand: %s", args[0])
		}
		cc = append(cc, *next)
		args = args[1:]
	}
	// Now we ask to select any remaining commands
	for cursor = cc[len(cc)-1]; cursor.Run == ""; cursor = cc[len(cc)-1] {
		if subcommand, err := selectCommand(cursor); err != nil {
			return cc, args, err
		} else {
			cc = append(cc, subcommand)
		}
	}
	return cc, args, nil
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
