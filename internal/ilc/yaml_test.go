package ilc

import (
	"testing"

	"github.com/evilmarty/ilc/internal/inputs"
	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"
)

func TestCommandAliasesUnmarshalYAML(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		content := `
- foobar
- foobaz
`
		expected := CommandAliases{"foobar", "foobaz"}
		var actual CommandAliases
		err := yaml.Unmarshal([]byte(content), &actual)
		assert.NoError(t, err)
		assert.Equal(t, expected, actual)
	})
	t.Run("invalid", func(t *testing.T) {
		content := `
- _foobar
- foobaz
`
		err := yaml.Unmarshal([]byte(content), &CommandAliases{})
		assert.ErrorContains(t, err, "invalid command name")
	})
}

func TestSubCommandsUnmarshalYAML(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		content := `
inline_command: echo foobar
command:
  description: foobar
  run: echo foobaz
  env:
    A: a
  pure: true
  aliases:
    - foo
subcommands:
  description: has subcommands
  commands:
    subcommand: echo foobar
`
		expected := SubCommands{
			{
				Command: Command{
					Name: "inline_command",
					Run:  "echo foobar",
				},
			},
			{
				Aliases: CommandAliases{"foo"},
				Command: Command{
					Name:        "command",
					Run:         "echo foobaz",
					Description: "foobar",
					Env:         map[string]string{"A": "a"},
					Pure:        true,
				},
			},
			{
				Command: Command{
					Name:        "subcommands",
					Description: "has subcommands",
					Commands: SubCommands{
						{
							Command: Command{
								Name: "subcommand",
								Run:  "echo foobar",
							},
						},
					},
				},
			},
		}
		var actual SubCommands
		err := yaml.Unmarshal([]byte(content), &actual)
		assert.NoError(t, err)
		assert.Equal(t, expected, actual)
	})
	t.Run("invalid name", func(t *testing.T) {
		content := `
-invalid: echo foobar
`
		err := yaml.Unmarshal([]byte(content), &SubCommands{})
		assert.ErrorContains(t, err, "invalid command name")
	})
}

func TestInputOptionsUnmarshalYAML(t *testing.T) {
	t.Run("as sequence", func(t *testing.T) {
		content := `
- a
- b
`
		expected := yamlInputOptions{
			{Label: "a", Value: "a"},
			{Label: "b", Value: "b"},
		}
		var actual yamlInputOptions
		err := yaml.Unmarshal([]byte(content), &actual)
		assert.NoError(t, err)
		assert.Equal(t, expected, actual)
	})
	t.Run("as map", func(t *testing.T) {
		content := `
a: A
b: B
`
		expected := yamlInputOptions{
			{Label: "a", Value: "A"},
			{Label: "b", Value: "B"},
		}
		var actual yamlInputOptions
		err := yaml.Unmarshal([]byte(content), &actual)
		assert.NoError(t, err)
		assert.Equal(t, expected, actual)
	})
}

func TestInputsUnmarshalYAML(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		content := `
inline_string: string
string:
  type: string
  default: foobar
  pattern: "^foo"
  options:
    a: foobar
    b: foobaz
default_string:
  pattern: "^foo"
  default: foobaz
inline_number: number
number:
  type: number
  default: 1
  min: 1
  max: 2
inline_boolean: boolean
boolean:
  type: boolean
  default: true
`
		expected := func() Inputs {
			fs := inputs.NewFlagSet("ilc", EnvVarPrefix)
			fs.Var(&inputs.Input{Name: "inline_string", Value: &inputs.StringValue{}})
			fs.Var(&inputs.Input{
				Name:  "string",
				Value: &inputs.StringValue{Value: "foobar", Pattern: "^foo"},
				Options: inputs.InputOptions{
					{Label: "a", Value: "foobar"},
					{Label: "b", Value: "foobaz"},
				},
			})
			fs.Var(&inputs.Input{Name: "default_string", Value: &inputs.StringValue{Value: "foobaz", Pattern: "^foo"}})
			fs.Var(&inputs.Input{Name: "inline_number", Value: &inputs.NumberValue{}})
			fs.Var(&inputs.Input{Name: "number", Value: &inputs.NumberValue{Value: 1.0, MinValue: 1.0, MaxValue: 2.0}})
			fs.Var(&inputs.Input{Name: "inline_boolean", Value: &inputs.BooleanValue{}})
			fs.Var(&inputs.Input{Name: "boolean", Value: &inputs.BooleanValue{Value: true}})
			return Inputs{FlagSet: fs}
		}()
		var actual Inputs
		err := yaml.Unmarshal([]byte(content), &actual)
		assert.NoError(t, err)
		assert.Equal(t, expected, actual)
	})
	t.Run("invalid name", func(t *testing.T) {
		content := `
-invalid: string
`
		err := yaml.Unmarshal([]byte(content), &Inputs{})
		assert.ErrorContains(t, err, "invalid input name")
	})
	t.Run("empty input", func(t *testing.T) {
		content := `
empty_input:
`
		expected := func() Inputs {
			fs := inputs.NewFlagSet("ilc", EnvVarPrefix)
			fs.Var(&inputs.Input{Name: "empty_input", Value: &inputs.StringValue{Value: "", Pattern: ""}})
			return Inputs{FlagSet: fs}
		}()
		var actual Inputs
		err := yaml.Unmarshal([]byte(content), &actual)
		assert.NoError(t, err)
		assert.Equal(t, expected, actual)
	})
}
