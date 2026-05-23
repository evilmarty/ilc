package inputs

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInputOptionsContains(t *testing.T) {
	options := InputOptions{
		{"A", "a"},
		{"B", "b"},
	}
	assert.True(t, options.Contains("a"))
	assert.False(t, options.Contains("c"))
}

func TestStringValue(t *testing.T) {
	t.Run("get and string", func(t *testing.T) {
		v := StringValue{Value: "foobar"}
		assert.Equal(t, "foobar", v.Get())
		assert.Equal(t, "foobar", v.String())
	})

	t.Run("no validation", func(t *testing.T) {
		v := StringValue{}
		assert.NoError(t, v.Set("hello"))
		assert.Equal(t, "hello", v.Value)
	})

	t.Run("with pattern", func(t *testing.T) {
		v := StringValue{Pattern: "bar$"}
		assert.NoError(t, v.Set("foobar"))
		assert.Equal(t, "foobar", v.Value)
		assert.Error(t, v.Set("foobaz"))
	})
}

func TestNumberValue(t *testing.T) {
	t.Run("get and string with decimals", func(t *testing.T) {
		v := NumberValue{Value: 1.234567}
		assert.Equal(t, 1.234567, v.Get())
		assert.Equal(t, "1.23457", v.String())
	})

	t.Run("get and string without decimals", func(t *testing.T) {
		v := NumberValue{Value: 1.0}
		assert.Equal(t, 1.0, v.Get())
		assert.Equal(t, "1", v.String())
	})

	t.Run("with min and max", func(t *testing.T) {
		v := NumberValue{MinValue: 1.0, MaxValue: 5.0}
		assert.NoError(t, v.Set("2.0"))
		assert.Equal(t, 2.0, v.Value)
		assert.Error(t, v.Set("0"))
		assert.Error(t, v.Set("5.1"))
	})
}

func TestBooleanValue(t *testing.T) {
	t.Run("get and string", func(t *testing.T) {
		v := BooleanValue{Value: true}
		assert.Equal(t, true, v.Get())
		assert.Equal(t, "true", v.String())
	})

	t.Run("set valid", func(t *testing.T) {
		v := BooleanValue{}
		assert.NoError(t, v.Set("true"))
		assert.True(t, v.Value)
		assert.NoError(t, v.Set("false"))
		assert.False(t, v.Value)
	})

	t.Run("set invalid", func(t *testing.T) {
		v := BooleanValue{}
		assert.Error(t, v.Set("nope"))
	})
}

func TestFlagSet_Parse(t *testing.T) {
	t.Run("parse from environment", func(t *testing.T) {
		fs := NewFlagSet("test", "ILC_INPUT_")
		sVal := &StringValue{}
		fs.Var(&Input{Name: "airport", Value: sVal})

		envs := map[string]string{
			"ILC_INPUT_AIRPORT": "lax",
		}
		err := fs.Parse(nil, envs, true)
		assert.NoError(t, err)
		assert.Equal(t, "lax", sVal.Value)
	})

	t.Run("parse from arguments", func(t *testing.T) {
		fs := NewFlagSet("test", "ILC_INPUT_")
		sVal := &StringValue{}
		fs.Var(&Input{Name: "airport", Value: sVal})

		err := fs.Parse([]string{"-airport", "lax"}, nil, true)
		assert.NoError(t, err)
		assert.Equal(t, "lax", sVal.Value)
	})

	t.Run("args override envs", func(t *testing.T) {
		fs := NewFlagSet("test", "ILC_INPUT_")
		sVal := &StringValue{}
		fs.Var(&Input{Name: "airport", Value: sVal})

		envs := map[string]string{
			"ILC_INPUT_AIRPORT": "bne",
		}
		err := fs.Parse([]string{"-airport", "lax"}, envs, true)
		assert.NoError(t, err)
		assert.Equal(t, "lax", sVal.Value)
	})

	t.Run("missing input errors in non-interactive", func(t *testing.T) {
		fs := NewFlagSet("test", "ILC_INPUT_")
		sVal := &StringValue{}
		fs.Var(&Input{Name: "airport", Value: sVal})

		err := fs.Parse(nil, nil, true)
		assert.Error(t, err)
	})
}

func TestFlagSet_Merge(t *testing.T) {
	fs1 := NewFlagSet("test", "ILC_INPUT_")
	v1 := &StringValue{}
	fs1.Var(&Input{Name: "airport", Value: v1})

	fs2 := NewFlagSet("test", "ILC_INPUT_")
	v2 := &NumberValue{}
	fs2.Var(&Input{Name: "count", Value: v2})

	merged := fs1.Merge(fs2)
	assert.True(t, merged.Has("airport"))
	assert.True(t, merged.Has("count"))
	assert.Equal(t, 2, len(merged.inputs))
}
