package main

import (
	"fmt"
	"io/ioutil"

	"gopkg.in/yaml.v3"
)

type ConfigCommand struct {
	Name        string `yaml:"-"`
	Description string
	Run         string
	Env         map[string]string
	Inputs      Inputs
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

func (cc *ConfigCommands) Inputs() Inputs {
	inputs := make(Inputs, 0)
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
