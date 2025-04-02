package main

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConfigSelect(t *testing.T) {
	config := Config{
		Commands: SubCommands{
			SubCommand{
				Command: Command{
					Name: "foo",
					Commands: SubCommands{
						SubCommand{
							Command: Command{Name: "bar"},
						},
						SubCommand{
							Command: Command{Name: "baz"},
						},
					},
				},
			},
		},
	}
	t.Run("selects all with no remaining args", func(t *testing.T) {
		expected_commands := Selection{
			Command(config),
			config.Commands[0].Command,
			config.Commands[0].Command.Commands[1].Command,
		}
		expected_args := []string{}
		actual_commands, actual_args := config.Select([]string{"foo", "baz"})
		assert.Equal(t, expected_commands, actual_commands)
		assert.Equal(t, expected_args, actual_args)
	})
	t.Run("selects all with remaining args", func(t *testing.T) {
		expected_commands := Selection{
			Command(config),
			config.Commands[0].Command,
			config.Commands[0].Command.Commands[1].Command,
		}
		expected_args := []string{"a", "b"}
		actual_commands, actual_args := config.Select([]string{"foo", "baz", "a", "b"})
		assert.Equal(t, expected_commands, actual_commands)
		assert.Equal(t, expected_args, actual_args)
	})
	t.Run("selects some with no remaining args", func(t *testing.T) {
		expected_commands := Selection{
			Command(config),
			config.Commands[0].Command,
		}
		expected_args := []string{}
		actual_commands, actual_args := config.Select([]string{"foo"})
		assert.Equal(t, expected_commands, actual_commands)
		assert.Equal(t, expected_args, actual_args)
	})
	t.Run("selects some with remaining args", func(t *testing.T) {
		expected_commands := Selection{
			Command(config),
			config.Commands[0].Command,
		}
		expected_args := []string{"a", "b"}
		actual_commands, actual_args := config.Select([]string{"foo", "a", "b"})
		assert.Equal(t, expected_commands, actual_commands)
		assert.Equal(t, expected_args, actual_args)
	})
	t.Run("selects none with no remaining args", func(t *testing.T) {
		expected_commands := Selection{
			Command(config),
		}
		expected_args := []string{}
		actual_commands, actual_args := config.Select([]string{})
		assert.Equal(t, expected_commands, actual_commands)
		assert.Equal(t, expected_args, actual_args)
	})
	t.Run("selects none with remaining args", func(t *testing.T) {
		expected_commands := Selection{
			Command(config),
		}
		expected_args := []string{"a", "b"}
		actual_commands, actual_args := config.Select([]string{"a", "b"})
		assert.Equal(t, expected_commands, actual_commands)
		assert.Equal(t, expected_args, actual_args)
	})
}

func TestLoadConfig(t *testing.T) {
	content := `
commands:
  test:
    run: go test
    inputs:
      bool: boolean
      num:
        type: number
        min: -1
        max: 10
      sequence:
        options: [A, B]
      map:
        options:
          a: A
          b: B
`
	expected := Config{
		Commands: SubCommands{
			{
				Command: Command{
					Name: "test",
					Run:  "go test",
					Inputs: Inputs{
						{
							Name:  "bool",
							Value: &BooleanValue{},
						},
						{
							Name:  "num",
							Value: &NumberValue{MinValue: -1.0, MaxValue: 10.0},
						},
						{
							Name:  "sequence",
							Value: &StringValue{},
							Options: InputOptions{
								{Label: "A", Value: "A"},
								{Label: "B", Value: "B"},
							},
						},
						{
							Name:  "map",
							Value: &StringValue{},
							Options: InputOptions{
								{Label: "a", Value: "A"},
								{Label: "b", Value: "B"},
							},
						},
					},
				},
			},
		},
	}
	tempFile, err := os.CreateTemp("", "")
	assert.NoError(t, err, "Failed to create temp file")

	_, err = tempFile.Write([]byte(content))
	assert.NoError(t, err, "Failed to write config to temp file")

	actual, err := LoadConfig(tempFile.Name())
	assert.NoError(t, err, "LoadConfig() returned an unexpected error")
	assert.Equal(t, expected, actual, "LoadConfig() returned unexpected results")
}
