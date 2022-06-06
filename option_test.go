package main

import (
	"os"
	"reflect"
	"testing"

	"golang.org/x/sys/unix"
	"gopkg.in/yaml.v3"
)

func TestOptionsEmpty(t *testing.T) {
	hasItems := Options{
		Items: []Option{
			Option{Label: "One", Value: "1"},
			Option{Label: "Two", Value: "2"},
		},
	}
	hasScript := Options{Script: "seq 5"}
	isEmpty := Options{}

	if hasItems.Empty() {
		t.Errorf("Expected options with items to not be empty")
	}
	if hasScript.Empty() {
		t.Errorf("Expected options with script to not be empty")
	}
	if !isEmpty.Empty() {
		t.Errorf("Expected empty options to be empty")
	}
}

func TestOptionsContains(t *testing.T) {
	options := Options{
		Items: []Option{
			Option{Label: "One", Value: "1"},
			Option{Label: "Two", Value: "2"},
		},
	}
	if !options.Contains("1") {
		t.Errorf("Expected options to contain value: 1")
	}
	if options.Contains("3") {
		t.Errorf("Expected options to not contain value: 3")
	}
}

func TestOptionsGet_Items(t *testing.T) {
	option1 := Option{Label: "One", Value: "1"}
	option2 := Option{Label: "Two", Value: "2"}
	options := Options{
		Items: []Option{option1, option2},
	}
	expected := []Option{option1, option2}
	actual, err := options.Get()

	if err != nil {
		t.Errorf("Unexpected error: %s", err)
	}

	if !reflect.DeepEqual(actual, expected) {
		fatalDiff(t, actual, expected)
	}
}

func TestOptionsGet_Script(t *testing.T) {
	if !isatty() {
		t.Skipf("No TTY present")
	}

	options := Options{
		Script: "seq 3",
	}
	expected := []Option{
		Option{Label: "1", Value: "1"},
		Option{Label: "2", Value: "2"},
		Option{Label: "3", Value: "3"},
	}
	actual, err := options.Get()

	if err != nil {
		t.Errorf("Unexpected error: %s", err)
	}

	if !reflect.DeepEqual(actual, expected) {
		fatalDiff(t, actual, expected)
	}
}

func TestOptionsUnmarshalYAML_Sequence(t *testing.T) {
	var actual Options
	content := `
- Megatron
- Soundwave
- Starscream
`
	expected := Options{
		Items: []Option{
			Option{Label: "Megatron", Value: "Megatron"},
			Option{Label: "Soundwave", Value: "Soundwave"},
			Option{Label: "Starscream", Value: "Starscream"},
		},
	}
	err := yaml.Unmarshal([]byte(content), &actual)

	if err != nil {
		t.Errorf("Received error from parser: %s", err)
	}

	if !reflect.DeepEqual(actual, expected) {
		fatalDiff(t, actual, expected)
	}
}

func TestOptionsUnmarshalYAML_Map(t *testing.T) {
	var actual Options
	content := `
Megatron: Decepticon
Optimus Prime: Autobot
Optimus Primal: Maximal
`
	expected := Options{
		Items: []Option{
			Option{Label: "Megatron", Value: "Decepticon"},
			Option{Label: "Optimus Prime", Value: "Autobot"},
			Option{Label: "Optimus Primal", Value: "Maximal"},
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

func TestOptionsUnmarshalYAML_Scalar(t *testing.T) {
	var actual Options
	content := `echo roll out`
	expected := Options{
		Script: "echo roll out",
	}
	err := yaml.Unmarshal([]byte(content), &actual)

	if err != nil {
		t.Errorf("Received error from parser: %s", err)
	}

	if !reflect.DeepEqual(actual, expected) {
		fatalDiff(t, expected, actual)
	}
}

func isatty() bool {
	_, err := unix.IoctlGetWinsize(int(os.Stdout.Fd()), unix.TIOCGWINSZ)
	return err == nil
}
