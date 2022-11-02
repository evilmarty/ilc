package main

import (
	"reflect"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestCommandHasSubCommands(t *testing.T) {
	command := Command{
		Name: "decepticons",
		Commands: Commands{
			Command{Name: "punish"},
			Command{Name: "enslave"},
		},
	}
	if !command.HasSubCommands() {
		t.Errorf("Expected to have sub commands")
	}
}

func TestCommandsUnmarshalYAML_WithRunAndCommands(t *testing.T) {
	var commands Commands
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

func TestCommandsUnmarshalYAML_WithoutRunAndCommands(t *testing.T) {
	var commands Commands
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

func TestCommandsUnmarshalYAML_WithCommands(t *testing.T) {
	var actual Commands
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
	expected := Commands{
		Command{
			Name: "protect",
			Commands: Commands{
				Command{
					Name: "foobar",
					Run:  "echo foobar",
				},
			},
		},
		Command{
			Name: "punish",
			Commands: Commands{
				Command{
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

func TestCommandsUnmarshalYAML_WithRun(t *testing.T) {
	var actual Commands
	content := `
protect:
  run: echo Protect all sentient life forms
punish:
  run: echo Punish and enslave
`
	expected := Commands{
		Command{
			Name: "protect",
			Run:  "echo Protect all sentient life forms",
		},
		Command{
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

func TestCommandsUnmarshalYAML_Shorthand(t *testing.T) {
	var actual Commands
	content := `
protect: echo Protect all sentient life forms
punish: echo Punish and enslave
`
	expected := Commands{
		Command{
			Name: "protect",
			Run:  "echo Protect all sentient life forms",
		},
		Command{
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

func TestCommandsGet(t *testing.T) {
	commands := Commands{
		Command{
			Name: "a",
		},
		Command{
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
