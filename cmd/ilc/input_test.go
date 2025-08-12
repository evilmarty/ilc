package main

import (
	"os"
	"reflect"
	"testing"
)

func TestGetInitialInputValues(t *testing.T) {
	inputs := map[string]Input{
		"name": {
			Default: "World",
		},
		"envvar": {},
	}
	values := make(map[string]any)

	// Test case 1: Use default value
	getInitialInputValues(inputs, values)
	if values["name"] != "World" {
		t.Errorf("expected name 'World', got '%s'", values["name"])
	}

	// Test case 2: Get value from env var
	os.Setenv("ILC_INPUT_envvar", "from_env")
	defer os.Unsetenv("ILC_INPUT_envvar")
	getInitialInputValues(inputs, values)
	if values["envvar"] != "from_env" {
		t.Errorf("expected envvar 'from_env', got '%s'", values["envvar"])
	}
}

func TestParseInputArgs(t *testing.T) {
	values := make(map[string]any)
	args := []string{"-greeting", "Hi"}
	parseInputArgs(args, values)
	if values["greeting"] != "Hi" {
		t.Errorf("expected greeting 'Hi', got '%s'", values["greeting"])
	}
}

func TestGetOptions(t *testing.T) {
	input := Input{
		Options: []any{"a", "b", map[string]any{"c": "d"}},
	}
	expected := []string{"a", "b", "c"}
	result := getOptions(input)
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("expected %v, got %v", expected, result)
	}
}
