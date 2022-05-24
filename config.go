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
	Name         string `yaml:"-"`
	DefaultValue string `yaml:"default"`
	Pattern      string
	Options      ConfigCommandInputOptions
}

func (cci ConfigCommandInput) Selectable() bool {
	return len(cci.Options) > 0
}

func (cci ConfigCommandInput) Valid(value any) bool {
	if cci.Selectable() {
		return cci.contains(value)
	} else {
		return cci.matches(value)
	}
}

func (cci ConfigCommandInput) contains(value any) bool {
	for _, option := range cci.Options {
		if option == value {
			return true
		}
	}
	return false
}

func (cci ConfigCommandInput) matches(value any) bool {
	s, ok := value.(string)
	if !ok {
		return false
	}
	matched, _ := regexp.MatchString(cci.Pattern, s)
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
	Commands    ConfigCommands
}

func (cc *ConfigCommand) HasSubCommands() bool {
	return len(cc.Commands) > 0
}

type ConfigCommands []ConfigCommand

func (cc *ConfigCommands) Get(name string) *ConfigCommand {
	for _, command := range *cc {
		if command.Name == name {
			return &command
		}
	}
	return nil
}

func (cc *ConfigCommands) Inputs() ConfigCommandInputs {
	inputs := make(ConfigCommandInputs, 0)
	for _, command := range *cc {
		inputs = append(inputs, command.Inputs...)
	}
	return inputs
}

func (x *ConfigCommands) UnmarshalYAML(value *yaml.Node) error {
	if value.Kind != yaml.MappingNode || len(value.Content)%2 != 0 {
		return fmt.Errorf("line %d: cannot unmarshal commands into map", value.Line)
	}

	commands := ConfigCommands{}
	content := value.Content

	for len(content) > 0 {
		keyNode := content[0]
		valueNode := content[1]

		if keyNode.Kind != yaml.ScalarNode || !(valueNode.Kind == yaml.MappingNode || valueNode.Kind == yaml.ScalarNode) {
			return fmt.Errorf("line %d: unexpected node type", keyNode.Line)
		}

		var command ConfigCommand
		if valueNode.Kind == yaml.ScalarNode {
			command = ConfigCommand{
				Name: keyNode.Value,
				Run:  valueNode.Value,
			}
		} else {
			if err := valueNode.Decode(&command); err != nil {
				return err
			}
			if numCommands := len(command.Commands); command.Run == "" && numCommands == 0 {
				return fmt.Errorf("line %d: '%s' command missing run or commands attribute", keyNode.Line, keyNode.Value)
			} else if command.Run != "" && numCommands > 0 {
				return fmt.Errorf("line %d: '%s' command cannot have both run and commands attribute", keyNode.Line, keyNode.Value)
			}
			command.Name = keyNode.Value
		}
		commands = append(commands, command)

		content = content[2:]
	}

	*x = commands

	return nil
}

type Config struct {
	Description string
	Shell       []string
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
