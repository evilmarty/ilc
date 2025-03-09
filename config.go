package main

import (
	"fmt"
	"math"
	"os"
	"regexp"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

var defaultBooleanOptions = ConfigInputOptions{
	ConfigInputOption{Label: "yes", Value: true},
	ConfigInputOption{Label: "no", Value: false},
}

type ConfigInputOption struct {
	Label string
	Value any
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
		var seqValue []interface{}
		if err := node.Decode(&seqValue); err != nil {
			return err
		}
		for _, option := range seqValue {
			options = append(options, ConfigInputOption{
				Label: fmt.Sprint(option),
				Value: option,
			})
		}
	case yaml.MappingNode:
		var mapValue map[string]any
		if err := node.Decode(&mapValue); err != nil {
			return err
		}
		for label, value := range mapValue {
			options = append(options, ConfigInputOption{
				Label: label,
				Value: value,
			})
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
	DefaultValue any     `yaml:"default"`
	MinValue     float64 `yaml:"min"`
	MaxValue     float64 `yaml:"max"`
	Pattern      string
	Options      ConfigInputOptions
	Description  string
}

func (input *ConfigInput) Parse(s string) (any, bool) {
	switch input.Type {
	case "number":
		if strings.Contains(s, ".") {
			n, err := strconv.ParseFloat(s, 64)
			return n, err == nil
		} else {
			n, err := strconv.ParseInt(s, 10, 64)
			return n, err == nil
		}
	case "boolean":
		b, err := strconv.ParseBool(s)
		return b, err == nil
	case "string":
		return s, true
	default:
		return s, false
	}
}

func (input *ConfigInput) SafeName() string {
	return strings.ReplaceAll(input.Name, "-", "_")
}

func (input *ConfigInput) Selectable() bool {
	return input.Options != nil && input.Options.Len() > 0
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
	case "number":
		return input.bound(value)
	default:
		return true
	}
}

func (input *ConfigInput) match(value any) bool {
	if input.Pattern == "" {
		return true
	}
	matched, _ := regexp.MatchString(input.Pattern, fmt.Sprint(value))
	return matched
}

func (input *ConfigInput) bound(value any) bool {
	// If min and max are the same then just ignore them
	if input.MinValue == input.MaxValue {
		return true
	}
	n := ToFloat64(value)
	return !math.IsNaN(n) && input.MinValue <= n && input.MaxValue >= n
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
		case "number":
		case "boolean":
			if input.DefaultValue == nil {
				input.DefaultValue = false
			}
			if input.Options == nil {
				input.Options = defaultBooleanOptions
			}
		case "":
			input.Type = "string"
		default:
			return fmt.Errorf("line %d: unsupported input type '%s'", valueNode.Line, input.Type)
		}
		if !validValue(input.DefaultValue, input.Type) {
			return fmt.Errorf("line %d: default value type mismatch", valueNode.Line)
		}
		if input.Options != nil {
			for _, option := range input.Options {
				if !validValue(option.Value, input.Type) {
					return fmt.Errorf("line %d: option value type mismatch", valueNode.Line)
				}
			}
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

func validValue(v any, t string) bool {
	switch v.(type) {
	case string:
		return t == "string"
	case int, uint, int8, uint8, int16, uint16, int32, uint32, int64, uint64, float32, float64:
		return t == "number"
	case bool:
		return t == "boolean"
	case nil:
		return true
	default:
		return false
	}
}
