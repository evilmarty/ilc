package main

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	"gopkg.in/yaml.v3"
)

type ConfigInputOption struct {
	Label string
	Value string
}

func (option ConfigInputOption) String() string {
	return option.Label
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
	Type         string
	DefaultValue string `yaml:"default"`
	Pattern      string
	Options      ConfigInputOptions
	Description  string
}

func (input *ConfigInput) SafeName() string {
	return strings.ReplaceAll(input.Name, "-", "_")
}

func (input *ConfigInput) Selectable() bool {
	return input.Options.Len() > 0
}

func (input *ConfigInput) Valid(value any) bool {
	if input.Selectable() {
		return input.Options.Contains(value)
	}
	switch input.Type {
	case "": // In case type is empty assume string
		return input.match(value)
	case "string":
		return input.match(value)
	default:
		return true
	}
}

func (input *ConfigInput) match(value any) bool {
	if input.Pattern == "" {
		return true
	}
	matched, _ := regexp.MatchString(input.Pattern, fmt.Sprintf("%v", value))
	return matched
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

		var input ConfigInput
		switch valueNode.Kind {
		case yaml.ScalarNode:
			input.Name = keyNode.Value
			input.Type = valueNode.Value
		case yaml.MappingNode:
			if err := valueNode.Decode(&input); err != nil {
				return err
			}
			input.Name = keyNode.Value
		default:
			return fmt.Errorf("line %d: unexpected node type", valueNode.Line)
		}

		if !validName(input.Name) {
			return fmt.Errorf("line %d: invalid input name", valueNode.Line)
		}

		switch input.Type {
		case "string":
		case "":
			input.Type = "string"
		default:
			return fmt.Errorf("line %d: unsupported input type '%s'", valueNode.Line, input.Type)
		}

		inputs = append(inputs, input)
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
	Aliases     []string
}

func (command ConfigCommand) String() string {
	return command.Name
}

type ConfigCommands []ConfigCommand

func (commands *ConfigCommands) Get(name string) *ConfigCommand {
	for _, command := range *commands {
		if command.Name == name {
			return &command
		}
		for _, alias := range command.Aliases {
			if alias == name {
				return &command
			}
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
			}
			command.Name = keyNode.Value
		}

		if !validName(command.Name) {
			return fmt.Errorf("line %d: invalid command name '%s'", valueNode.Line, command.Name)
		}
		for _, alias := range command.Aliases {
			if !validName(alias) {
				return fmt.Errorf("line %d: invalid command alias '%s'", valueNode.Line, alias)
			}
			if c := commands.Get(alias); c != nil {
				return fmt.Errorf("line %d: alias '%s' already defined by command '%s'", valueNode.Line, alias, c.Name)
			}
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

func validName(s string) bool {
	m, _ := regexp.MatchString("^[a-zA-Z0-9][a-zA-Z0-9-_]*$", s)
	return m
}
