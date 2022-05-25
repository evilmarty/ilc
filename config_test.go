package main

import (
	"io/ioutil"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/kr/pretty"
)

func TestLoadConfig(t *testing.T) {
	content := `
commands:
  test:
    run: go test
`
	expected := &Config{
		Commands: Commands{
			Command{
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
