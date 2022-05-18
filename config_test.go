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
