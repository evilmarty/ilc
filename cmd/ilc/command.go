package main

import (
	"flag"
	"fmt"
	"slices"

	tea "github.com/charmbracelet/bubbletea"
)

func selectCommand(cfg *Config) (*Command, []string, error) {
	cmd, cmdArgs, err := parseCommand(cfg)
	if err != nil {
		return nil, nil, err
	}

	// If help flag is present, don't run interactive selector
	for _, arg := range flag.Args() {
		if arg == "--help" || arg == "-h" {
			return cmd, cmdArgs, nil
		}
	}

	if cmd != nil && len(cmd.Commands) == 0 {
		return cmd, cmdArgs, nil
	}

	return runCommandSelector(cfg, cmd, cmdArgs)
}

func parseCommand(cfg *Config) (*Command, []string, error) {
	args := flag.Args()
	commands := cfg.Commands
	var cmd *Command
	var cmdArgs []string

	for i, arg := range args {
		found := false
		for name, c := range commands {
			if name == arg {
				cmd = &c
				commands = c.Commands
				cmdArgs = append(cmdArgs, arg)
				found = true
				break
			}
			if slices.Contains(c.Aliases, arg) {
				cmd = &c
				commands = c.Commands
				cmdArgs = append(cmdArgs, arg)
				found = true
			}
		}

		if !found {
			return nil, nil, fmt.Errorf("command not found: %s", arg)
		}

		if len(commands) == 0 || i == len(args)-1 {
			break
		}
	}

	return cmd, cmdArgs, nil
}

func runCommandSelector(cfg *Config, cmd *Command, cmdArgs []string) (*Command, []string, error) {
	commands := cfg.Commands
	if cmd != nil {
		commands = cmd.Commands
	}

	for {
		var commandNames []string
		for name := range commands {
			commandNames = append(commandNames, name)
		}

		initialModel := commandModel{commands: commandNames}
		p := tea.NewProgram(initialModel)

		m, err := p.Run()
		if err != nil {
			return nil, nil, err
		}

		selectedCommand := m.(commandModel).selected
		c := commands[selectedCommand]
		cmd = &c
		cmdArgs = append(cmdArgs, selectedCommand)

		if len(cmd.Commands) > 0 {
			commands = cmd.Commands
		} else {
			break
		}
	}

	return cmd, cmdArgs, nil
}
