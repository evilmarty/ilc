package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRunnerGetInputValuesFromEnv(t *testing.T) {
	runner := Runner{
		Env: EnvMap{
			"ILC_INPUT_foo_bar": "foobar",
			"ILC_INPUT_foobar":  "nope",
			"ILC_INPUT_num_ber": "10",
			"TEST":              "true",
		},
	}
	expected := map[string]string{
		"foo-bar": "foobar",
		"num_ber": "10",
	}
	actual := runner.getInputValuesFromEnv(Inputs{
		{Name: "foo-bar", Value: &StringValue{}},
		{Name: "num_ber", Value: &NumberValue{}},
	})
	assert.Equal(t, expected, actual)
}

func TestRunnerIsReplay(t *testing.T) {
	runner := Runner{}
	runner.Args = []string{"arg1", "arg2"}
	assert.False(t, runner.isReplay())
	runner.Args = []string{"!arg1", "arg2"}
	assert.True(t, runner.isReplay())
	runner.Args = []string{"!", "arg1", "arg2"}
	assert.True(t, runner.isReplay())
}
