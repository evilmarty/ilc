package ilc

import (
	"fmt"
	"regexp"

	"github.com/evilmarty/ilc/internal/inputs"
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

type yamlInputOptions inputs.InputOptions

func (x *yamlInputOptions) UnmarshalYAML(node *yaml.Node) error {
	if node.Kind == yaml.SequenceNode {
		var values []string
		if err := node.Decode(&values); err != nil {
			return err
		}
		for _, value := range values {
			*x = append(*x, inputs.InputOption{Label: value, Value: value})
		}
	} else {
		om := orderedmap.New[string, string]()
		if err := node.Decode(&om); err != nil {
			return err
		}
		for pair := om.Oldest(); pair != nil; pair = pair.Next() {
			*x = append(*x, inputs.InputOption{Label: pair.Key, Value: pair.Value})
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

func (y yamlInputType) newValue() inputs.Value {
	switch y.Type {
	case "number":
		return &inputs.NumberValue{}
	case "boolean":
		return &inputs.BooleanValue{}
	default:
		return &inputs.StringValue{}
	}
}

type yamlInput struct {
	Name        string
	Description string
	Options     yamlInputOptions
	Value       inputs.Value
}

func (x *yamlInput) UnmarshalYAML(node *yaml.Node) error {
	var inputType yamlInputType
	if node.Kind == yaml.ScalarNode {
		inputType.Type = node.Value
	} else if err := node.Decode(&inputType); err != nil {
		return err
	}

	val := inputType.newValue()

	type tempInput struct {
		Description string           `yaml:"description"`
		Options     yamlInputOptions `yaml:"options,flow"`
	}
	var temp tempInput

	if node.Kind == yaml.MappingNode {
		if err := node.Decode(&temp); err != nil {
			return err
		}
		if err := node.Decode(val); err != nil {
			return err
		}
	}

	if _, isBool := val.(*inputs.BooleanValue); isBool && len(temp.Options) > 0 {
		var optionsNode *yaml.Node
		for i := 0; i < len(node.Content); i += 2 {
			if node.Content[i].Value == "options" {
				optionsNode = node.Content[i+1]
				break
			}
		}

		if optionsNode != nil {
			if optionsNode.Kind == yaml.SequenceNode {
				if len(temp.Options) != 2 {
					return fmt.Errorf("line %d: boolean input options array must have exactly 2 items, got %d", optionsNode.Line, len(temp.Options))
				}
				temp.Options[0] = inputs.InputOption{Label: temp.Options[0].Label, Value: "false"}
				temp.Options[1] = inputs.InputOption{Label: temp.Options[1].Label, Value: "true"}
			} else {
				hasTrue := false
				hasFalse := false
				var newOptions yamlInputOptions
				for _, opt := range temp.Options {
					key := opt.Label
					valStr := opt.Value
					if key == "true" {
						hasTrue = true
						newOptions = append(newOptions, inputs.InputOption{Label: valStr, Value: "true"})
					} else if key == "false" {
						hasFalse = true
						newOptions = append(newOptions, inputs.InputOption{Label: valStr, Value: "false"})
					} else {
						return fmt.Errorf("line %d: invalid boolean option key: %s (must be true or false)", optionsNode.Line, key)
					}
				}
				if !hasTrue || !hasFalse {
					return fmt.Errorf("line %d: boolean option map must contain both true and false keys", optionsNode.Line)
				}
				temp.Options = newOptions
			}
		}
	}

	x.Description = temp.Description
	x.Options = temp.Options
	x.Value = val
	return nil
}

func (x *Inputs) UnmarshalYAML(node *yaml.Node) error {
	om := orderedmap.New[inputName, yamlInput]()
	if err := node.Decode(&om); err != nil {
		return err
	}
	fs := inputs.NewFlagSet("ilc", EnvVarPrefix)
	for pair := om.Oldest(); pair != nil; pair = pair.Next() {
		pair.Value.Name = string(pair.Key)
		if pair.Value.Value == nil {
			pair.Value.Value = &inputs.StringValue{}
		}
		inp := inputs.Input{
			Name:        pair.Value.Name,
			Description: pair.Value.Description,
			Options:     inputs.InputOptions(pair.Value.Options),
			Value:       pair.Value.Value,
		}
		fs.Var(&inp)
	}
	*x = Inputs{FlagSet: fs}
	return nil
}

func validName(s string) bool {
	m, _ := regexp.MatchString("^[a-zA-Z0-9][a-zA-Z0-9-_]*$", s)
	return m
}
