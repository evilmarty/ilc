package main

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInputOptionsContains(t *testing.T) {
	options := InputOptions{
		{"A", "a"},
		{"B", "b"},
	}
	assert.True(t, options.Contains("a"), "InputOptions.Contains() returned unexpected result")
	assert.False(t, options.Contains("c"), "InputOptions.Contains() returned unexpected result")
}

func TestStringValueKind(t *testing.T) {
	i := StringValue{}
	assert.Equal(t, reflect.String, i.Kind())
}

func TestStringValueString(t *testing.T) {
	i := StringValue{Value: "foobar"}
	assert.Equal(t, "foobar", i.String())
}

func TestStringValueGet(t *testing.T) {
	i := StringValue{Value: "foobar"}
	assert.Equal(t, "foobar", i.Get())
}

func TestStringValueSet(t *testing.T) {
	t.Run("no validation", func(t *testing.T) {
		i := StringValue{}
		assert.NoError(t, i.Set("foobar"))
		assert.Equal(t, "foobar", i.Value)
	})
	t.Run("with pattern", func(t *testing.T) {
		i := StringValue{Pattern: "bar$"}
		assert.NoError(t, i.Set("foobar"))
		assert.Equal(t, "foobar", i.Value)
		assert.Error(t, i.Set("foobaz"))
		assert.Equal(t, "foobar", i.Value)
	})
}

func TestNumberValueKind(t *testing.T) {
	i := NumberValue{}
	assert.Equal(t, reflect.Float64, i.Kind())
}

func TestNumberValueString(t *testing.T) {
	t.Run("with decimals", func(t *testing.T) {
		i := NumberValue{Value: 1.23456789}
		assert.Equal(t, "1.23457", i.String())
	})
	t.Run("without decimals", func(t *testing.T) {
		i := NumberValue{Value: 1.0}
		assert.Equal(t, "1", i.String())
	})
}

func TestNumberValueGet(t *testing.T) {
	i := NumberValue{Value: 1.2}
	assert.Equal(t, 1.2, i.Get())
}

func TestNumberValueSet(t *testing.T) {
	t.Run("no validation", func(t *testing.T) {
		i := NumberValue{}
		assert.NoError(t, i.Set("1.2"))
		assert.Equal(t, 1.2, i.Value)
	})
	t.Run("with min and max", func(t *testing.T) {
		i := NumberValue{MinValue: 1.0, MaxValue: 5.0}
		assert.NoError(t, i.Set("2.0"))
		assert.Equal(t, 2.0, i.Value)
		assert.Error(t, i.Set("0"))
		assert.Equal(t, 2.0, i.Value)
		assert.Error(t, i.Set("5.1"))
		assert.Equal(t, 2.0, i.Value)
	})
	t.Run("with min only", func(t *testing.T) {
		i := NumberValue{MinValue: 1.0}
		assert.NoError(t, i.Set("2.0"))
		assert.Equal(t, 2.0, i.Value)
		assert.Error(t, i.Set("0"))
		assert.Equal(t, 2.0, i.Value)
	})
	t.Run("with max only", func(t *testing.T) {
		i := NumberValue{MaxValue: 1.0}
		assert.NoError(t, i.Set("1.0"))
		assert.Equal(t, 1.0, i.Value)
		assert.Error(t, i.Set("2"))
		assert.Equal(t, 1.0, i.Value)
	})
}

func TestBooleanValueKind(t *testing.T) {
	i := BooleanValue{}
	assert.Equal(t, reflect.Bool, i.Kind())
}

func TestBooleanValueString(t *testing.T) {
	i := BooleanValue{Value: true}
	assert.Equal(t, "true", i.String())
}

func TestBooleanValueGet(t *testing.T) {
	i := BooleanValue{Value: true}
	assert.Equal(t, true, i.Get())
}

func TestBooleanValueSet(t *testing.T) {
	t.Run("no validation", func(t *testing.T) {
		i := BooleanValue{}
		assert.NoError(t, i.Set("true"))
		assert.Equal(t, true, i.Value)
	})
}

func TestInputEnvName(t *testing.T) {
	input := Input{Name: "foo-bar"}
	assert.Equal(t, "foo_bar", input.EnvName())
}

func TestInputsGet(t *testing.T) {
	inputs := Inputs{
		Input{Name: "foobar", Value: &NumberValue{Value: 123}},
	}
	assert.Equal(t, float64(123), inputs.Get("foobar"))
	assert.Equal(t, nil, inputs.Get("foobaz"))
}

func TestInputsGetAll(t *testing.T) {
	t.Run("empty", func(t *testing.T) {
		inputs := Inputs{}
		expected := map[string]any{}
		actual := inputs.GetAll()
		assert.Equal(t, expected, actual)
	})
	t.Run("with inputs", func(t *testing.T) {
		inputs := Inputs{
			Input{Name: "foobar", Value: &NumberValue{Value: 123}},
			Input{Name: "foobaz", Value: &StringValue{Value: "foobar"}},
		}
		expected := map[string]any{
			"foobar": float64(123),
			"foobaz": "foobar",
		}
		actual := inputs.GetAll()
		assert.Equal(t, expected, actual)
	})
}

