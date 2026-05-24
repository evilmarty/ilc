package ilc

import (
	"text/template"

	"github.com/evilmarty/ilc/internal/inputs"
)

var DefaultShell = []string{"/bin/sh"}

type CommandAliases []string

type Inputs struct {
	*inputs.FlagSet
}

type Command struct {
	Name        string `yaml:"-"`
	Description string
	Run         string
	Shell       []string
	Env         EnvMap
	Pure        bool
	Inputs      Inputs
	Commands    SubCommands `yaml:",flow"`
}

func (command Command) String() string {
	return command.Name
}

func (command Command) Runnable() bool {
	return command.Run != "" && len(command.Commands) == 0
}

func (command Command) Get(name string) (SubCommand, bool) {
	for _, subcommand := range command.Commands {
		if name == subcommand.Name {
			return subcommand, true
		}
		for _, alias := range subcommand.Aliases {
			if name == alias {
				return subcommand, true
			}
		}
	}
	return SubCommand{}, false
}

type SubCommand struct {
	Command `yaml:",inline"`
	Aliases CommandAliases `yaml:",flow"`
}

type SubCommands []SubCommand

func (command Command) Validate() error {
	if command.Run != "" {
		_, err := template.New(command.Name).Funcs(defaultTemplateFuncs).Parse(command.Run)
		if err != nil {
			return &TemplateError{
				Type:    "run",
				Command: command.Name,
				Err:     err,
			}
		}
	}

	for name, envTmpl := range command.Env {
		_, err := template.New(name).Funcs(defaultTemplateFuncs).Parse(envTmpl)
		if err != nil {
			return &TemplateError{
				Type:      "env",
				Command:   command.Name,
				FieldName: name,
				Err:       err,
			}
		}
	}

	for _, subcommand := range command.Commands {
		if err := subcommand.Validate(); err != nil {
			return err
		}
	}

	return nil
}
