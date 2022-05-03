package main

import (
	"reflect"
	"testing"
)

func TestModelShell_Default(t *testing.T) {
	model := &model{
		config: &Config{},
	}
	actual := model.shell()

	if !reflect.DeepEqual(actual, DefaultShell) {
		fatalDiff(t, DefaultShell, actual)
	}
}

func TestModelShell_Custom(t *testing.T) {
	expected := []string{"foo", "bar"}
	model := &model{
		config: &Config{
			Shell: expected,
		},
	}
	actual := model.shell()

	if !reflect.DeepEqual(actual, expected) {
		fatalDiff(t, DefaultShell, actual)
	}
}
