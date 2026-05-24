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

type MockPrompter struct {
	PromptFunc    func(title string, missing []*Input) error
	Called        bool
	PassedTitle   string
	PassedMissing []*Input
}

func (m *MockPrompter) Prompt(title string, missing []*Input) error {
	m.Called = true
	m.PassedTitle = title
	m.PassedMissing = missing
	if m.PromptFunc != nil {
		return m.PromptFunc(title, missing)
	}
	return nil
}

func TestFlagSet_ParseInteractiveMocked(t *testing.T) {
	fs := NewFlagSet("my-flagset", "ILC_INPUT_")
	sVal := &StringValue{}
	fs.Var(&Input{Name: "airport", Value: sVal})

	mock := &MockPrompter{
		PromptFunc: func(title string, missing []*Input) error {
			assert.Equal(t, "my-flagset", title)
			assert.Len(t, missing, 1)
			assert.Equal(t, "airport", missing[0].Name)
			// Simulate user typing input in the TUI prompt:
			return missing[0].Value.Set("bne")
		},
	}
	fs.Prompter = mock

	// Run interactive Parse with nonInteractive = false
	err := fs.Parse(nil, nil, false)
	assert.NoError(t, err)
	assert.True(t, mock.Called)
	assert.Equal(t, "bne", sVal.Value)
}

func TestInputOptionString(t *testing.T) {
	opt := InputOption{Label: "Brisbane", Value: "bne"}
	assert.Equal(t, "Brisbane", opt.String())
}

func TestFlagSet_InputsValuesToEnvMapToArgs(t *testing.T) {
	fs := NewFlagSet("test-fs", "ILC_INPUT_")
	strIn := &Input{Name: "str", Value: &StringValue{Value: "hello"}}
	numIn := &Input{Name: "num", Value: &NumberValue{Value: 42}}
	boolInTrue := &Input{Name: "bool-true", Value: &BooleanValue{Value: true}}
	boolInFalse := &Input{Name: "bool-false", Value: &BooleanValue{Value: false}}
	fs.Var(strIn)
	fs.Var(numIn)
	fs.Var(boolInTrue)
	fs.Var(boolInFalse)

	// Test Inputs()
	assert.Len(t, fs.Inputs(), 4)

	// Test Values()
	vals := fs.Values()
	assert.Equal(t, "hello", vals["str"])
	assert.Equal(t, 42.0, vals["num"])
	assert.Equal(t, true, vals["bool-true"])
	assert.Equal(t, false, vals["bool-false"])

	// Test ToEnvMap()
	em := fs.ToEnvMap()
	assert.Equal(t, "hello", em["ILC_INPUT_STR"])
	assert.Equal(t, "42", em["ILC_INPUT_NUM"])
	assert.Equal(t, "true", em["ILC_INPUT_BOOL_TRUE"])
	assert.Equal(t, "false", em["ILC_INPUT_BOOL_FALSE"])

	// Test ToArgs()
	args := fs.ToArgs()
	assert.Contains(t, args, "-str")
	assert.Contains(t, args, "hello")
	assert.Contains(t, args, "-bool-true")
	assert.Contains(t, args, "-bool-false=false")
}

func TestValues_ValidateLive(t *testing.T) {
	// StringValue ValidateLive
	sVal := StringValue{Pattern: "^[a-z]+$"}
	assert.NoError(t, sVal.ValidateLive("valid"))
	assert.Error(t, sVal.ValidateLive("INVALID"))

	// NumberValue ValidateLive
	nVal := NumberValue{MinValue: 1.0, MaxValue: 10.0}
	assert.NoError(t, nVal.ValidateLive("5"))
	assert.NoError(t, nVal.ValidateLive("-")) // incomplete number is ignored/valid live
	assert.NoError(t, nVal.ValidateLive("+"))
	assert.NoError(t, nVal.ValidateLive("."))
	assert.NoError(t, nVal.ValidateLive("1e-"))
	assert.NoError(t, nVal.ValidateLive("1e+"))
	assert.NoError(t, nVal.ValidateLive("1E"))
	assert.Error(t, nVal.ValidateLive("12"))

	// BooleanValue ValidateLive
	bVal := BooleanValue{}
	assert.NoError(t, bVal.ValidateLive("true"))
	assert.Error(t, bVal.ValidateLive("nope"))
}

