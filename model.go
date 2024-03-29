package main

import (
	"flag"
	"fmt"
)

type model struct {
	config   Config
	commands CommandChain
	values   InputValues
	showHelp bool
}

func (m *model) Selected() bool {
	return len(m.commands) > 0
}

func (m *model) Inputs() Inputs {
	return m.commands.Inputs()
}

func (m *model) outstandingInputs() Inputs {
	inputs := m.Inputs()
	outstandingInputs := make(Inputs, 0)

	for _, input := range inputs {
		if _, ok := m.values[input.Name]; !ok {
			outstandingInputs = append(outstandingInputs, input)
		}
	}

	return outstandingInputs
}

func (m *model) AvailableCommands() Commands {
	if m.Selected() {
		return m.commands[len(m.commands)-1].Commands
	} else {
		return m.config.Commands
	}
}

func (m *model) askCommands() error {
	if lastCommand := m.commands.Last(); lastCommand != nil {
		subcommands, err := lastCommand.Commands.Select()
		m.commands = append(m.commands, subcommands...)
		return err
	} else {
		commands, err := m.config.Commands.Select()
		m.commands = commands
		return err
	}
}

func (m *model) askInputs() error {
	outstandingInputs := m.outstandingInputs()
	values, err := outstandingInputs.Get()
	if err != nil {
		return err
	}

	for key, value := range values {
		m.values[key] = value
	}

	return nil
}

func (m *model) ask() error {
	if err := m.askCommands(); err != nil {
		return err
	}
	if err := m.askInputs(); err != nil {
		return err
	}
	return nil
}

func (m *model) env(baseEnv []string) []string {
	var env = make([]string, 0)

	if !m.commands.Pure() {
		env = append(env, baseEnv...)
	}

	for _, command := range m.commands {
		for name, rawValue := range command.Env {
			if value, err := RenderTemplate(rawValue, m.values); err == nil {
				env = append(env, fmt.Sprintf("%s=%s", name, value))
			}
		}
	}

	return env
}

func (m *model) renderScript() (string, error) {
	command := m.commands.Last()
	if command == nil {
		return "", fmt.Errorf("no command specified")
	}

	if command.Run == "" {
		return "", fmt.Errorf("invalid run command for %s", command.Name)
	}
	return RenderTemplate(command.Run, m.values)
}

func (m *model) exec(env []string) error {
	script, err := m.renderScript()
	if err != nil {
		return err
	}

	cmd := ScriptCommand(script)
	cmd.Env = m.env(env)
	return cmd.Run()
}

func (m *model) Run(env []string) error {
	if err := m.ask(); err != nil {
		return err
	}

	if err := m.exec(env); err != nil {
		return err
	}

	return nil
}

func parseCommands(initCommands Commands, args []string) (CommandChain, []string) {
	commands := initCommands
	foundCommands := make(CommandChain, 0)
	for len(args) > 0 {
		command := commands.Get(args[0])
		if command == nil {
			break
		}
		foundCommands = append(foundCommands, command)
		commands = command.Commands
		args = args[1:]
	}
	return foundCommands, args
}

func parseInputValues(inputs Inputs, args []string) (bool, InputValues, error) {
	values := make(InputValues, len(inputs))
	fs := flag.NewFlagSet("", flag.ExitOnError)
	showHelp := fs.Bool("help", false, "Show this help screen")

	for _, input := range inputs {
		fs.Func(input.Name, "", func(value string) error {
			if input.Valid(value) {
				values[input.Name] = value
				return nil
			} else {
				return fmt.Errorf("invalid value given for input '%s'", input.Name)
			}
		})
	}

	err := fs.Parse(args)
	return *showHelp, values, err
}

func newModel(configFile string, args []string) (*model, error) {
	config, err := LoadConfig(configFile)
	if err != nil {
		return nil, err
	}
	SetShell(config.Shell)
	commands, remainingArgs := parseCommands(config.Commands, args)
	showHelp, values, err := parseInputValues(commands.Inputs(), remainingArgs)

	if err != nil {
		return nil, err
	}

	return &model{config: *config, commands: commands, values: values, showHelp: showHelp}, nil
}
