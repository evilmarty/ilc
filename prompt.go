package main

import (
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/erikgeiser/promptkit/selection"
	"github.com/erikgeiser/promptkit/textinput"
	"github.com/muesli/termenv"
)

const (
	maxPadSpacing      = 5
	minChoiceFiltering = 5
	accentColor        = termenv.ANSI256Color(32)
)

func askCommands(commands SelectedCommands) (SelectedCommands, error) {
	if len(commands) == 0 {
		return commands, nil
	}
	command := commands[len(commands)-1]
	numCommands := len(command.Commands)
	maxNameLength := maxStringLength(command.Commands, func(c SubCommand) string {
		return c.Name
	})
	sp := selection.New("Choose command", command.Commands)
	sp.SelectedChoiceStyle = func(c *selection.Choice[SubCommand]) string {
		return renderChoiceStyle(c.Value.Name, c.Value.Description, maxNameLength, true)
	}
	sp.UnselectedChoiceStyle = func(c *selection.Choice[SubCommand]) string {
		return renderChoiceStyle(c.Value.Name, c.Value.Description, maxNameLength, false)
	}

	if numCommands <= minChoiceFiltering {
		sp.Filter = nil
	}

	if choice, err := sp.RunPrompt(); err != nil {
		return commands, err
	} else {
		logger.Printf("selected command: %s", choice)
		return append(commands, choice.Command), nil
	}
}

func askInputs(inputs Inputs) error {
	for _, input := range inputs {
		prompt := input.Description
		if prompt == "" {
			inputName := termenv.String(input.Name).Underline().String()
			if input.Selectable() {
				prompt = fmt.Sprintf("Choose a %s", inputName)
			} else {
				prompt = fmt.Sprintf("Please specify a %s", inputName)
			}
		}
		if input.Selectable() {
			sp := selection.New(prompt, input.Options)
			logger.Printf("choosing input: %s", input.Name)
			if option, err := sp.RunPrompt(); err != nil {
				return err
			} else if err := input.Value.Set(option.Value); err != nil {
				return err
			}
		} else {
			ti := textinput.New(prompt)
			ti.InitialValue = input.Value.String()
			ti.Validate = func(s string) error {
				return input.Value.Set(s)
			}
			logger.Printf("asking input: %s", input.Name)
			if _, err := ti.RunPrompt(); err != nil {
				return err
			}
		}
	}
	return nil
}

func renderChoiceStyle(name, desc string, maxNameLength int, selected bool) string {
	padLen := maxNameLength - utf8.RuneCountInString(name)
	if padLen < 0 {
		padLen = 0
	}
	if selected {
		name = termenv.String(name).Foreground(accentColor).Bold().String()
	}
	if desc != "" {
		desc = strings.Repeat(" ", padLen+maxPadSpacing) + termenv.String(desc).Faint().String()
	}
	return name + desc
}

func maxStringLength[T any](items []T, fun func(T) string) int {
	maxLen := 0
	for _, item := range items {
		s := fun(item)
		maxLen = max(maxLen, utf8.RuneCountInString(s))
	}
	return maxLen
}
