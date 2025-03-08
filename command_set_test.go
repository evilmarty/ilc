package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCommandSetString_Empty(t *testing.T) {
	cs := CommandSet{}
	assert.Empty(t, cs.String(), "CommandSet.String() to return an empty string")
}

func TestCommandSetString_One(t *testing.T) {
	cs := CommandSet{
		Commands: []ConfigCommand{
			{Name: "foobar"},
		},
	}
	assert.Equal(t, "foobar", cs.String(), "CommandSet.String() to return expected value")
}

func TestCommandSetString_Many(t *testing.T) {
	cs := CommandSet{
		Commands: []ConfigCommand{
			{Name: ""},
			{Name: "foobar"},
			{Name: "foobaz"},
		},
	}
	assert.Equal(t, "foobar foobaz", cs.String(), "CommandSet.String() to return expected value")
}

func TestCommandSetPure_Empty(t *testing.T) {
	cs := CommandSet{
		Commands: []ConfigCommand{},
	}
	assert.Equal(t, false, cs.Pure(), "CommandSet.Pure() to return false")
}

func TestCommandSetPure_One(t *testing.T) {
	cs := CommandSet{
		Commands: []ConfigCommand{
			{Pure: true},
		},
	}
	assert.True(t, cs.Pure(), "CommandSet.Pure() to return true")
}

func TestCommandSetPure_Many(t *testing.T) {
	cs := CommandSet{
		Commands: []ConfigCommand{
			{Pure: true},
			{Pure: false},
		},
	}
	assert.False(t, cs.Pure(), "CommandSet.Pure() to return false")
}

func TestCommandSetShell_Default(t *testing.T) {
	cs := CommandSet{
		Commands: []ConfigCommand{
			{Shell: []string{}},
			{Shell: []string{}},
		},
	}
	assert.Equal(t, DefaultShell, cs.Shell(), "CommandSet.Shell() to return default")
}

func TestCommandSetShell_Latest(t *testing.T) {
	cs := CommandSet{
		Commands: []ConfigCommand{
			{Shell: []string{"foobaz"}},
			{Shell: []string{"foobar"}},
		},
	}
	assert.Equal(t, []string{"foobar"}, cs.Shell(), "CommandSet.Shell() to return the latest entry")
}

func TestCommandSetShell_Parent(t *testing.T) {
	cs := CommandSet{
		Commands: []ConfigCommand{
			{Shell: []string{"foobaz"}},
			{Shell: []string{}},
		},
	}
	assert.Equal(t, []string{"foobaz"}, cs.Shell(), "CommandSet.Shell() to return the parent's entry")
}

func TestCommandSetInputs(t *testing.T) {
	cs := CommandSet{
		Commands: []ConfigCommand{
			{
				Inputs: ConfigInputs{
					ConfigInput{Name: "A"},
					ConfigInput{Name: "B"},
				},
			},
			{
				Inputs: ConfigInputs{
					ConfigInput{Name: "C"},
				},
			},
		},
	}
	expected := []ConfigInput{
		cs.Commands[0].Inputs[0],
		cs.Commands[0].Inputs[1],
		cs.Commands[1].Inputs[0],
	}
	assert.Equal(t, expected, cs.Inputs(), "CommandSet.Inputs() returned unexpected results")
}

func TestCommandSetEnv(t *testing.T) {
	cs := CommandSet{
		Commands: []ConfigCommand{
			{
				Env: map[string]string{
					"A": "a",
					"B": "b",
				},
			},
			{
				Env: map[string]string{
					"A": "aa",
					"C": "c",
				},
			},
		},
	}
	expected := EnvMap{
		"A": "aa",
		"B": "b",
		"C": "c",
	}
	assert.Equal(t, expected, cs.Env(), "CommandSet.Env() returned unexpected results")
}

func TestCommandSetRenderEnv_NonError(t *testing.T) {
	data := TemplateData{
		Input: map[string]any{
			"A": "a",
			"B": "b",
			"C": "c",
		},
	}
	cs := CommandSet{
		Commands: []ConfigCommand{
			{
				Env: map[string]string{
					"A": "{{.Input.A}}",
					"B": "{{.Input.B}}",
				},
			},
			{
				Env: map[string]string{
					"A": "aa",
					"C": "{{.Input.C}}",
				},
			},
		},
	}
	expected := EnvMap{
		"A": "aa",
		"B": "b",
		"C": "c",
	}
	actual, err := cs.RenderEnv(data)
	assert.NoError(t, err, "CommandSet.RenderEnv() returned an unexpected error")
	assert.Equal(t, expected, actual, "CommandSet.RenderEnv() returned unexpected results")
}

