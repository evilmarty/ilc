package main

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUsage_EmptyCommands(t *testing.T) {
	u := usageFixture()
	u.commands = [][]string{}
	expected := `
test

this is a fixture


USAGE
  ilc config.yaml subcommand [inputs]

INPUTS
  -c                   c input
  -d                   d input


`
	assert.Equal(t, expected, u.String(), "Usage.String() should not include entrypoint")
}

func TestUsage_EmptyInputs(t *testing.T) {
	u := usageFixture()
	u.inputs = [][]string{}
	expected := `
test

this is a fixture


USAGE
  ilc config.yaml subcommand <commands>

COMMANDS
  a, aa                a subcommand
  b                    b subcommand


`
	assert.Equal(t, expected, u.String(), "Usage.String() should not include entrypoint")
}

func TestUsage_Entrypoint_Empty(t *testing.T) {
	u := usageFixture()
	u.Entrypoint = []string{}
	expected := `
test

this is a fixture


USAGE
  <config> <commands> [inputs]

COMMANDS
  a, aa                a subcommand
  b                    b subcommand

INPUTS
  -c                   c input
  -d                   d input


`
	assert.Equal(t, expected, u.String(), "Usage.String() should not include entrypoint")
}

func TestUsage_Entrypoint_One(t *testing.T) {
	u := usageFixture()
	u.Entrypoint = []string{"ilc"}
	expected := `
test

this is a fixture


USAGE
  ilc <config> <commands> [inputs]

COMMANDS
  a, aa                a subcommand
  b                    b subcommand

INPUTS
  -c                   c input
  -d                   d input


`
	assert.Equal(t, expected, u.String(), "Usage.String() should not include entrypoint")
}

func TestUsage_Entrypoint_Two(t *testing.T) {
	u := usageFixture()
	u.Entrypoint = []string{"ilc", "config.yaml"}
	expected := `
test

this is a fixture


USAGE
  ilc config.yaml <commands> [inputs]

COMMANDS
  a, aa                a subcommand
  b                    b subcommand

INPUTS
  -c                   c input
  -d                   d input


`
	assert.Equal(t, expected, u.String(), "Usage.String() should not include entrypoint")
}

func TestUsage_Entrypoint_Many(t *testing.T) {
	u := usageFixture()
	u.Entrypoint = []string{"ilc", "config.yaml", "command", "subcommand"}
	expected := `
test

this is a fixture


USAGE
  ilc config.yaml command subcommand <commands> [inputs]

COMMANDS
  a, aa                a subcommand
  b                    b subcommand

INPUTS
  -c                   c input
  -d                   d input


`
	assert.Equal(t, expected, u.String(), "Usage.String() should not include entrypoint")
}

func TestUsage_Title_Blank(t *testing.T) {
	u := usageFixture()
	u.Title = ""
	expected := `
this is a fixture


USAGE
  ilc config.yaml subcommand <commands> [inputs]

COMMANDS
  a, aa                a subcommand
  b                    b subcommand

INPUTS
  -c                   c input
  -d                   d input


`
	assert.Equal(t, expected, u.String(), "Usage.String() should not include entrypoint")
}

func TestUsage_Description_Blank(t *testing.T) {
	u := usageFixture()
	u.Description = ""
	expected := `
test


USAGE
  ilc config.yaml subcommand <commands> [inputs]

COMMANDS
  a, aa                a subcommand
  b                    b subcommand

INPUTS
  -c                   c input
  -d                   d input


`
	assert.Equal(t, expected, u.String(), "Usage.String() should not include entrypoint")
}

func usageFixture() Usage {
	commands := []ConfigCommand{
		{Name: "a", Description: "a subcommand", Aliases: []string{"aa"}},
		{Name: "b", Description: "b subcommand"},
	}
	inputs := []ConfigInput{
		{Name: "c", Description: "c input"},
		{Name: "d", Description: "d input"},
	}
	u := NewUsage(os.Stdout).ImportCommands(commands).ImportInputs(inputs)
	u.Entrypoint = []string{"ilc", "config.yaml", "subcommand"}
	u.Title = "test"
	u.Description = "this is a fixture"
	return u
}
