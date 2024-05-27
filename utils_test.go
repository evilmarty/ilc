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

	assertEqual(t, expected, actual, "RenderTemplate() returned unexpected results")
}

func TestDiffStrings(t *testing.T) {
	a := []string{"a", "b", "c"}
	b := []string{"a"}
	expected := []string{"b", "c"}
	assertDeepEqual(t, expected, DiffStrings(a, b), "DiffStrings() returned unexpected results")
}