func TestCommandSetRenderEnv_TemplateError(t *testing.T) {
	data := TemplateData{
		Input: map[string]any{
			"A": "a",
		},
	}
	cs := CommandSet{
		Commands: []ConfigCommand{
			{
				Env: map[string]string{
					"A": "{{.Input.A}",
				},
			},
		},
	}
	_, actual := cs.RenderEnv(data)
	expected := "template error for environment variable: 'A' - template: :1: bad character U+007D '}'"
	assert.EqualError(t, actual, expected, "CommandSet.RenderEnv() returned unexpected error")
}

func TestCommandSetRenderScript(t *testing.T) {
	t.Run("empty command set", func(t *testing.T) {
		data := TemplateData{}
		cs := CommandSet{}
		_, err := cs.RenderScript(data)
		expected := "no script present"
		actual := fmt.Sprintf("%s", err)
		assert.Equal(t, expected, actual, "CommandSet.RenderScript() returned unexpected error")
	})
	t.Run("template error", func(t *testing.T) {
		data := TemplateData{
			Input: map[string]any{
				"A": "a",
			},
		}
		cs := CommandSet{
			Commands: []ConfigCommand{
				{
					Name: "foobar",
					Run:  "{{.Input.A}",
				},
			},
		}
		_, err := cs.RenderScript(data)
		expected := "template error: template: foobar:1: bad character U+007D '}'"
		actual := fmt.Sprintf("%s", err)
		assert.Equal(t, expected, actual, "CommandSet.RenderScript() returned unexpected error")
	})
	t.Run("script error", func(t *testing.T) {
		data := TemplateData{
			Input: map[string]any{
				"A": "a",
			},
		}
		cs := CommandSet{
			Commands: []ConfigCommand{
				{
					Name: "foobar",
					Run:  "{{template \"foobaz\"}}",
				},
			},
		}
		_, err := cs.RenderScript(data)
		expected := "script error: template: foobar:1:11: executing \"foobar\" at <{{template \"foobaz\"}}>: template \"foobaz\" not defined"
		actual := fmt.Sprintf("%s", err)
		assert.Equal(t, expected, actual, "CommandSet.RenderScript() returned unexpected error")
	})
	t.Run("render single template", func(t *testing.T) {
		data := TemplateData{
			Input: map[string]any{
				"A": "a",
				"B": "b",
			},
		}
		cs := CommandSet{
			Commands: []ConfigCommand{
				{
					Name: "foobaz",
					Run:  "echo {{.Input.B}}",
				},
				{
					Name: "foobar",
					Run:  "echo {{.Input.A}}",
				},
			},
		}
		expected := "echo a"
		actual, err := cs.RenderScript(data)
		if err != nil {
			t.Fatalf("CommandSet.RenderScript() returned an unexpected error: %v", err)
		}
		assert.Equal(t, expected, actual, "CommandSet.RenderScript() returned unexpected result")
	})
	t.Run("render multiple templates", func(t *testing.T) {
		data := TemplateData{
			Input: map[string]any{
				"A": "a",
				"B": "b",
			},
		}
		cs := CommandSet{
			Commands: []ConfigCommand{
				{
					Name: "foobaz",
					Run:  "echo {{.Input.B}}",
				},
				{
					Name: "foobar",
					Run:  "echo {{.Input.A}} {{template \"foobaz\" .}}",
				},
			},
		}
		expected := "echo a echo b"
		actual, err := cs.RenderScript(data)
		if err != nil {
			t.Fatalf("CommandSet.RenderScript() returned an unexpected error: %v", err)
		}
		assert.Equal(t, expected, actual, "CommandSet.RenderScript() returned unexpected result")
	})
	t.Run("latest command overrides existing template", func(t *testing.T) {
		data := TemplateData{
			Input: map[string]any{
				"A": "a",
				"B": "b",
			},
		}
		cs := CommandSet{
			Commands: []ConfigCommand{
				{
					Name: "foobar",
					Run:  "echo {{.Input.A}}",
				},
				{
					Name: "foobaz",
					Run:  "echo {{.Input.B}}",
				},
				{
					Name: "foobar",
					Run:  "echo {{.Input.A}} {{template \"foobaz\" .}}",
				},
			},
		}
		expected := "echo a echo b"
		actual, err := cs.RenderScript(data)
		assert.NoError(t, err, "CommandSet.RenderScript() returned an unexpected error")
		assert.Equal(t, expected, actual, "CommandSet.RenderScript() returned unexpected result")
	})
	t.Run("helper functions", func(t *testing.T) {
		data := TemplateData{
			Input: map[string]any{
				"A": "a",
				"B": "b",
			},
			Env: map[string]string{
				"C": "c",
				"D": "d",
			},
		}
		cs := CommandSet{
			Commands: []ConfigCommand{
				{
					Name: "foobar",
					Run:  "echo {{input \"A\"}} {{env \"C\"}}",
				},
			},
		}
		expected := "echo a c"
		actual, err := cs.RenderScript(data)
		assert.NoError(t, err, "CommandSet.RenderScript() returned an unexpected error")
		assert.Equal(t, expected, actual, "CommandSet.RenderScript() returned unexpected result")
	})
}

