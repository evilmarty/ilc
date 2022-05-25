package main

import (
	"fmt"
	"regexp"

	"github.com/charmbracelet/lipgloss"
	"github.com/erikgeiser/promptkit/selection"
	"github.com/erikgeiser/promptkit/textinput"
	"gopkg.in/yaml.v3"
)

const (
	minChoiceFiltering = 5
)

var (
	promptStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("#666"))
	inputNameStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#0cc")).Bold(true)
)

type InputValues map[string]any

type InputOption struct {
	Label string
	Value string
}

type InputOptions []InputOption

func (x *InputOptions) UnmarshalYAML(node *yaml.Node) error {
	var options InputOptions

	switch node.Kind {
	case yaml.SequenceNode:
		var seqValue []string
		if err := node.Decode(&seqValue); err != nil {
			return err
		}
		options = make(InputOptions, len(seqValue))
		for i, item := range seqValue {
			options[i].Label = item
			options[i].Value = item
		}
	case yaml.MappingNode:
		var mapValue map[string]string
		if err := node.Decode(&mapValue); err != nil {
			return err
		}
		options = make(InputOptions, 0, len(mapValue))
		for label, value := range mapValue {
			options = append(options, InputOption{Label: label, Value: value})
		}
	}

	*x = options

	return nil
}

type Input struct {
	Name         string `yaml:"-"`
	DefaultValue string `yaml:"default"`
	Pattern      string
	Options      InputOptions
	Description  string
}

func (input Input) Selectable() bool {
	return len(input.Options) > 0
}

func (input Input) Valid(value any) bool {
	if input.Selectable() {
		return input.contains(value)
	} else {
		return input.matches(value)
	}
}

func (input Input) contains(value any) bool {
	for _, option := range input.Options {
		if option.Value == value {
			return true
		}
	}
	return false
}

func (input Input) matches(value any) bool {
	s, ok := value.(string)
	if !ok {
		return false
	}
	matched, _ := regexp.MatchString(input.Pattern, s)
	return matched
}

func (input Input) choose() (any, error) {
	var choices = make([]*selection.Choice, 0, len(input.Options))
	prompt := fmt.Sprintf("%s %s", promptStyle.Render("Choose a"), inputNameStyle.Render(input.Name))
	for _, option := range input.Options {
		choices = append(choices, &selection.Choice{String: option.Label, Value: option.Value})
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

func (input Input) get() (any, error) {
	prompt := fmt.Sprintf("%s %s", promptStyle.Render("Please specify a"), inputNameStyle.Render(input.Name))
	ti := textinput.New(prompt)
	ti.InitialValue = input.DefaultValue
	ti.Validate = func(value string) bool {
		return input.Valid(value)
	}
	return ti.RunPrompt()
}

func (input Input) Get() (any, error) {
	if input.Selectable() {
		return input.choose()
	} else {
		return input.get()
	}
}

type Inputs []Input

func (inputs Inputs) Get() (InputValues, error) {
	values := make(InputValues, len(inputs))
	for _, input := range inputs {
		value, err := input.Get()
		if err != nil {
			return values, err
		}
		values[input.Name] = value
	}
	return values, nil
}

func (x *Inputs) UnmarshalYAML(value *yaml.Node) error {
	if value.Kind != yaml.MappingNode || len(value.Content)%2 != 0 {
		return fmt.Errorf("line %d: cannot unmarshal inputs into map", value.Line)
	}

	inputs := Inputs{}
	content := value.Content

	for len(content) > 0 {
		keyNode := content[0]
		valueNode := content[1]

		if keyNode.Kind != yaml.ScalarNode {
			return fmt.Errorf("line %d: unexpected node type", keyNode.Line)
		}

		switch valueNode.Kind {
		case yaml.ScalarNode:
			inputs = append(inputs, Input{Name: keyNode.Value})
		case yaml.MappingNode:
			var input Input
			if err := valueNode.Decode(&input); err != nil {
				return err
			}
			input.Name = keyNode.Value
			inputs = append(inputs, input)
		default:
			return fmt.Errorf("line %d: unexpected node type", valueNode.Line)
		}

		content = content[2:]
	}

	*x = inputs

	return nil
}
