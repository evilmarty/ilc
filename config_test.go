package main

import (
	"fmt"
	"os"
	"testing"
)

func TestConfigInputOptionsLen(t *testing.T) {
	options := ConfigInputOptions{}
	assertEqual(t, 0, options.Len(), "ConfigInputOptions.Len() expected to return zero")

	options = ConfigInputOptions{
		{Label: "A", Value: "A"},
	}
	assertEqual(t, 1, options.Len(), "ConfigInputOptions.Len() expected to return one")
}

func TestConfigInputOptionsContains(t *testing.T) {
	options := ConfigInputOptions{}
	assertEqual(t, false, options.Contains("foobar"), "ConfigInputOptions.Contains() with no items expected to return false")

	options = ConfigInputOptions{
		{Label: "Foobar", Value: "foobar"},
	}
	assertEqual(t, true, options.Contains("foobar"), "ConfigInputOptions.Contains() that has value expected to return true")
}

func TestConfigInputSelectable(t *testing.T) {
	input := ConfigInput{
		Options: ConfigInputOptions{},
	}
	assertEqual(t, false, input.Selectable(), "ConfigInput.Selectable() with no options expected to return false")

	input = ConfigInput{
		Options: ConfigInputOptions{
			{Label: "A", Value: "A"},
		},
	}
	assertEqual(t, true, input.Selectable(), "ConfigInput.Selectable() with options expected to return true")
}

func TestConfigInputValid(t *testing.T) {
	input := ConfigInput{
		Options: ConfigInputOptions{},
	}
	assertEqual(t, true, input.Valid(""), "ConfigInput.Valid() with empty string expected to return true")

	input = ConfigInput{
		Options: ConfigInputOptions{
			{Label: "A", Value: "A"},
		},
	}
	assertEqual(t, false, input.Valid("foobar"), "ConfigInput.Valid() with no matching options expected to return false")
	assertEqual(t, true, input.Valid("A"), "ConfigInput.Valid() with matching options expected to return true")

	input = ConfigInput{
		Options: ConfigInputOptions{},
		Pattern: "^foo",
	}
	assertEqual(t, false, input.Valid("booboo"), "ConfigInput.Valid() when pattern does not match expected to return false")
	assertEqual(t, true, input.Valid("foobar"), "ConfigInput.Valid() when pattern does match expected to return true")
}

func TestParseConfig_CommandsNotAMap(t *testing.T) {
	content := `
commands: ooops
`
	expected := "line 2: cannot unmarshal commands into map"
	_, err := ParseConfig([]byte(content))
	actual := fmt.Sprintf("%s", err)
	assertEqual(t, expected, actual, "ParseConfig() returned unexpected error")
}

func TestParseConfig_CommandName(t *testing.T) {
	content := `
commands:
  _foobar: ooops
`
	expected := "line 3: invalid command name"
	_, err := ParseConfig([]byte(content))
	actual := fmt.Sprintf("%s", err)
	assertEqual(t, expected, actual, "ParseConfig() returned unexpected error")
}

func TestParseConfig_CommandInvalid(t *testing.T) {
	content := `
commands:
  invalidCommand: []
`
	expected := "line 3: invalid definition for command 'invalidCommand'"
	_, err := ParseConfig([]byte(content))
	actual := fmt.Sprintf("%s", err)
	assertEqual(t, expected, actual, "ParseConfig() returned unexpected error")
}

func TestParseConfig_CommandNoRunOrSubcommands(t *testing.T) {
	content := `
commands:
  invalidCommand:
    name: ooops
`
	expected := "line 3: command 'invalidCommand' must have either 'run' or 'commands' attribute"
	_, err := ParseConfig([]byte(content))
	actual := fmt.Sprintf("%s", err)
	assertEqual(t, expected, actual, "ParseConfig() returned unexpected error")
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
	if err != nil {
		t.Fatalf("ParseConfig() returned unexpected error: %v", err)
	}
	assertDeepEqual(t, expected, actual, "ParseConfig() returned unexpected error")
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
	if err != nil {
		t.Fatalf("ParseConfig() returned unexpected error: %v", err)
	}
	assertDeepEqual(t, expected, actual, "ParseConfig() returned unexpected error")
}

func TestParseConfig_InputOptionsIsScaler(t *testing.T) {
	content := `
run: ok
inputs:
  foobar:
    options: ooops
`
	expected := "line 5: unexpected node type"
	_, err := ParseConfig([]byte(content))
	actual := fmt.Sprintf("%s", err)
	assertEqual(t, expected, actual, "ParseConfig() returned unexpected error")
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
	if err != nil {
		t.Fatalf("ParseConfig() returned unexpected error: %v", err)
	}
	assertDeepEqual(t, expected, actual, "ParseConfig() returned unexpected error")
}

func TestParseConfig_InputsIsSequence(t *testing.T) {
	content := `
run: ooops
inputs: []
`
	expected := "line 3: cannot unmarshal inputs into map"
	_, err := ParseConfig([]byte(content))
	actual := fmt.Sprintf("%s", err)
	assertEqual(t, expected, actual, "ParseConfig() returned unexpected error")
}

func TestParseConfig_InputsIsScalar(t *testing.T) {
	content := `
run: ooops
inputs: nope
`
	expected := "line 3: cannot unmarshal inputs into map"
	_, err := ParseConfig([]byte(content))
	actual := fmt.Sprintf("%s", err)
	assertEqual(t, expected, actual, "ParseConfig() returned unexpected error")
}

func TestParseConfig_InputNames(t *testing.T) {
	content := `
run: ooops
inputs:
  _invalid:
    default: Nope
`
	expected := "line 5: invalid input name"
	_, err := ParseConfig([]byte(content))
	actual := fmt.Sprintf("%s", err)
	assertEqual(t, expected, actual, "ParseConfig() returned unexpected error")
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
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	if _, err := tempFile.Write([]byte(content)); err != nil {
		t.Fatalf("Failed to write config to temp file: %v", err)
	}

	actual, err := LoadConfig(tempFile.Name())
	if err != nil {
		t.Fatalf("LoadConfig() returned an unexpected error: %v", err)
	}

	assertDeepEqual(t, expected, actual, "LoadConfig() returned unexpected results")
}