func TestCommandSetRenderScriptToTemp(t *testing.T) {
	data := TemplateData{
		Input: map[string]any{
			"A": "a",
		},
	}
	cs := CommandSet{
		Commands: []ConfigCommand{
			{
				Name: "foobar",
				Run:  "echo {{.Input.A}}",
			},
		},
	}
	expected := "echo a"
	file, err := cs.RenderScriptToTemp(data)
	if err != nil {
		t.Fatalf("CommandSet.RenderScriptToTemp() returned an unexpected error: %v", err)
	}
	actual, err := readTextFile(file)
	assert.NoError(t, err, "Could not read file containing rendered script")
	assert.Equal(t, expected, actual, "CommandSet.RenderScriptToTemp() returned unexpected result")
}

func TestCommandSetCmd_IsPure(t *testing.T) {
	data := TemplateData{
		Input: map[string]any{
			"A": "a",
			"B": "b",
			"C": "c",
		},
	}
	moreEnviron := []string{
		"D=d",
	}
	cs := CommandSet{
		Commands: []ConfigCommand{
			{
				Shell: []string{"/bin/sh"},
				Run:   "foobaz",
				Env: map[string]string{
					"A": "{{.Input.A}}",
					"B": "{{.Input.B}}",
				},
				Inputs: ConfigInputs{
					ConfigInput{Name: "A"},
					ConfigInput{Name: "B"},
				},
			},
			{
				Shell: []string{"/bin/bash", "-x"},
				Run:   "foobar",
				Env: map[string]string{
					"A": "aa",
					"C": "{{.Input.C}}",
				},
				Inputs: ConfigInputs{
					ConfigInput{Name: "C"},
				},
				Pure: true,
			},
		},
	}
	cmd, err := cs.Cmd(data, moreEnviron)
	assert.NoError(t, err, "CommandSet.Cmd() returned an unexpected error")
	assert.ElementsMatch(
		t,
		[]string{"ILC_INPUT_A=a", "ILC_INPUT_B=b", "ILC_INPUT_C=c", "A=aa", "B=b", "C=c"},
		cmd.Env,
		"CommandSet.Cmd() did not set cmd.Env with correct values",
	)
	assert.Equal(t, "/bin/bash", cmd.Path, "CommandSet.Cmd() did not set cmd.Path to the shell path")
	assert.Equal(t, []string{"/bin/bash", "-x"}, cmd.Args[:len(cmd.Args)-1], "CommandSet.Cmd() did not set cmd.Args with the correct values")
}

