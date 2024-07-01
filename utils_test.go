package main

import (
	"testing"
	"text/template"

	"github.com/stretchr/testify/assert"
)

func TestEnvMap(t *testing.T) {
	env := []string{
		"A=a",
		"B=",
		"C",
	}
	expected := map[string]string{"A": "a", "B": "", "C": ""}
	actual := EnvMap(env)
	assert.Equal(t, expected, actual, "EnvMap() returned unexpected results")
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
	assert.Equal(t, expected, actual, "NewTemplateData() returned unexpected results")
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
		assert.NoError(t, err, "RenderTemplate() returned unexpected error")
		assert.Equal(t, expected, actual, "RenderTemplate() returned unexpected results")
	})

	t.Run("given a template object", func(t *testing.T) {
		tmpl := template.New("")
		_, err := tmpl.Parse(text)
		assert.NoError(t, err, "Could not render template")
		actual, err := RenderTemplate(tmpl, data)
		assert.NoError(t, err, "RenderTemplate() returned unexpected error")
		assert.Equal(t, expected, actual, "RenderTemplate() returned unexpected results")
	})

	t.Run("given other", func(t *testing.T) {
		_, actual := RenderTemplate(nil, data)
		assert.EqualError(t, actual, "unsupported type: <nil>", "RenderTemplate() returned unexpected error")
	})
}

func TestDiffStrings(t *testing.T) {
	a := []string{"a", "b", "c"}
	b := []string{"a"}
	expected := []string{"b", "c"}
	assert.Equal(t, expected, DiffStrings(a, b), "DiffStrings() returned unexpected results")
}
