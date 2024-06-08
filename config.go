package main

import (
	"fmt"
	"os"
	"regexp"

	"gopkg.in/yaml.v3"
)

type ConfigInputOption struct {
	Label string
	Value string
}

type ConfigInputOptions []ConfigInputOption

func (options ConfigInputOptions) Len() int {
	return len(options)
}

func (options ConfigInputOptions) Contains(value any) bool {
	for _, item := range options {
		if item.Value == value {
			return true
		}
	}

	return false
}

func (x *ConfigInputOptions) UnmarshalYAML(node *yaml.Node) error {
	var options ConfigInputOptions

	switch node.Kind {
	case yaml.SequenceNode:
		var seqValue []string
		if err := node.Decode(&seqValue); err != nil {
			return err
		}
		for _, option := range seqValue {
			options = append(options, ConfigInputOption{
				Label: option,
				Value: option,
			})
		}
	case yaml.MappingNode:
		content := node.Content
		for len(content) > 0 {
			options = append(options, ConfigInputOption{
				Label: content[0].Value,
				Value: content[1].Value,
			})
			content = content[2:]
		}
	default:
		return fmt.Errorf("line %d: unexpected node type", node.Line)
	}

	*x = options

	return nil
}

type ConfigInput struct {
	Name         string `yaml:"-"`
	DefaultValue string `yaml:"default"`
	Pattern      string
	Options      ConfigInputOptions
	Description  string
}

func (input *ConfigInput) Selectable() bool {
	return input.Options.Len() > 0
}

func (input *ConfigInput) Valid(value string) bool {
	if input.Selectable() {
		return input.Options.Contains(value)
	} else if input.Pattern != "" {
		matched, _ := regexp.MatchString(input.Pattern, value)
		return matched
	} else {
		return true
	}
}

type ConfigInputs []ConfigInput

func (x *ConfigInputs) UnmarshalYAML(value *yaml.Node) error {
	if value.Kind != yaml.MappingNode || len(value.Content)%2 != 0 {
		return fmt.Errorf("line %d: cannot unmarshal inputs into map", value.Line)
	}

	inputs := ConfigInputs{}
	content := value.Content

	for len(content) > 0 {
		keyNode := content[0]
		valueNode := content[1]

		if keyNode.Kind != yaml.ScalarNode {
			return fmt.Errorf("line %d: unexpected node type", keyNode.Line)
		}

		switch valueNode.Kind {
		case yaml.ScalarNode:
			inputs = append(inputs, ConfigInput{Name: keyNode.Value})
		case yaml.MappingNode:
			var input ConfigInput
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
	Shell       []string
	Env         map[string]string
	Pure        bool
	Inputs      ConfigInputs
	Commands    ConfigCommands `yaml:",flow"`
}

type ConfigCommands []ConfigCommand

func (commands *ConfigCommands) Get(name string) *ConfigCommand {
	for _, command := range *commands {
		if command.Name == name {
			return &command
		}
	}
	return nil
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
			return fmt.Errorf("line %d: invalid definition for command '%s'", keyNode.Line, keyNode.Value)
		}

		var command ConfigCommand
		if valueNode.Kind == yaml.ScalarNode {
			command.Name = keyNode.Value
			command.Run = valueNode.Value
		} else {
			if err := valueNode.Decode(&command); err != nil {
				return err
			}
			if numCommands := len(command.Commands); command.Run == "" && numCommands == 0 {
				return fmt.Errorf("line %d: command '%s' must have either 'run' or 'commands' attribute", keyNode.Line, keyNode.Value)
			} else if command.Run != "" && numCommands > 0 {
				return fmt.Errorf("line %d: command '%s' must only have 'run' or 'commands' attribute", keyNode.Line, keyNode.Value)
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
	Run         string
	Shell       []string
	Env         map[string]string
	Pure        bool
	Inputs      ConfigInputs
	Commands    ConfigCommands `yaml:",flow"`
}

func (c Config) Runnable() bool {
	return c.Run != ""
}

func ParseConfig(content []byte) (Config, error) {
	var config Config

	if err := yaml.Unmarshal(content, &config); err != nil {
		return config, err
	}

	return config, nil
}

func LoadConfig(path string) (Config, error) {
	logger.Printf("Attempting to load config file: %s", path)
	if content, err := os.ReadFile(path); err != nil {
		return Config{}, err
	} else {
		return ParseConfig(content)
	}
}
