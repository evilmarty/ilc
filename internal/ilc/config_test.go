package ilc

import (
	"errors"
	"os"
	"testing"

	"github.com/evilmarty/ilc/internal/inputs"
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
		expected := NewSelection(
			Command(config),
			config.Commands[0].Command,
			config.Commands[0].Command.Commands[1].Command,
		)
		actual := config.Select([]string{"foo", "baz"})
		assert.Equal(t, expected, actual)
	})
	t.Run("selects all with remaining args", func(t *testing.T) {
		expected := Selection{
			commands: []Command{
				Command(config),
				config.Commands[0].Command,
				config.Commands[0].Command.Commands[1].Command,
			},
			Args: []string{"a", "b"},
		}
		actual := config.Select([]string{"foo", "baz", "a", "b"})
		assert.Equal(t, expected, actual)
	})
	t.Run("selects some with no remaining args", func(t *testing.T) {
		expected := NewSelection(
			Command(config),
			config.Commands[0].Command,
		)
		actual := config.Select([]string{"foo"})
		assert.Equal(t, expected, actual)
	})
	t.Run("selects some with remaining args", func(t *testing.T) {
		expected := Selection{
			commands: []Command{
				Command(config),
				config.Commands[0].Command,
			},
			Args: []string{"a", "b"},
		}
		actual := config.Select([]string{"foo", "a", "b"})
		assert.Equal(t, expected, actual)
	})
	t.Run("selects none with no remaining args", func(t *testing.T) {
		expected := NewSelection(Command(config))
		actual := config.Select([]string{})
		assert.Equal(t, expected, actual)
	})
	t.Run("selects none with remaining args", func(t *testing.T) {
		expected := Selection{
			commands: []Command{
				Command(config),
			},
			Args: []string{"a", "b"},
		}
		actual := config.Select([]string{"a", "b"})
		assert.Equal(t, expected, actual)
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
					Inputs: func() Inputs {
						fs := inputs.NewFlagSet("ilc", EnvVarPrefix)
						fs.Var(&inputs.Input{Name: "bool", Value: &inputs.BooleanValue{}})
						fs.Var(&inputs.Input{Name: "num", Value: &inputs.NumberValue{MinValue: -1.0, MaxValue: 10.0}})
						fs.Var(&inputs.Input{
							Name:  "sequence",
							Value: &inputs.StringValue{},
							Options: inputs.InputOptions{
								{Label: "A", Value: "A"},
								{Label: "B", Value: "B"},
							},
						})
						fs.Var(&inputs.Input{
							Name:  "map",
							Value: &inputs.StringValue{},
							Options: inputs.InputOptions{
								{Label: "a", Value: "A"},
								{Label: "b", Value: "B"},
							},
						})
						return Inputs{FlagSet: fs}
					}(),
				},
			},
		},
	}
	tempFile, err := os.CreateTemp("", "")
	assert.NoError(t, err, "Failed to create temp file")
	defer os.Remove(tempFile.Name())

	_, err = tempFile.Write([]byte(content))
	assert.NoError(t, err, "Failed to write config to temp file")

	actual, err := LoadConfig(tempFile.Name())
	assert.NoError(t, err, "LoadConfig() returned an unexpected error")
	assert.Equal(t, expected, actual, "LoadConfig() returned unexpected results")
}

func TestConfigValidate(t *testing.T) {
	t.Run("valid templates", func(t *testing.T) {
		cfg := Config{
			Run: "echo {{.Input.Name}}",
			Env: map[string]string{
				"GREETING": "hello {{.Input.Greeting}}",
			},
			Commands: SubCommands{
				{
					Command: Command{
						Name: "child",
						Run:  "echo {{.Input.Child}}",
					},
				},
			},
		}
		err := cfg.Validate()
		assert.NoError(t, err)
	})

	t.Run("invalid run template", func(t *testing.T) {
		cfg := Config{
			Name: "test-cmd",
			Run:  "echo {{.Input.Name", // missing closing braces
		}
		err := cfg.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid run template in command \"test-cmd\"")

		// Verify custom structured TemplateError properties
		var tmplErr *TemplateError
		if assert.True(t, errors.As(err, &tmplErr)) {
			assert.Equal(t, "run", tmplErr.Type)
			assert.Equal(t, "test-cmd", tmplErr.Command)
			assert.Equal(t, "", tmplErr.FieldName)
			assert.Error(t, tmplErr.Err)
		}
	})

	t.Run("invalid env template", func(t *testing.T) {
		cfg := Config{
			Name: "test-cmd",
			Env: map[string]string{
				"BAD_VAR": "hello {{.Input.Greeting", // missing closing braces
			},
		}
		err := cfg.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid env template \"BAD_VAR\" in command \"test-cmd\"")

		// Verify custom structured TemplateError properties
		var tmplErr *TemplateError
		if assert.True(t, errors.As(err, &tmplErr)) {
			assert.Equal(t, "env", tmplErr.Type)
			assert.Equal(t, "test-cmd", tmplErr.Command)
			assert.Equal(t, "BAD_VAR", tmplErr.FieldName)
			assert.Error(t, tmplErr.Err)
		}
	})
}
