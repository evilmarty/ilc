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

func TestConfigCommandInputOptionsUnmarshalYAMLSequence(t *testing.T) {
	var actual ConfigCommandInputOptions
	content := `
- Megatron
- Soundwave
- Starscream
`
	expected := ConfigCommandInputOptions{
		"Megatron":   "Megatron",
		"Soundwave":  "Soundwave",
		"Starscream": "Starscream",
	}
	err := yaml.Unmarshal([]byte(content), &actual)

	if err != nil {
		t.Errorf("Received error from parser: %s", err)
	}

	if !reflect.DeepEqual(actual, expected) {
		fatalDiff(t, actual, expected)
	}
}

func TestConfigCommandInputOptionsUnmarshalYAMLMap(t *testing.T) {
	var actual ConfigCommandInputOptions
	content := `
Megatron: Decepticon
Optimus Prime: Autobot
Optimus Primal: Maximal
`
	expected := ConfigCommandInputOptions{
		"Megatron":       "Decepticon",
		"Optimus Prime":  "Autobot",
		"Optimus Primal": "Maximal",
	}
	err := yaml.Unmarshal([]byte(content), &actual)

	if err != nil {
		t.Errorf("Received error from parser: %s", err)
	}

	if !reflect.DeepEqual(actual, expected) {
		fatalDiff(t, expected, actual)
	}
}

func TestConfigCommandInputsUnmarshalYAML(t *testing.T) {
	var actual ConfigCommandInputs
	content := `
name:
city:
  default: Autobot City
`
	expected := ConfigCommandInputs{
		ConfigCommandInput{
			Name: "name",
		},
		ConfigCommandInput{
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

func TestConfigCommandInputSelectable_WithOptions(t *testing.T) {
	input := ConfigCommandInput{
		Options: ConfigCommandInputOptions{
			"a": "1",
			"b": "2",
		},
	}

	if !input.Selectable() {
		t.Error("Expected input to be selectable")
	}
}

func TestConfigCommandInputSelectable_WithoutOptions(t *testing.T) {
	input := ConfigCommandInput{}

	if input.Selectable() {
		t.Error("Expected input to not be selectable")
	}
}

func TestConfigCommandInputValid_WithOptions(t *testing.T) {
	input := ConfigCommandInput{
		Options: ConfigCommandInputOptions{
			"a": "1",
			"b": "2",
		},
	}

	if !input.Valid("1") {
		t.Fatal("Expected value '1' to be valid")
	}
	if input.Valid("3") {
		t.Fatal("Expected value '3' to be invalid")
	}
}

func TestConfigCommandInputValid_WithPattern(t *testing.T) {
	input := ConfigCommandInput{
		Pattern: "^[0-9]+$",
	}

	if !input.Valid("1") {
		t.Fatal("Expected value '1' to be valid")
	}
	if input.Valid("a") {
		t.Fatal("Expected value 'a' to be invalid")
	}
}

func TestConfigCommandInputValid_WithoutPattern(t *testing.T) {
	input := ConfigCommandInput{}
	values := []string{"a", " ", ""}

	for _, value := range values {
		if !input.Valid(value) {
			t.Fatalf("Expected value '%s' to be valid", value)
		}
	}
}

func TestConfigCommandsUnmarshalYAML(t *testing.T) {
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