func TestCommandSetCmd_NotPure(t *testing.T) {
	data := TemplateData{
		Input: map[string]any{
			"A": "a",
			"B": "b",
			"C": "c",
		},
	}
	moreEnviron := []string{
		"C=x",
		"D=d",
	}
	cs := CommandSet{
		Commands: []ConfigCommand{
			{
				Shell: []string{"/bin/sh"},
				Run:   "foobaz",
				Env: map[string]string{
					"A": "{{.Input.A}}",
					"B": "{{.Input.B}}",
				},
				Inputs: ConfigInputs{
					ConfigInput{Name: "A"},
					ConfigInput{Name: "B"},
				},
			},
			{
				Shell: []string{"/bin/bash", "-x"},
				Run:   "foobar",
				Env: map[string]string{
					"A": "aa",
					"C": "{{.Input.C}}",
				},
				Inputs: ConfigInputs{
					ConfigInput{Name: "C"},
				},
			},
		},
	}
	cmd, err := cs.Cmd(data, moreEnviron)
	assert.NoError(t, err, "CommandSet.Cmd() returned an unexpected error")
	assert.ElementsMatch(
		t,
		[]string{"ILC_INPUT_A=a", "ILC_INPUT_B=b", "ILC_INPUT_C=c", "A=aa", "B=b", "C=c", "D=d"},
		cmd.Env,
		"CommandSet.Cmd() did not set cmd.Env with correct values",
	)
	assert.Equal(t, "/bin/bash", cmd.Path, "CommandSet.Cmd() did not set cmd.Path to the shell path")
	assert.Equal(t, []string{"/bin/bash", "-x"}, cmd.Args[:len(cmd.Args)-1], "CommandSet.Cmd() did not set cmd.Args with the correct values")
}

func TestCommandSetParseArgs(t *testing.T) {
	t.Run("string", func(t *testing.T) {
		cs := CommandSet{
			Commands: []ConfigCommand{
				{
					Inputs: ConfigInputs{
						ConfigInput{Name: "A", DefaultValue: ""},
						ConfigInput{Name: "B", DefaultValue: "b"},
					},
				},
				{
					Inputs: ConfigInputs{
						ConfigInput{Name: "C", DefaultValue: "c"},
					},
				},
			},
			Args: []string{"-A", "aa", "--C", "cc"},
		}
		expected := map[string]any{
			"A": "aa",
			"C": "cc",
		}
		actual := make(map[string]any)
		err := cs.ParseArgs(&actual)
		assert.NoError(t, err, "CommandSet.ParseArgs() returned unexpected error")
		assert.Equal(t, expected, actual, "CommandSet.ParseArgs() returned unexpected results")
	})
	t.Run("boolean", func(t *testing.T) {
		cs := CommandSet{
			Commands: []ConfigCommand{
				{
					Inputs: ConfigInputs{
						ConfigInput{Name: "A", Type: "boolean", DefaultValue: false},
						ConfigInput{Name: "B", Type: "boolean", DefaultValue: false},
						ConfigInput{Name: "C", Type: "boolean", DefaultValue: true},
					},
				},
			},
			Args: []string{"--A", "-C=false"},
		}
		expected := map[string]any{
			"A": true,
			"C": false,
		}
		actual := make(map[string]any)
		err := cs.ParseArgs(&actual)
		assert.NoError(t, err, "CommandSet.ParseArgs() returned unexpected error")
		assert.Equal(t, expected, actual, "CommandSet.ParseArgs() returned unexpected results")
	})
	t.Run("help", func(t *testing.T) {
		cs := CommandSet{
			Commands: []ConfigCommand{
				{
					Inputs: ConfigInputs{
						ConfigInput{Name: "A", DefaultValue: ""},
						ConfigInput{Name: "B", DefaultValue: "b"},
					},
				},
				{
					Inputs: ConfigInputs{
						ConfigInput{Name: "C", DefaultValue: "c"},
					},
				},
			},
			Args: []string{"-help"},
		}
		values := make(map[string]any)
		actual := cs.ParseArgs(&values)
		assert.Equal(t, flag.ErrHelp, actual, "CommandSet.ParseArgs() did not acknowledge help")
	})
}

func TestCommandSetParseEnv(t *testing.T) {
	cs := CommandSet{
		Commands: []ConfigCommand{
			{
				Inputs: ConfigInputs{
					ConfigInput{Name: "A", DefaultValue: ""},
					ConfigInput{Name: "B", DefaultValue: "b"},
				},
			},
			{
				Inputs: ConfigInputs{
					ConfigInput{Name: "C", DefaultValue: "c"},
				},
			},
		},
		Args: []string{"-help"},
	}
	env := []string{"ILC_INPUT_A=a=a", "ILC_INPUT_C=", "ILC_INPUT_D=dd"}
	expected := map[string]any{"A": "a=a", "C": ""}
	actual := make(map[string]any)
	cs.ParseEnv(&actual, env)
	assert.Equal(t, expected, actual, "CommandSet.ParseEnv() returned unexpected results")
}

