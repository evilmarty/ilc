package main

import (
	"fmt"
	"testing"
)

func TestNewCommandSet_PreselectedFromArgs(t *testing.T) {
	config := Config{
		Commands: ConfigCommands{
			ConfigCommand{Name: "foobar", Run: "true"},
		},
	}
	cs, err := NewCommandSet(config, []string{"foobar", "-a", "1"})
	if err != nil {
		t.Fatalf("NewCommandSet() returned unexpected error: %v", err)
	}
	assertDeepEqual(t, cs.Config, config, "NewCommandSet() did not set config")
	assertDeepEqual(t, cs.Args, []string{"-a", "1"}, "NewCommandSet() has unexpected args")
}

func TestNewCommandSet_Invalid(t *testing.T) {
	config := Config{
		Commands: ConfigCommands{
			ConfigCommand{Name: "foobar", Run: "true"},
		},
	}
	_, err := NewCommandSet(config, []string{"foobaz", "-a", "1"})
	expected := "invalid subcommand: foobaz"
	actual := fmt.Sprintf("%s", err)
	assertEqual(t, expected, actual, "NewCommandSet() returned unexpected error")
}
