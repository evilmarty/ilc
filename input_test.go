package main

import (
	"reflect"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestInputOptionsUnmarshalYAMLSequence(t *testing.T) {
	var actual InputOptions
	content := `
- Megatron
- Soundwave
- Starscream
`
	expected := InputOptions{
		InputOption{Label: "Megatron", Value: "Megatron"},
		InputOption{Label: "Soundwave", Value: "Soundwave"},
		InputOption{Label: "Starscream", Value: "Starscream"},
	}
	err := yaml.Unmarshal([]byte(content), &actual)

	if err != nil {
		t.Errorf("Received error from parser: %s", err)
	}

	if !reflect.DeepEqual(actual, expected) {
		fatalDiff(t, actual, expected)
	}
}

func TestInputOptionsUnmarshalYAMLMap(t *testing.T) {
	var actual InputOptions
	content := `
Megatron: Decepticon
Optimus Prime: Autobot
Optimus Primal: Maximal
`
	expected := InputOptions{
		InputOption{Label: "Megatron", Value: "Decepticon"},
		InputOption{Label: "Optimus Prime", Value: "Autobot"},
		InputOption{Label: "Optimus Primal", Value: "Maximal"},
	}
	err := yaml.Unmarshal([]byte(content), &actual)

	if err != nil {
		t.Errorf("Received error from parser: %s", err)
	}

	if !reflect.DeepEqual(actual, expected) {
		fatalDiff(t, expected, actual)
	}
}

func TestInputsUnmarshalYAML(t *testing.T) {
	var actual Inputs
	content := `
name:
city:
  default: Autobot City
`
	expected := Inputs{
		Input{
			Name: "name",
		},
		Input{
			Name:         "city",
			DefaultValue: "Autobot City",
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

func TestInputSelectable_WithOptions(t *testing.T) {
	input := Input{
		Options: InputOptions{
			InputOption{Label: "a", Value: "1"},
			InputOption{Label: "b", Value: "2"},
		},
	}

	if !input.Selectable() {
		t.Error("Expected input to be selectable")
	}
}

func TestInputSelectable_WithoutOptions(t *testing.T) {
	input := Input{}

	if input.Selectable() {
		t.Error("Expected input to not be selectable")
	}
}

func TestInputValid_WithOptions(t *testing.T) {
	input := Input{
		Options: InputOptions{
			InputOption{Label: "a", Value: "1"},
			InputOption{Label: "b", Value: "2"},
		},
	}

	if !input.Valid("1") {
		t.Fatal("Expected value '1' to be valid")
	}
	if input.Valid("3") {
		t.Fatal("Expected value '3' to be invalid")
	}
}

func TestInputValid_WithPattern(t *testing.T) {
	input := Input{
		Pattern: "^[0-9]+$",
	}

	if !input.Valid("1") {
		t.Fatal("Expected value '1' to be valid")
	}
	if input.Valid("a") {
		t.Fatal("Expected value 'a' to be invalid")
	}
}

func TestInputValid_WithoutPattern(t *testing.T) {
	input := Input{}
	values := []string{"a", " ", ""}

	for _, value := range values {
		if !input.Valid(value) {
			t.Fatalf("Expected value '%s' to be valid", value)
		}
	}
}
