package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCommandString(t *testing.T) {
	command := Command{Name: "foobar"}
	assert.Equal(t, "foobar", command.String())
}

func TestCommandRunnable(t *testing.T) {
	t.Run("when empty", func(t *testing.T) {
		command := Command{}
		assert.False(t, command.Runnable())
	})
	t.Run("when has run and no subcommands", func(t *testing.T) {
		command := Command{Run: "echo foobar"}
		assert.True(t, command.Runnable())
	})
	t.Run("when has run and subcommands", func(t *testing.T) {
		command := Command{Run: "echo foobar", Commands: SubCommands{SubCommand{}}}
		assert.False(t, command.Runnable())
	})
	t.Run("when subcommands", func(t *testing.T) {
		command := Command{Commands: SubCommands{SubCommand{}}}
		assert.False(t, command.Runnable())
	})
}

func TestCommandGet(t *testing.T) {
	command := Command{
		Commands: SubCommands{
			{
				Command: Command{Name: "foobar"},
				Aliases: CommandAliases{"foobaz"},
			},
		},
	}
	t.Run("by name", func(t *testing.T) {
		expected := command.Commands[0]
		actual, found := command.Get("foobar")
		assert.True(t, found)
		assert.Equal(t, expected, actual)
	})
	t.Run("by alias", func(t *testing.T) {
		expected := command.Commands[0]
		actual, found := command.Get("foobaz")
		assert.True(t, found)
		assert.Equal(t, expected, actual)
	})
	t.Run("not found", func(t *testing.T) {
		expected := SubCommand{}
		actual, found := command.Get("nope")
		assert.False(t, found)
		assert.Equal(t, expected, actual)
	})
}
