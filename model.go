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
	config  *Config
	command *ConfigCommand
	values  map[string]any
}

func (m *model) setCommand(command *ConfigCommand) {
	m.command = command
	m.values = make(map[string]any, len(command.Inputs))
}

func (m *model) parse(args []string) error {
	if len(args) == 0 {
		return nil
	}

	for _, command := range m.config.Commands {
		if command.Name == args[0] {
			m.setCommand(&command)
			break
		}
	}

	if m.command == nil {
		return fmt.Errorf("Unknown command: %s", args[0])
	}

	fs := m.getFlagSet()
	if err := fs.Parse(args[1:]); err != nil {
		return err
	}

	return nil
}

func (m *model) getFlagSet() *flag.FlagSet {
	usage := fmt.Sprintf("command '%s'", m.command.Name)
	fs := flag.NewFlagSet(usage, flag.ExitOnError)
	for _, input := range m.command.Inputs {
		m.addFlagSetInput(fs, input)
	}

	return fs
}

func (m *model) addFlagSetInput(fs *flag.FlagSet, input ConfigCommandInput) {
	fs.Func(input.Name, "", func(value string) error {
		if input.Validate(value) {
			m.values[input.Name] = value
			return nil
		} else {
			return fmt.Errorf("Invalid value given")
		}
	})
}

func (m *model) askCommand() error {
	if m.command != nil {
		return nil
	}

	command, err := askCommand(m.config.Commands)
	if err == nil {
		m.setCommand(command)
	}

	return err
}

func (m *model) askInputs() error {
	if m.command == nil {
		return nil
	}

	for _, input := range m.command.Inputs {
		if _, ok := m.values[input.Name]; ok {
			continue
		}

		value, err := askInput(&input)
		if err != nil {
			return err
		} else {
			m.values[input.Name] = value
		}
	}

	return nil
}

func (m *model) ask() error {
	if err := m.askCommand(); err != nil {
		return err
	}
	if err := m.askInputs(); err != nil {
		return err
	}
	return nil
}

func (m *model) env() []string {
	if m.command == nil {
		return []string{}
	}

	var env = make([]string, len(m.command.Env))

	for name, rawValue := range m.command.Env {
		if value, err := RenderTemplate(rawValue, m.values); err == nil {
			env = append(env, fmt.Sprintf("%s=%s", name, value))
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

func (m *model) exec() error {
	if m.command == nil {
		return fmt.Errorf("No command specified")
	}

	script, err := RenderTemplate(m.command.Run, m.values)
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

func (m *model) Run(args []string) error {
	if err := m.parse(args); err != nil {
		return err
	}

	if err := m.ask(); err != nil {
		return err
	}

	if err := m.exec(); err != nil {
		return err
	}

	return nil
}

func newModel(configFile string) (*model, error) {
	config, err := LoadConfig(configFile)
	if err != nil {
		return nil, err
	}

	return &model{config: config}, nil
}
