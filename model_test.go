package main

import (
	"reflect"
	"testing"
)

func TestModelEnv(t *testing.T) {
	model := &model{
		commands: Commands{
			Command{
				Env: map[string]string{"FOOBAR": "{{.a}} {{.b}}"},
			},
		},
		values: map[string]any{"a": "123", "b": 456},
	}

	expected := []string{"FOOBAR=123 456"}
	actual := model.env()

	if !reflect.DeepEqual(actual, expected) {
		fatalDiff(t, expected, actual)
	}
}

func TestModelShell_Default(t *testing.T) {
	model := &model{
		config: Config{},
	}
	actual := model.shell()

	if !reflect.DeepEqual(actual, DefaultShell) {
		fatalDiff(t, DefaultShell, actual)
	}
}

func TestModelShell_Custom(t *testing.T) {
	expected := []string{"foo", "bar"}
	model := &model{
		config: Config{
			Shell: expected,
		},
	}
	actual := model.shell()

	if !reflect.DeepEqual(actual, expected) {
		fatalDiff(t, DefaultShell, actual)
	}
}
