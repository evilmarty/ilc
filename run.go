package main

import (
	"fmt"
	"strings"

	"github.com/erikgeiser/promptkit/selection"
	"github.com/erikgeiser/promptkit/textinput"
	// "os/exec"
)

const (
	minChoiceFiltering = 5
)

func run(config Config, args []string) error {
	argset := ParseArgSet(args)
	selectedCommands, err := selectCommands(config, argset)
	if err != nil {
		return fmt.Errorf("failed to select command: %v", err)
	}
	values, err := collectValues(selectedCommands.Inputs(), argset)
	if err != nil {
		return fmt.Errorf("failed to collect inputs: %v", err)
	}
	return selectedCommands.Run(values)
}

func selectCommands(config Config, argset ArgSet) (CommandChain, error) {
	var cursor ConfigCommand
	rootCommand := ConfigCommand{
		Name:        "",
		Description: config.Description,
		Run:         config.Run,
		Shell:       config.Shell,
		Env:         config.Env,
		Pure:        config.Pure,
		Inputs:      config.Inputs,
		Commands:    config.Commands,
	}
	preselection := make(CommandChain, 1, len(argset.Commands)+1)
	preselection = append(preselection, rootCommand)

	for _, command := range argset.Commands {
		cursor = preselection[len(preselection)-1]
		if cursor.Run != "" {
			return preselection, fmt.Errorf("subcommands are unavailable")
		}
		if next := cursor.Commands.Get(command); next != nil {
			preselection = append(preselection, *next)
		} else {
			return preselection, fmt.Errorf("invalid subcommand: %s", command)
		}
	}
	// Now we ask to select any remaining commands
	for cursor = preselection[len(preselection)-1]; cursor.Run == ""; cursor = preselection[len(preselection)-1] {
		if subcommand, err := selectCommand(cursor); err != nil {
			return preselection, err
		} else {
			preselection = append(preselection, subcommand)
		}
	}
	return preselection, nil
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

func collectValues(inputs []ConfigInput, argset ArgSet) (map[string]any, error) {
	inputsLength := len(inputs)
	paramsLength := len(argset.Params)
	values := make(map[string]any, inputsLength)
	inputsPending := make([]ConfigInput, 0, inputsLength)
	usedParams := make([]string, paramsLength)
	for _, input := range inputs {
		for _, usedParam := range usedParams {
			if usedParam == input.Name {
				return values, fmt.Errorf("conflict with similarly named inputs: %s", input.Name)
			}
		}
		if val, ok := argset.Params[input.Name]; ok {
			values[input.Name] = val
			usedParams = append(usedParams, input.Name)
		} else {
			inputsPending = append(inputsPending, input)
		}
	}
	if unusedParams := DiffStrings(argset.ParamNames(), usedParams); len(unusedParams) > 0 {
		return values, fmt.Errorf("unknown inputs: %s", strings.Join(unusedParams, ", "))
	}
	moreValues, err := askInputs(inputsPending)
	if err != nil {
		return values, err
	}
	for name, value := range moreValues {
		values[name] = value
	}
	return values, nil
}

func askInputs(inputs []ConfigInput) (map[string]string, error) {
	values := make(map[string]string, len(inputs))
	for _, input := range inputs {
		if val, err := askInput(input); err != nil {
			return values, err
		} else {
			values[input.Name] = val
		}
	}
	return values, nil
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
