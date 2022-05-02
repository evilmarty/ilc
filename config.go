package main

import (
	"fmt"
	"io/ioutil"
	"regexp"

	"github.com/kr/pretty"
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

func (cci ConfigCommandInput) Selectable() bool {
	return len(cci.Options) > 0
}

func (cci ConfigCommandInput) Validate(value string) bool {
	if cci.Selectable() {
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
	matched, err := regexp.MatchString(cci.Pattern, value)
	pretty.Println(err)
	return matched
}

type ConfigCommandInputs []ConfigCommandInput

func (x *ConfigCommandInputs) UnmarshalYAML(value *yaml.Node) error {
	if value.Kind != yaml.MappingNode || len(value.Content)%2 != 0 {
		return fmt.Errorf("line %d: cannot unmarshal inputs into map", value.Line)
	}

	inputs := ConfigCommandInputs{}
	content := value.Content

	for len(content) > 0 {
		keyNode := content[0]
		valueNode := content[1]

		if keyNode.Kind != yaml.ScalarNode {
			return fmt.Errorf("line %d: unexpected node type", keyNode.Line)
		}

		switch valueNode.Kind {
		case yaml.ScalarNode:
			inputs = append(inputs, ConfigCommandInput{Name: keyNode.Value})
		case yaml.MappingNode:
			var input ConfigCommandInput
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

type ConfigCommand struct {
	Name        string `yaml:"-"`
	Description string
	Run         string
	Env         map[string]string
	Inputs      ConfigCommandInputs
}

type ConfigCommands []ConfigCommand

func (x *ConfigCommands) UnmarshalYAML(value *yaml.Node) error {
	if value.Kind != yaml.MappingNode || len(value.Content)%2 != 0 {
		return fmt.Errorf("line %d: cannot unmarshal commands into map", value.Line)
	}

	commands := ConfigCommands{}
	content := value.Content

	for len(content) > 0 {
		keyNode := content[0]
		valueNode := content[1]

		if keyNode.Kind != yaml.ScalarNode || valueNode.Kind != yaml.MappingNode {
			return fmt.Errorf("line %d: unexpected node type", keyNode.Line)
		}

		var command ConfigCommand
		if err := valueNode.Decode(&command); err != nil {
			return err
		}
		command.Name = keyNode.Value
		commands = append(commands, command)

		content = content[2:]
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
