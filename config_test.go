package main

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConfigInputOptionsLen(t *testing.T) {
	options := ConfigInputOptions{}
	assert.Equal(t, 0, options.Len(), "ConfigInputOptions.Len() expected to return zero")

	options = ConfigInputOptions{
		{Label: "A", Value: "A"},
	}
	assert.Equal(t, 1, options.Len(), "ConfigInputOptions.Len() expected to return one")
}

func TestConfigInputOptionsContains(t *testing.T) {
	options := ConfigInputOptions{}
	assert.Equal(t, false, options.Contains("foobar"), "ConfigInputOptions.Contains() with no items expected to return false")

	options = ConfigInputOptions{
		{Label: "Foobar", Value: "foobar"},
	}
	assert.Equal(t, true, options.Contains("foobar"), "ConfigInputOptions.Contains() that has value expected to return true")
}

func TestConfigInputSelectable(t *testing.T) {
	input := ConfigInput{
		Options: ConfigInputOptions{},
	}
	assert.Equal(t, false, input.Selectable(), "ConfigInput.Selectable() with no options expected to return false")

	input = ConfigInput{
		Options: ConfigInputOptions{
			{Label: "A", Value: "A"},
		},
	}
	assert.Equal(t, true, input.Selectable(), "ConfigInput.Selectable() with options expected to return true")
}

func TestConfigInputSafeName(t *testing.T) {
	input := ConfigInput{
		Name: "foo-bar",
	}
	assert.Equal(t, "foo_bar", input.SafeName(), "ConfigInput.SafeName() returned unexpected result")
}

func TestConfigInputValid(t *testing.T) {
	input := ConfigInput{
		Options: ConfigInputOptions{},
	}
	assert.Equal(t, true, input.Valid(""), "ConfigInput.Valid() with empty string expected to return true")

	input = ConfigInput{
		Options: ConfigInputOptions{
			{Label: "A", Value: "A"},
		},
	}
	assert.Equal(t, false, input.Valid("foobar"), "ConfigInput.Valid() with no matching options expected to return false")
	assert.Equal(t, true, input.Valid("A"), "ConfigInput.Valid() with matching options expected to return true")

	input = ConfigInput{
		Options: ConfigInputOptions{},
		Pattern: "^foo",
	}
	assert.Equal(t, false, input.Valid("booboo"), "ConfigInput.Valid() when pattern does not match expected to return false")
	assert.Equal(t, true, input.Valid("foobar"), "ConfigInput.Valid() when pattern does match expected to return true")
}

func TestConfigCommandsGet(t *testing.T) {
	a := ConfigCommand{Name: "a", Aliases: []string{"A"}}
	b := ConfigCommand{Name: "b", Aliases: []string{"B"}}
	commands := ConfigCommands{a, b}
	t.Run("matches name", func(t *testing.T) {
		actual := commands.Get("b")
		assert.Equal(t, &b, actual, "ConfigCommands.Get() did not find command")
	})
	t.Run("matches alias", func(t *testing.T) {
		actual := commands.Get("B")
		assert.Equal(t, &b, actual, "ConfigCommands.Get() did not find command")
	})
	t.Run("no match", func(t *testing.T) {
		var expected *ConfigCommand
		actual := commands.Get("c")
		assert.Equal(t, expected, actual, "ConfigCommands.Get() returned unexpected result")
	})
}

func TestParseConfig_CommandsAliases(t *testing.T) {
	t.Run("duplicate aliases", func(t *testing.T) {
		content := `
commands:
  foobar:
    run: true
  foobaz:
    run: true
    aliases:
      - foobar
`
		expected := "line 6: alias 'foobar' already defined by command 'foobar'"
		_, actual := ParseConfig([]byte(content))
		assert.EqualError(t, actual, expected, "ParseConfig() returned unexpected error")
	})
	t.Run("invalid alias", func(t *testing.T) {
		content := `
commands:
  foobar:
    run: true
    aliases:
      - _foobar
`
		expected := "line 4: invalid command alias '_foobar'"
		_, actual := ParseConfig([]byte(content))
		assert.EqualError(t, actual, expected, "ParseConfig() returned unexpected error")
	})
}

func TestParseConfig_CommandsNotAMap(t *testing.T) {
	content := `
commands: ooops
`
	expected := "line 2: cannot unmarshal commands into map"
	_, actual := ParseConfig([]byte(content))
	assert.EqualError(t, actual, expected, "ParseConfig() returned unexpected error")
}

func TestParseConfig_CommandName(t *testing.T) {
	content := `
commands:
  _foobar: ooops
`
	expected := "line 3: invalid command name '_foobar'"
	_, actual := ParseConfig([]byte(content))
	assert.EqualError(t, actual, expected, "ParseConfig() returned unexpected error")
}

