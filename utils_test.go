package main

import (
	"math"
	"testing"
	"text/template"

	"github.com/stretchr/testify/assert"
)

func TestNewEnvMap(t *testing.T) {
	env := []string{
		"A=a",
		"B=",
		"C",
	}
	expected := EnvMap{"A": "a", "B": "", "C": ""}
	actual := NewEnvMap(env)
	assert.Equal(t, expected, actual, "NewEnvMap() returned unexpected results")
}

func TestNewTemplateData(t *testing.T) {
	inputs := map[string]any{
		"foo_bar": "foobar",
		"foo-baz": "foobaz",
	}
	env := map[string]string{"A": "a", "B": "b"}
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
	data := TemplateData{
		Input: map[string]any{"foobar": "a"},
		Env:   map[string]string{"FOOBAR": "b"},
	}
	expected := "Input: a, Input: <no value>, Env: b, Env: <no value>"
	t.Run("given a string", func(t *testing.T) {
		text := "Input: {{input \"foobar\"}}, Input: {{input \"foobaz\"}}, Env: {{env \"FOOBAR\"}}, Env: {{env \"FOOBAZ\"}}"
		actual, err := RenderTemplate(text, data)
		assert.NoError(t, err, "RenderTemplate() returned unexpected error")
		assert.Equal(t, expected, actual, "RenderTemplate() returned unexpected results")
	})

	t.Run("given a template object", func(t *testing.T) {
		text := "Input: {{.Input.foobar}}, Input: {{.Input.foobaz}}, Env: {{.Env.FOOBAR}}, Env: {{.Env.FOOBAZ}}"
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

func TestToFloat64(t *testing.T) {
	t.Run("int", func(t *testing.T) {
		assert.Equal(t, float64(1), ToFloat64(int(1)), "toFloat64() returned unexpected result")
	})
	t.Run("uint", func(t *testing.T) {
		assert.Equal(t, float64(1), ToFloat64(uint(1)), "toFloat64() returned unexpected result")
	})
	t.Run("int16", func(t *testing.T) {
		assert.Equal(t, float64(1), ToFloat64(int16(1)), "toFloat64() returned unexpected result")
	})
	t.Run("uint16", func(t *testing.T) {
		assert.Equal(t, float64(1), ToFloat64(uint16(1)), "toFloat64() returned unexpected result")
	})
	t.Run("int32", func(t *testing.T) {
		assert.Equal(t, float64(1), ToFloat64(int32(1)), "toFloat64() returned unexpected result")
	})
	t.Run("uint32", func(t *testing.T) {
		assert.Equal(t, float64(1), ToFloat64(uint32(1)), "toFloat64() returned unexpected result")
	})
	t.Run("int64", func(t *testing.T) {
		assert.Equal(t, float64(1), ToFloat64(int64(1)), "toFloat64() returned unexpected result")
	})
	t.Run("uint64", func(t *testing.T) {
		assert.Equal(t, float64(1), ToFloat64(uint64(1)), "toFloat64() returned unexpected result")
	})
	t.Run("float32", func(t *testing.T) {
		assert.Equal(t, float64(1), ToFloat64(float32(1)), "toFloat64() returned unexpected result")
	})
	t.Run("float64", func(t *testing.T) {
		assert.Equal(t, float64(1), ToFloat64(float64(1)), "toFloat64() returned unexpected result")
	})
	t.Run("string", func(t *testing.T) {
		assert.True(t, math.IsNaN(ToFloat64("1")), "toFloat64() returned unexpected result")
	})
}
