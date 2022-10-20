package main

import (
	"testing"
)

func TestRenderTemplate(t *testing.T) {
	template := "{{ .foobar }} {{ .foobaz }}"
	data := map[string]any{"foobar": "foobar"}
	expected := "foobar <no value>"
	actual, err := RenderTemplate(template, data)

	if err != nil {
		t.Errorf("Error rendering template: %s", err)
	}

	if actual != expected {
		t.Fatalf("Expected result to be '%s', not '%s'", expected, actual)
	}
}
