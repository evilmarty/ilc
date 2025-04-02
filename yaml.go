package main

import (
	"fmt"
	"regexp"

	orderedmap "github.com/wk8/go-ordered-map/v2"
	"gopkg.in/yaml.v3"
)

type commandName string

func (x *commandName) UnmarshalYAML(node *yaml.Node) error {
	var s string
	if err := node.Decode(&s); err != nil {
		return err
	}
	if !validName(s) {
		return fmt.Errorf("line %d: invalid command name", node.Line)
	}
	*x = commandName(s)
	return nil
}

func (x *CommandAliases) UnmarshalYAML(node *yaml.Node) error {
	var aliases []commandName
	if err := node.Decode(&aliases); err != nil {
		return err
	}
	for _, alias := range aliases {
		*x = append(*x, string(alias))
	}
	return nil
}

type yamlSubCommand SubCommand

func (x *yamlSubCommand) UnmarshalYAML(node *yaml.Node) error {
	var subcommand SubCommand
	if node.Kind == yaml.ScalarNode {
		subcommand.Run = node.Value
	} else if err := node.Decode(&subcommand); err != nil {
		return err
	}
	*x = yamlSubCommand(subcommand)
	return nil
}

func (x *SubCommands) UnmarshalYAML(node *yaml.Node) error {
	om := orderedmap.New[commandName, yamlSubCommand]()
	if err := node.Decode(&om); err != nil {
		return err
	}
	for pair := om.Oldest(); pair != nil; pair = pair.Next() {
		pair.Value.Name = string(pair.Key)
		*x = append(*x, SubCommand(pair.Value))
	}
	return nil
}

func (x *InputOptions) UnmarshalYAML(node *yaml.Node) error {
	if node.Kind == yaml.SequenceNode {
		var values []string
		if err := node.Decode(&values); err != nil {
			return err
		}
		for _, value := range values {
			*x = append(*x, InputOption{value, value})
		}
	} else {
		om := orderedmap.New[string, string]()
		if err := node.Decode(&om); err != nil {
			return err
		}
		for pair := om.Oldest(); pair != nil; pair = pair.Next() {
			*x = append(*x, InputOption{pair.Key, pair.Value})
		}
	}
	return nil
}

type inputName string

func (x *inputName) UnmarshalYAML(node *yaml.Node) error {
	var s string
	if err := node.Decode(&s); err != nil {
		return err
	}
	if !validName(s) {
		return fmt.Errorf("line %d: invalid input name", node.Line)
	}
	*x = inputName(s)
	return nil
}

type yamlInputType struct {
	Type string
}

func (y yamlInputType) newValue() Value {
	switch y.Type {
	case "number":
		return &NumberValue{}
	case "boolean":
		return &BooleanValue{}
	default:
		return &StringValue{}
	}
}

type yamlInput Input

func (x *yamlInput) UnmarshalYAML(node *yaml.Node) error {
	var inputType yamlInputType
	var input Input
	if node.Kind == yaml.ScalarNode {
		inputType.Type = node.Value
	} else if err := node.Decode(&inputType); err != nil {
		return err
	}
	input.Value = inputType.newValue()
	if node.Kind == yaml.MappingNode {
		if err := node.Decode(&input); err != nil {
			return err
		}
		if err := node.Decode(input.Value); err != nil {
			return err
		}
	}
	*x = yamlInput(input)
	return nil
}

func (x *Inputs) UnmarshalYAML(node *yaml.Node) error {
	om := orderedmap.New[inputName, yamlInput]()
	if err := node.Decode(&om); err != nil {
		return err
	}
	for pair := om.Oldest(); pair != nil; pair = pair.Next() {
		pair.Value.Name = string(pair.Key)
		// In some situations where the value is scalar it does not call UnmarshalYAML so it
		// is set to the default values. If so then it should be defaulted to be a string value.
		if pair.Value.Value == nil {
			pair.Value.Value = &StringValue{}
		}
		*x = append(*x, Input(pair.Value))
	}
	return nil
}

func validName(s string) bool {
	m, _ := regexp.MatchString("^[a-zA-Z0-9][a-zA-Z0-9-_]*$", s)
	return m
}
