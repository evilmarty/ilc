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

	expected := []string{"FOOBAZ=123", "FOOBAR=123 456"}
	actual := model.env([]string{"FOOBAZ=123"})

	if !reflect.DeepEqual(actual, expected) {
		fatalDiff(t, expected, actual)
	}
}

func TestModelPureEnv(t *testing.T) {
	model := &model{
		commands: Commands{
			Command{
				Env:  map[string]string{"FOOBAR": "{{.a}} {{.b}}"},
				Pure: true,
			},
		},
		values: map[string]any{"a": "123", "b": 456},
	}

	expected := []string{"FOOBAR=123 456"}
	actual := model.env([]string{"FOOBAZ=123"})

	if !reflect.DeepEqual(actual, expected) {
		fatalDiff(t, expected, actual)
	}
}
