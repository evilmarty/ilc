package main

import (
	"fmt"
	"strings"
)

func printHelp(cfg *Config, cmd *Command) {
	if cfg == nil {
		printBasicHelp()
		return
	}

	if cmd == nil {
		printTopLevelHelp(cfg)
	} else {
		printCommandHelp(cmd)
	}
}

func printBasicHelp() {
	fmt.Println("The simple way to create a command-line utility")
	fmt.Println("\nUSAGE:")
	fmt.Println("  ilc [CONFIG] [COMMANDS] [INPUTS] [OPTIONS]")
	fmt.Println("\nOPTIONS:")
	fmt.Println("  -f, --file    Path to config file")
	fmt.Println("  --help        Show help")
	fmt.Println("  --version     Show version")
}

func printTopLevelHelp(cfg *Config) {
	if cfg.Description != "" {
		fmt.Println(cfg.Description)
	}
	fmt.Println("\nUSAGE:")
	fmt.Println("  ilc [CONFIG] [COMMANDS] [INPUTS] [OPTIONS]")

	if len(cfg.Commands) > 0 {
		fmt.Println("\nCOMMANDS:")
		for name, command := range cfg.Commands {
			fmt.Printf("  %-15s %s\n", name, command.Description)
		}
	}

	if len(cfg.Inputs) > 0 {
		fmt.Println("\nINPUTS:")
		for name, input := range cfg.Inputs {
			fmt.Printf("  -%-14s %s\n", name, input.Description)
		}
	}

	fmt.Println("\nOPTIONS:")
	fmt.Println("  -f, --file    Path to config file")
	fmt.Println("  --help        Show help")
	fmt.Println("  --version     Show version")
}

func printCommandHelp(cmd *Command) {
	if cmd.Description != "" {
		fmt.Println(cmd.Description)
	}
	fmt.Println("\nUSAGE:")
	// This is a simplification, a real implementation would show the full command path
	fmt.Println("  ilc [COMMAND] [SUBCOMMANDS] [INPUTS]")

	if len(cmd.Commands) > 0 {
		fmt.Println("\nSUBCOMMANDS:")
		for name, command := range cmd.Commands {
			fmt.Printf("  %-15s %s\n", name, command.Description)
		}
	}

	if len(cmd.Inputs) > 0 {
		fmt.Println("\nINPUTS:")
		for name, input := range cmd.Inputs {
			var usage []string
			if input.Type != "" {
				usage = append(usage, input.Type)
			}
			if input.Default != nil {
				usage = append(usage, fmt.Sprintf("default: %v", input.Default))
			}
			fmt.Printf("  -%-14s %s (%s)\n", name, input.Description, strings.Join(usage, ", "))
		}
	}
}
