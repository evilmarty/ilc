package main

import (
	"errors"
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

var ErrInvalidValue = errors.New("Invalid value")

func selectCommand(command ConfigCommand) (ConfigCommand, error) {
	commandsLength := len(command.Commands)
	maxNameLength := maxStringLength(command.Commands, func(c ConfigCommand) string {
		return c.Name
	})

	sp := selection.New("Choose command", command.Commands)
	sp.SelectedChoiceStyle = func(c *selection.Choice[ConfigCommand]) string {
		return renderChoiceStyle(c.Value.Name, c.Value.Description, maxNameLength, true)
	}
	sp.UnselectedChoiceStyle = func(c *selection.Choice[ConfigCommand]) string {
		return renderChoiceStyle(c.Value.Name, c.Value.Description, maxNameLength, false)
	}

	if commandsLength <= minChoiceFiltering {
		sp.Filter = nil
	}

	if choice, err := sp.RunPrompt(); err != nil {
		return command, err
	} else {
		return choice, nil
	}
}

func askInput(input ConfigInput) (any, error) {
	if input.Selectable() {
		return selectInput(input)
	} else {
		return getInput(input)
	}
}

func selectInput(input ConfigInput) (any, error) {
	prompt := input.Description
	if prompt == "" {
		prompt = fmt.Sprintf("Choose a %s", termenv.String(input.Name).Underline().String())
	}
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

func getInput(input ConfigInput) (any, error) {
	prompt := input.Description
	if prompt == "" {
		prompt = fmt.Sprintf("Please specify a %s", termenv.String(input.Name).Underline().String())
	}
	ti := textinput.New(prompt)
	ti.InitialValue = fmt.Sprint(input.DefaultValue)
	ti.Validate = func(s string) error {
		if _, ok := input.Parse(s); ok {
			return nil
		} else {
			return ErrInvalidValue
		}
	}
	if s, err := ti.RunPrompt(); err != nil {
		return s, err
	} else if value, ok := input.Parse(s); ok {
		return value, nil
	} else {
		return nil, ErrInvalidValue
	}
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
