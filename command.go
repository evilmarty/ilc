package main

import (
	"fmt"

	"github.com/erikgeiser/promptkit/selection"
	"gopkg.in/yaml.v3"
)

type Command struct {
	Name        string `yaml:"-"`
	Description string
	Run         string
	Env         map[string]string
	Pure        bool
	Inputs      Inputs
	Commands    Commands
}

func (command *Command) HasSubCommands() bool {
	return len(command.Commands) > 0
}

type Commands []Command

func (commands *Commands) Get(name string) *Command {
	for _, command := range *commands {
		if command.Name == name {
			return &command
		}
	}
	return nil
}

func (initCommands Commands) Select() (CommandChain, error) {
	askedCommands := make(CommandChain, 0)
	commands := initCommands

	for numCommands := len(commands); numCommands > 0; numCommands = len(commands) {
		choices := make([]*selection.Choice, numCommands)

		for i, command := range commands {
			choices[i] = &selection.Choice{String: command.Name, Value: command}
		}

		prompt := promptStyle.Render("Choose command")
		sp := selection.New(prompt, choices)

		if numCommands <= minChoiceFiltering {
			sp.Filter = nil
		}

		choice, err := sp.RunPrompt()
		if err != nil {
			return askedCommands, err
		}

		command := choice.Value.(Command)
		askedCommands = append(askedCommands, &command)
		commands = command.Commands
	}

	return askedCommands, nil
}

func (x *Commands) UnmarshalYAML(value *yaml.Node) error {
	if value.Kind != yaml.MappingNode || len(value.Content)%2 != 0 {
		return fmt.Errorf("line %d: cannot unmarshal commands into map", value.Line)
	}

	commands := Commands{}
	content := value.Content

	for len(content) > 0 {
		keyNode := content[0]
		valueNode := content[1]

		if keyNode.Kind != yaml.ScalarNode || !(valueNode.Kind == yaml.MappingNode || valueNode.Kind == yaml.ScalarNode) {
			return fmt.Errorf("line %d: unexpected node type", keyNode.Line)
		}

		var command Command
		if valueNode.Kind == yaml.ScalarNode {
			command = Command{
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