func TestParseConfig_CommandInvalid(t *testing.T) {
	content := `
commands:
  invalidCommand: []
`
	expected := "line 3: invalid definition for command 'invalidCommand'"
	_, actual := ParseConfig([]byte(content))
	assert.EqualError(t, actual, expected, "ParseConfig() returned unexpected error")
}

func TestParseConfig_CommandNoRunOrSubcommands(t *testing.T) {
	content := `
commands:
  invalidCommand:
    name: ooops
`
	expected := "line 3: command 'invalidCommand' must have either 'run' or 'commands' attribute"
	_, actual := ParseConfig([]byte(content))
	assert.EqualError(t, actual, expected, "ParseConfig() returned unexpected error")
}

func TestParseConfig_InputOptionsIsMap(t *testing.T) {
	content := `
run: ok
inputs:
  foobar:
    options:
      a: A
      b: B
`
	expected := Config{
		Run: "ok",
		Inputs: ConfigInputs{
			ConfigInput{
				Name: "foobar",
				Options: ConfigInputOptions{
					{Label: "a", Value: "A"},
					{Label: "b", Value: "B"},
				},
			},
		},
	}
	actual, err := ParseConfig([]byte(content))
	assert.NoError(t, err, "ParseConfig() returned unexpected error")
	assert.Equal(t, expected, actual, "ParseConfig() returned unexpected error")
}

func TestParseConfig_InputOptionsIsSequence(t *testing.T) {
	content := `
run: ok
inputs:
  foobar:
    options:
      - A
      - B
`
	expected := Config{
		Run: "ok",
		Inputs: ConfigInputs{
			ConfigInput{
				Name: "foobar",
				Options: ConfigInputOptions{
					{Label: "A", Value: "A"},
					{Label: "B", Value: "B"},
				},
			},
		},
	}
	actual, err := ParseConfig([]byte(content))
	assert.NoError(t, err, "ParseConfig() returned unexpected error")
	assert.Equal(t, expected, actual, "ParseConfig() returned unexpected error")
}

func TestParseConfig_InputOptionsIsScaler(t *testing.T) {
	content := `
run: ok
inputs:
  foobar:
    options: ooops
`
	expected := "line 5: unexpected node type"
	_, actual := ParseConfig([]byte(content))
	assert.EqualError(t, actual, expected, "ParseConfig() returned unexpected error")
}

func TestParseConfig_InputsIsMap(t *testing.T) {
	content := `
run: ok
inputs:
  foobar:
    default: Foobar
`
	expected := Config{
		Run: "ok",
		Inputs: ConfigInputs{
			ConfigInput{
				Name:         "foobar",
				DefaultValue: "Foobar",
			},
		},
	}
	actual, err := ParseConfig([]byte(content))
	assert.NoError(t, err, "ParseConfig() returned unexpected error")
	assert.Equal(t, expected, actual, "ParseConfig() returned unexpected error")
}

func TestParseConfig_InputsIsSequence(t *testing.T) {
	content := `
run: ooops
inputs: []
`
	expected := "line 3: cannot unmarshal inputs into map"
	_, actual := ParseConfig([]byte(content))
	assert.EqualError(t, actual, expected, "ParseConfig() returned unexpected error")
}

func TestParseConfig_InputsIsScalar(t *testing.T) {
	content := `
run: ooops
inputs: nope
`
	expected := "line 3: cannot unmarshal inputs into map"
	_, actual := ParseConfig([]byte(content))
	assert.EqualError(t, actual, expected, "ParseConfig() returned unexpected error")
}

func TestParseConfig_InputNames(t *testing.T) {
	content := `
run: ooops
inputs:
  _invalid:
    default: Nope
`
	expected := "line 5: invalid input name"
	_, actual := ParseConfig([]byte(content))
	assert.EqualError(t, actual, expected, "ParseConfig() returned unexpected error")
}

func TestLoadConfig(t *testing.T) {
	content := `
commands:
  test:
    run: go test
    inputs:
      sequence:
        options: [A, B]
      map:
        options:
          a: A
          b: B
`
	expected := Config{
		Commands: ConfigCommands{
			{
				Name: "test",
				Run:  "go test",
				Inputs: ConfigInputs{
					{
						Name: "sequence",
						Options: ConfigInputOptions{
							{Label: "A", Value: "A"},
							{Label: "B", Value: "B"},
						},
					},
					{
						Name: "map",
						Options: ConfigInputOptions{
							{Label: "a", Value: "A"},
							{Label: "b", Value: "B"},
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