func TestCommandSetValidate_Empty(t *testing.T) {
	cs := CommandSet{
		Commands: []ConfigCommand{
			{},
			{},
		},
		Args: []string{},
	}
	actual := cs.Validate(map[string]any{})
	assert.Nil(t, actual, "CommandSet.Validate() returned unexpected error")
}

func TestCommandSetValidate_Missing(t *testing.T) {
	cs := CommandSet{
		Commands: []ConfigCommand{
			{
				Inputs: []ConfigInput{
					{Name: "A", DefaultValue: ""},
				},
			},
		},
		Args: []string{},
	}
	actual := cs.Validate(map[string]any{})
	expected := "missing input: A"
	assert.EqualError(t, actual, expected, "CommandSet.Validate() returned unexpected error")
}

func TestCommandSetValidate_Invalid(t *testing.T) {
	cs := CommandSet{
		Commands: []ConfigCommand{
			{
				Inputs: []ConfigInput{
					{Name: "A", Pattern: "[a-z]+"},
				},
			},
		},
		Args: []string{},
	}
	actual := cs.Validate(map[string]any{"A": "123"})
	expected := "invalid input: A"
	assert.EqualError(t, actual, expected, "CommandSet.Validate() returned unexpected error")
}

func TestCommandSetValidate_Valid(t *testing.T) {
	cs := CommandSet{
		Commands: []ConfigCommand{
			{
				Inputs: []ConfigInput{
					{Name: "A"},
				},
			},
		},
		Args: []string{},
	}
	actual := cs.Validate(map[string]any{"A": "123"})
	assert.Nil(t, actual, "CommandSet.Validate() returned unexpected error")
}

func TestCommandSetRunnable_True(t *testing.T) {
	cs := CommandSet{
		Commands: []ConfigCommand{
			{
				Run: "",
			},
			{
				Run: "true",
			},
		},
	}
	assert.True(t, cs.Runnable(), "CommandSet.Runnable() expected to return true")
}

func TestCommandSetRunnable_False(t *testing.T) {
	cs := CommandSet{
		Commands: []ConfigCommand{
			{
				Run: "",
			},
		},
	}
	assert.False(t, cs.Runnable(), "CommandSet.Runnable() expected to return false")
}

func TestCommandSetSelected_True(t *testing.T) {
	cs := CommandSet{
		Commands: []ConfigCommand{
			{},
		},
	}
	assert.True(t, cs.Selected(), "CommandSet.Selected() expected to return true")
}

func TestCommandSetSelected_False(t *testing.T) {
	cs := CommandSet{
		Commands: []ConfigCommand{
			{
				Commands: []ConfigCommand{
					{},
				},
			},
		},
	}
	assert.False(t, cs.Selected(), "CommandSet.Selected() expected to return false")
}

func TestNewCommandSet_PreselectedFromArgs(t *testing.T) {
	config := Config{
		Commands: ConfigCommands{
			ConfigCommand{Name: "foobar", Run: "true"},
		},
	}
	cs, err := NewCommandSet(config, []string{"foobar", "-a", "1"})
	assert.NoError(t, err, "NewCommandSet() returned unexpected error")
	assert.Equal(t, cs.Config, config, "NewCommandSet() did not set config")
	assert.Equal(t, cs.Args, []string{"-a", "1"}, "NewCommandSet() has unexpected args")
}

func TestNewCommandSet_Invalid(t *testing.T) {
	config := Config{
		Commands: ConfigCommands{
			ConfigCommand{Name: "foobar", Run: "true"},
		},
	}
	_, actual := NewCommandSet(config, []string{"foobaz", "-a", "1"})
	expected := "invalid subcommand: foobaz"
	assert.EqualError(t, actual, expected, "NewCommandSet() returned unexpected error")
}

func readTextFile(name string) (string, error) {
	var str strings.Builder
	data, err := os.ReadFile(name)
	if err != nil {
		return "", err
	}
	if _, err := str.Write(data); err != nil {
		return "", err
	}
	return str.String(), nil
}
