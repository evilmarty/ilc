package main

import (
	"io/ioutil"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/kr/pretty"
	"gopkg.in/yaml.v3"
)

func TestConfigCommandHasSubCommands(t *testing.T) {
	command := ConfigCommand{
		Name: "decepticons",
		Commands: ConfigCommands{
			ConfigCommand{Name: "punish"},
			ConfigCommand{Name: "enslave"},
		},
	}
	if !command.HasSubCommands() {
		t.Errorf("Expected to have sub commands")
	}
}

func TestConfigCommandsUnmarshalYAML_WithRunAndCommands(t *testing.T) {
	var commands ConfigCommands
	content := `
protect:
  run: echo fail
  commands:
    foobar:
      run: echo foobar
punish:
  run: echo Punish and enslave
`
	expected := "line 2: 'protect' command cannot have both run and commands attribute"
	actual := yaml.Unmarshal([]byte(content), &commands)

	if expected != actual.Error() {
		t.Errorf("Expected error, but received: %s", actual)
	}
}

func TestConfigCommandsUnmarshalYAML_WithoutRunAndCommands(t *testing.T) {
	var commands ConfigCommands
	content := `
protect: {}
punish:
  run: echo Punish and enslave
`
	expected := "line 2: 'protect' command missing run or commands attribute"
	actual := yaml.Unmarshal([]byte(content), &commands)

	if expected != actual.Error() {
		t.Errorf("Expected error, but received: %s", actual)
	}
}

func TestConfigCommandsUnmarshalYAML_WithCommands(t *testing.T) {
	var actual ConfigCommands
	content := `
protect:
  commands:
    foobar:
      run: echo foobar
punish:
  commands:
    foobaz:
      run: echo foobaz
`
	expected := ConfigCommands{
		ConfigCommand{
			Name: "protect",
			Commands: ConfigCommands{
				ConfigCommand{
					Name: "foobar",
					Run:  "echo foobar",
				},
			},
		},
		ConfigCommand{
			Name: "punish",
			Commands: ConfigCommands{
				ConfigCommand{
					Name: "foobaz",
					Run:  "echo foobaz",
				},
			},
		},
	}
	err := yaml.Unmarshal([]byte(content), &actual)

	if err != nil {
		t.Errorf("Received error from parser: %s", err)
	}

	if !reflect.DeepEqual(actual, expected) {
		fatalDiff(t, expected, actual)
	}
}

func TestConfigCommandsUnmarshalYAML_WithRun(t *testing.T) {
	var actual ConfigCommands
	content := `
protect:
  run: echo Protect all sentient life forms
punish:
  run: echo Punish and enslave
`
	expected := ConfigCommands{
		ConfigCommand{
			Name: "protect",
			Run:  "echo Protect all sentient life forms",
		},
		ConfigCommand{
			Name: "punish",
			Run:  "echo Punish and enslave",
		},
	}
	err := yaml.Unmarshal([]byte(content), &actual)

	if err != nil {
		t.Errorf("Received error from parser: %s", err)
	}

	if !reflect.DeepEqual(actual, expected) {
		fatalDiff(t, expected, actual)
	}
}

func TestConfigCommandsUnmarshalYAML_Shorthand(t *testing.T) {
	var actual ConfigCommands
	content := `
protect: echo Protect all sentient life forms
punish: echo Punish and enslave
`
	expected := ConfigCommands{
		ConfigCommand{
			Name: "protect",
			Run:  "echo Protect all sentient life forms",
		},
		ConfigCommand{
			Name: "punish",
			Run:  "echo Punish and enslave",
		},
	}
	err := yaml.Unmarshal([]byte(content), &actual)

	if err != nil {
		t.Errorf("Received error from parser: %s", err)
	}

	if !reflect.DeepEqual(actual, expected) {
		fatalDiff(t, expected, actual)
	}
}

func TestConfigCommandsGet(t *testing.T) {
	commands := ConfigCommands{
		ConfigCommand{
			Name: "a",
		},
		ConfigCommand{
			Name: "b",
		},
	}

	expected := &commands[1]
	actual := commands.Get("b")
	if !reflect.DeepEqual(actual, expected) {
		fatalDiff(t, expected, actual)
	}

	expected = nil
	actual = commands.Get("c")
	if !reflect.DeepEqual(actual, expected) {
		fatalDiff(t, expected, actual)
	}
}

func TestConfigCommandsInputs(t *testing.T) {
	input1 := Input{}
	input2 := Input{}
	commands := ConfigCommands{
		ConfigCommand{
			Name:   "a",
			Inputs: Inputs{input1},
		},
		ConfigCommand{
			Name:   "b",
			Inputs: Inputs{input2},
		},
	}

	expected := Inputs{input1, input2}
	actual := commands.Inputs()
	if !reflect.DeepEqual(actual, expected) {
		fatalDiff(t, expected, actual)
	}
}

func TestLoadConfig(t *testing.T) {
	content := `
commands:
  test:
    run: go test
`
	expected := &Config{
		Commands: ConfigCommands{
			ConfigCommand{
				Name: "test",
				Run:  "go test",
			},
		},
	}
	tempFile := filepath.Join(t.TempDir(), "ilc.yml")

	if err := ioutil.WriteFile(tempFile, []byte(content), 0644); err != nil {
		t.Errorf("Failed to write temp file: %s", err)
	}

	actual, err := LoadConfig(tempFile)

	if err != nil {
		t.Errorf("Error loading config: %s", err)
	}

	if !reflect.DeepEqual(actual, expected) {
		fatalDiff(t, expected, actual)
	}
}

func fatalDiff(t *testing.T, expected, actual interface{}) {
	t.Helper()
	b := strings.Builder{}
	pretty.Fdiff(&b, expected, actual)
	t.Fatal(b.String())
}
