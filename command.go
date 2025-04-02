package main

var DefaultShell = []string{"/bin/sh"}

type CommandAliases []string

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
