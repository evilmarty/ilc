package main

import (
	"math"
	"slices"
	"testing"
	"text/template"

	"github.com/stretchr/testify/assert"
)

func TestEnvMapMerge(t *testing.T) {
	em1 := EnvMap{"A": "a", "B": "b"}
	em2 := EnvMap{"B": "bb", "C": "c"}
	expected := EnvMap{"A": "a", "B": "bb", "C": "c"}
	actual := em1.Merge(em2)
	assert.Equal(t, expected, actual)
}

func TestEnvMapPrefix(t *testing.T) {
	em := EnvMap{"BAR": "foobar", "BAZ": "foobaz"}
	expected := EnvMap{"FOOBAR": "foobar", "FOOBAZ": "foobaz"}
	actual := em.Prefix("FOO")
	assert.Equal(t, expected, actual)
}

func TestEnvMapTrimPrefix(t *testing.T) {
	em := EnvMap{"FOOBAR": "foobar", "FOOBAZ": "foobaz"}
	expected := EnvMap{"BAR": "foobar", "BAZ": "foobaz"}
	actual := em.TrimPrefix("FOO")
	assert.Equal(t, expected, actual)
}

func TestEnvMapFilterPrefix(t *testing.T) {
	em := EnvMap{"FOO_BAR": "a", "FOO_BAZ": "b", "FOOBAR": "c"}
	expected := EnvMap{"FOO_BAR": "a", "FOO_BAZ": "b"}
	actual := em.FilterPrefix("FOO_")
	assert.Equal(t, expected, actual)
}

func TestEnvMapToList(t *testing.T) {
	em := EnvMap{"FOOBAR": "foobar", "FOOBAZ": "foobaz"}
	expected := []string{"FOOBAR=foobar", "FOOBAZ=foobaz"}
	actual := em.ToList()
	slices.Sort(expected)
	slices.Sort(actual)
	assert.Equal(t, expected, actual)
}

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