func TestInputsSet(t *testing.T) {
	t.Run("empty", func(t *testing.T) {
		inputs := Inputs{}
		err := inputs.Set("foobar", "456")
		assert.NoError(t, err)
	})
	t.Run("valid", func(t *testing.T) {
		inputs := Inputs{
			Input{Name: "foobar", Value: &NumberValue{Value: 123}},
		}
		expected := float64(456)
		err := inputs.Set("foobar", "456")
		actual := inputs.Get("foobar")
		assert.NoError(t, err)
		assert.Equal(t, expected, actual)
	})
	t.Run("invalid", func(t *testing.T) {
		inputs := Inputs{
			Input{Name: "foobar", Value: &NumberValue{Value: 123}},
		}
		err := inputs.Set("foobar", "foobar")
		assert.Error(t, err)
	})
}

func TestInputsSetAll(t *testing.T) {
	t.Run("empty", func(t *testing.T) {
		inputs := Inputs{}
		err := inputs.SetAll(map[string]string{"foobar": "123"})
		assert.NoError(t, err)
	})
	t.Run("all valid", func(t *testing.T) {
		inputs := Inputs{
			Input{Name: "foobar", Value: &NumberValue{Value: 123}},
			Input{Name: "foobaz", Value: &StringValue{Value: "foobar"}},
		}
		expected := map[string]any{
			"foobar": float64(456),
			"foobaz": "foobaz",
		}
		err := inputs.SetAll(map[string]string{
			"foobar": "456",
			"foobaz": "foobaz",
			"other":  "ignored",
		})
		actual := inputs.GetAll()
		assert.NoError(t, err)
		assert.Equal(t, expected, actual)
	})
	t.Run("one invalid", func(t *testing.T) {
		inputs := Inputs{
			Input{Name: "foobar", Value: &NumberValue{Value: 123}},
			Input{Name: "foobaz", Value: &StringValue{Value: "foobar"}},
		}
		expected := map[string]any{
			"foobar": float64(123),
			"foobaz": "foobaz",
		}
		err := inputs.SetAll(map[string]string{
			"foobar": "ooops",
			"foobaz": "foobaz",
			"other":  "ignored",
		})
		actual := inputs.GetAll()
		assert.ErrorContains(t, err, "strconv.ParseFloat")
		assert.Equal(t, expected, actual)
	})
	t.Run("all invalid", func(t *testing.T) {
		inputs := Inputs{
			Input{Name: "foobar", Value: &NumberValue{Value: 123}},
			Input{Name: "foobaz", Value: &BooleanValue{Value: false}},
		}
		expected := map[string]any{
			"foobar": float64(123),
			"foobaz": false,
		}
		err := inputs.SetAll(map[string]string{
			"foobar": "ooops",
			"foobaz": "foobaz",
			"other":  "ignored",
		})
		actual := inputs.GetAll()
		assert.ErrorContains(t, err, "strconv.ParseFloat")
		assert.ErrorContains(t, err, "strconv.ParseBool")
		assert.Equal(t, expected, actual)
	})
}

func TestInputsHas(t *testing.T) {
	inputs := Inputs{
		Input{Name: "foobar"},
	}
	assert.True(t, inputs.Has("foobar"))
	assert.False(t, inputs.Has("foobaz"))
}

func TestInputsMerge(t *testing.T) {
	inputs1 := Inputs{
		Input{Name: "a", Value: &StringValue{}},
		Input{Name: "b", Value: &StringValue{}},
	}
	inputs2 := Inputs{
		Input{Name: "a", Value: &NumberValue{}},
		Input{Name: "c", Value: &NumberValue{}},
	}
	expected := Inputs{
		Input{Name: "a", Value: &NumberValue{}},
		Input{Name: "b", Value: &StringValue{}},
		Input{Name: "c", Value: &NumberValue{}},
	}
	actual := inputs1.Merge(inputs2)
	assert.Equal(t, expected, actual)
}

func TestInputsFlagSet(t *testing.T) {
	inputs := Inputs{
		Input{Name: "s", Description: "about s", Value: &StringValue{}},
		Input{Name: "n", Description: "about n", Value: &NumberValue{}},
		Input{Name: "b", Description: "about b", Value: &BooleanValue{}},
	}
	fs := inputs.FlagSet()
	f := fs.Lookup("s")
	assert.NotNil(t, f)
	assert.Equal(t, f.Name, "s")
	assert.Equal(t, f.Usage, "about s")
	f = fs.Lookup("n")
	assert.NotNil(t, f)
	assert.Equal(t, f.Name, "n")
	assert.Equal(t, f.Usage, "about n")
	f = fs.Lookup("b")
	assert.NotNil(t, f)
	assert.Equal(t, f.Name, "b")
	assert.Equal(t, f.Usage, "about b")
}

func TestInputsToEnvMap(t *testing.T) {
	inputs := Inputs{
		Input{Name: "s", Value: &StringValue{Value: "s"}},
		Input{Name: "n", Value: &NumberValue{Value: 1.0}},
		Input{Name: "b", Value: &BooleanValue{Value: true}},
	}
	expected := EnvMap{"s": "s", "n": "1", "b": "true"}
	actual := inputs.ToEnvMap()
	assert.Equal(t, expected, actual)
}
