package main

import (
	"fmt"
	"testing"
	"text/template"
)

func TestEnvMap(t *testing.T) {
	env := []string{
		"A=a",
		"B=",
		"C",
	}
	expected := map[string]string{"A": "a", "B": "", "C": ""}
	actual := EnvMap(env)
	assertDeepEqual(t, expected, actual, "EnvMap() returned unexpected results")
}

func TestNewTemplateData(t *testing.T) {
	inputs := map[string]any{
		"foo_bar": "foobar",
		"foo-baz": "foobaz",
	}
	env := []string{"A=a", "B=b"}
	expected := TemplateData{
		Input: map[string]any{
			"foo_bar": "foobar",
			"foo_baz": "foobaz",
		},
		Env: map[string]string{
			"A": "a",
			"B": "b",
		},
	}
	actual := NewTemplateData(inputs, env)
	assertDeepEqual(t, expected, actual, "NewTemplateData() returned unexpected results")
}

func TestRenderTemplate(t *testing.T) {
	text := "Input: {{ .Input.foobar }}, Input: {{ .Input.foobaz }}, Env: {{ .Env.FOOBAR }}, Env: {{ .Env.FOOBAZ }}"
	data := TemplateData{
		Input: map[string]any{"foobar": "a"},
		Env:   map[string]string{"FOOBAR": "b"},
	}
	expected := "Input: a, Input: <no value>, Env: b, Env: <no value>"
	t.Run("given a string", func(t *testing.T) {
		actual, err := RenderTemplate(text, data)
		if err != nil {
			t.Fatalf("RenderTemplate() returned unexpected error: %v", err)
		}

		assertEqual(t, expected, actual, "RenderTemplate() returned unexpected results")
	})

	t.Run("given a template object", func(t *testing.T) {
		tmpl := template.New("")
		if _, err := tmpl.Parse(text); err != nil {
			t.Fatalf("Could not render template: %v", err)
		}
		actual, err := RenderTemplate(tmpl, data)
		if err != nil {
			t.Fatalf("RenderTemplate() returned unexpected error: %v", err)
		}

		assertEqual(t, expected, actual, "RenderTemplate() returned unexpected results")
	})

	t.Run("given other", func(t *testing.T) {
		_, err := RenderTemplate(nil, data)
		actual := fmt.Sprintf("%s", err)
		assertEqual(t, "unsupported type: <nil>", actual, "RenderTemplate() returned unexpected error")
	})
}

func TestDiffStrings(t *testing.T) {
	a := []string{"a", "b", "c"}
	b := []string{"a"}
	expected := []string{"b", "c"}
	assertDeepEqual(t, expected, DiffStrings(a, b), "DiffStrings() returned unexpected results")
}
