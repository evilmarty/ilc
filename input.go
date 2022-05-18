package main

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
	"github.com/erikgeiser/promptkit/selection"
	"github.com/erikgeiser/promptkit/textinput"
)

const (
	minChoiceFiltering = 5
)

var (
	promptStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("#666"))
	inputNameStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#0cc")).Bold(true)
)

type InputValues map[string]any

func askCommand(commands ConfigCommands) (*ConfigCommand, error) {
	var choices = make([]*selection.Choice, len(commands))

	for i, command := range commands {
		choices[i] = &selection.Choice{String: command.Name, Value: command}
	}

	prompt := promptStyle.Render("Choose command")
	sp := selection.New(prompt, choices)

	if len(choices) <= minChoiceFiltering {
		sp.Filter = nil
	}

	choice, err := sp.RunPrompt()
	if err != nil {
		return nil, err
	}
	if command, ok := choice.Value.(ConfigCommand); ok {
		return &command, nil
	} else {
		return nil, fmt.Errorf("Failed to cast choice: %s", choice.String)
	}
}

func askInputChoice(input ConfigCommandInput) (any, error) {
	var choices = make([]*selection.Choice, 0, len(input.Options))
	prompt := fmt.Sprintf("%s %s", promptStyle.Render("Choose a"), inputNameStyle.Render(input.Name))
	for label, value := range input.Options {
		choices = append(choices, &selection.Choice{String: label, Value: value})
	}
	sp := selection.New(prompt, choices)

	if len(choices) <= minChoiceFiltering {
		sp.Filter = nil
	}

	if choice, err := sp.RunPrompt(); err == nil {
		return choice.Value, err
	} else {
		return nil, err
	}
}

func askInputText(input ConfigCommandInput) (any, error) {
	prompt := fmt.Sprintf("%s %s", promptStyle.Render("Please specify a"), inputNameStyle.Render(input.Name))
	ti := textinput.New(prompt)
	ti.InitialValue = input.DefaultValue
	ti.Validate = func(value string) bool {
		return input.Valid(value)
	}
	return ti.RunPrompt()
}

func askInput(input ConfigCommandInput) (any, error) {
	if input.Selectable() {
		return askInputChoice(input)
	} else {
		return askInputText(input)
	}
}

func askInputs(inputs ConfigCommandInputs) (InputValues, error) {
	values := make(InputValues, len(inputs))
	for _, input := range inputs {
		value, err := askInput(input)
		if err != nil {
			return values, err
		}
		values[input.Name] = value
	}
	return values, nil
}
