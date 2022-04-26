package main

import (
	"fmt"
	"io/ioutil"
	"regexp"

	"gopkg.in/yaml.v3"
)

type ConfigCommandInputOptions map[string]string

func (x *ConfigCommandInputOptions) UnmarshalYAML(node *yaml.Node) error {
	var mapValue map[string]string

	switch node.Kind {
	case yaml.SequenceNode:
		var seqValue []string
		if err := node.Decode(&seqValue); err != nil {
			return err
		}
		mapValue = make(map[string]string, len(seqValue))
		for _, item := range seqValue {
			mapValue[item] = item
		}
	case yaml.MappingNode:
		if err := node.Decode(&mapValue); err != nil {
			return err
		}
	}

	*x = mapValue

	return nil
}

type ConfigCommandInput struct {
	Name    string `yaml:"-"`
	Default string
	Pattern string
	Options ConfigCommandInputOptions
}

func (cci ConfigCommandInput) CanSelect() bool {
	return len(cci.Options) > 0
}

func (cci ConfigCommandInput) Validate(value string) bool {
	if cci.CanSelect() {
		return cci.contains(value)
	} else {
		return cci.matches(value)
	}
}

func (cci ConfigCommandInput) contains(value string) bool {
	for _, option := range cci.Options {
		if option == value {
			return true
		}
	}
	return false
}

func (cci ConfigCommandInput) matches(value string) bool {
	matched, _ := regexp.MatchString(cci.Pattern, value)
	return matched
}

type ConfigCommandInputs []ConfigCommandInput

func (x *ConfigCommandInputs) UnmarshalYAML(value *yaml.Node) error {
	if value.Kind != yaml.MappingNode {
		return fmt.Errorf("line %d: cannot unmarshal commands into map", value.Line)
	}

	var name string
	inputs := ConfigCommandInputs{}

	for _, node := range value.Content {
		switch node.Kind {
		case yaml.ScalarNode:
			name = node.Value
		case yaml.MappingNode:
			var input ConfigCommandInput
			if err := node.Decode(&input); err != nil {
				return err
			}
			input.Name = name
			inputs = append(inputs, input)
		}
	}

	*x = inputs

	return nil
}

type ConfigCommand struct {
	Name        string `yaml:"-"`
	Description string
	Run         string
	Env         map[string]string
	Inputs      ConfigCommandInputs
}

type ConfigCommands []ConfigCommand

func (x *ConfigCommands) UnmarshalYAML(value *yaml.Node) error {
	if value.Kind != yaml.MappingNode {
		return fmt.Errorf("line %d: cannot unmarshal commands into map", value.Line)
	}

	var name string
	commands := ConfigCommands{}

	for _, node := range value.Content {
		switch node.Kind {
		case yaml.ScalarNode:
			name = node.Value
		case yaml.MappingNode:
			var command ConfigCommand
			if err := node.Decode(&command); err != nil {
				return err
			}
			command.Name = name
			commands = append(commands, command)
		}
	}

	*x = commands

	return nil
}

type Config struct {
	Description string
	Commands    ConfigCommands `yaml:",flow"`
}

func LoadConfig(path string) (*Config, error) {
	var config Config

	if content, err := ioutil.ReadFile(path); err != nil {
		return nil, err
	} else if err = yaml.Unmarshal(content, &config); err != nil {
		return nil, err
	}

	return &config, nil
}
