package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
)

var (
	DefaultShell = []string{"sh", "-c"}
)

type model struct {
	config   Config
	commands ConfigCommands
	values   InputValues
	showHelp bool
}

func (m *model) Selected() bool {
	return len(m.commands) > 0
}

func (m *model) Inputs() ConfigCommandInputs {
	return m.commands.Inputs()
}

func (m *model) outstandingInputs() ConfigCommandInputs {
	inputs := m.Inputs()
	outstandingInputs := make(ConfigCommandInputs, 0)

	for _, input := range inputs {
		if _, ok := m.values[input.Name]; !ok {
			outstandingInputs = append(outstandingInputs, input)
		}
	}

	return outstandingInputs
}

func (m *model) AvailableCommands() ConfigCommands {
	if m.Selected() {
		return m.commands[len(m.commands)-1].Commands
	} else {
		return m.config.Commands
	}
}

func (m *model) askCommands() error {
	numCommands := len(m.commands)
	if numCommands > 0 {
		lastCommand := m.commands[numCommands-1]
		subcommands, err := askCommands(lastCommand.Commands)
		m.commands = append(m.commands, subcommands...)
		return err
	} else {
		commands, err := askCommands(m.config.Commands)
		m.commands = commands
		return err
	}
}

func (m *model) askInputs() error {
	outstandingInputs := m.outstandingInputs()
	values, err := askInputs(outstandingInputs)
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

func (m *model) env() []string {
	var env = make([]string, 0)

	for _, command := range m.commands {
		for name, rawValue := range command.Env {
			if value, err := RenderTemplate(rawValue, m.values); err == nil {
				env = append(env, fmt.Sprintf("%s=%s", name, value))
			}
		}
	}

	return env
}

func (m *model) shell() []string {
	if len(m.config.Shell) > 0 {
		return m.config.Shell
	} else {
		return DefaultShell
	}
}

func (m *model) runScript() (string, error) {
	numCommands := len(m.commands)
	if numCommands == 0 {
		return "", fmt.Errorf("No command specified")
	}

	command := m.commands[numCommands-1]
	if command.Run == "" {
		return "", fmt.Errorf("Invalid run command for %s", command.Name)
	}
	return RenderTemplate(command.Run, m.values)
}

func (m *model) exec() error {
	script, err := m.runScript()
	if err != nil {
		return err
	}

	args := append(m.shell(), script)
	cmd := exec.Command(args[0], args[1:]...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = m.env()
	return cmd.Run()
}

func (m *model) Run() error {
	if err := m.ask(); err != nil {
		return err
	}

	if err := m.exec(); err != nil {
		return err
	}

	return nil
}

func parseCommands(initCommands ConfigCommands, args []string) (ConfigCommands, []string) {
	commands := initCommands
	foundCommands := make(ConfigCommands, 0)
	for len(args) > 0 {
		command := commands.Get(args[0])
		if command == nil {
			break
		}
		foundCommands = append(foundCommands, *command)
		commands = command.Commands
		args = args[1:]
	}
	return foundCommands, args
}

func parseInputValues(inputs ConfigCommandInputs, args []string) (bool, InputValues, error) {
	values := make(InputValues, len(inputs))
	fs := flag.NewFlagSet("", flag.ExitOnError)
	showHelp := fs.Bool("help", false, "Show this help screen")

	for _, input := range inputs {
		fs.Func(input.Name, "", func(value string) error {
			if input.Valid(value) {
				values[input.Name] = value
				return nil
			} else {
				return fmt.Errorf("Invalid value given for input '%s'", input.Name)
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
	commands, remainingArgs := parseCommands(config.Commands, args)
	showHelp, values, err := parseInputValues(commands.Inputs(), remainingArgs)

	if err != nil {
		return nil, err
	}

	return &model{config: *config, commands: commands, values: values, showHelp: showHelp}, nil
}
