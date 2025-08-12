package main

import (
	"flag"
	"reflect"
	"testing"
)

func TestParseCommand(t *testing.T) {
	cfg := &Config{
		Commands: map[string]Command{
			"test": {
				Description: "Test command",
				Commands: map[string]Command{
					"sub": {
						Description: "Sub command",
					},
				},
			},
			"alias": {
				Aliases: []string{"a"},
			},
		},
	}

	// Test case 1: Select command by argument
	t.Run("select command by argument", func(t *testing.T) {
		flag.CommandLine = flag.NewFlagSet("test", flag.ExitOnError)
		flag.CommandLine.Parse([]string{"test"})

		cmd, cmdArgs, err := parseCommand(cfg)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if cmd.Description != "Test command" {
			t.Errorf("expected description 'Test command', got '%s'", cmd.Description)
		}
		if !reflect.DeepEqual(cmdArgs, []string{"test"}) {
			t.Errorf("expected cmdArgs '[test]', got '%v'", cmdArgs)
		}
	})

	// Test case 2: Select subcommand by arguments
	t.Run("select subcommand by arguments", func(t *testing.T) {
		flag.CommandLine = flag.NewFlagSet("test", flag.ExitOnError)
		flag.CommandLine.Parse([]string{"test", "sub"})

		cmd, cmdArgs, err := parseCommand(cfg)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if cmd.Description != "Sub command" {
			t.Errorf("expected description 'Sub command', got '%s'", cmd.Description)
		}
		if !reflect.DeepEqual(cmdArgs, []string{"test", "sub"}) {
			t.Errorf("expected cmdArgs '[test sub]', got '%v'", cmdArgs)
		}
	})

	// Test case 3: Select command by alias
	t.Run("select command by alias", func(t *testing.T) {
		flag.CommandLine = flag.NewFlagSet("test", flag.ExitOnError)
		flag.CommandLine.Parse([]string{"a"})

		cmd, _, err := parseCommand(cfg)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if cmd.Aliases[0] != "a" {
			t.Errorf("expected alias 'a', got '%s'", cmd.Aliases[0])
		}
	})
}
