package main

import (
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSelectionSelect(t *testing.T) {
	selection := NewSelection(
		Command{
			Name: "foo",
			Commands: SubCommands{
				{Command: Command{Name: "bar"}},
			},
		},
	)
	selection.Args = []string{"foo"}
	t.Run("when args are empty", func(t *testing.T) {
		expected := Selection{
			commands: selection.commands,
			Args:     []string{},
		}
		actual := selection.Select([]string{})
		assert.Equal(t, expected, actual)
		assert.NotEqual(t, selection, actual)
	})
	t.Run("when command is not found", func(t *testing.T) {
		expected := Selection{
			commands: selection.commands,
			Args:     []string{"baz"},
		}
		actual := selection.Select([]string{"baz"})
		assert.Equal(t, expected, actual)
		assert.NotEqual(t, selection, actual)
	})
	t.Run("when command is found", func(t *testing.T) {
		expected := Selection{
			commands: append(selection.commands, selection.commands[0].Commands[0].Command),
			Args:     []string{"baz"},
		}
		actual := selection.Select([]string{"bar", "baz"})
		assert.Equal(t, expected, actual)
		assert.NotEqual(t, selection, actual)
	})
}

func TestSelectionSelectCommand(t *testing.T) {
	selection := NewSelection(
		Command{Name: "foo"},
	)
	selection.Args = []string{"foo"}
	expected := Selection{
		commands: []Command{
			{Name: "foo"},
			{Name: "bar"},
		},
		Args: []string{"bar"},
	}
	actual := selection.SelectCommand(Command{Name: "bar"}, []string{"bar"})
	assert.Equal(t, expected, actual)
	assert.NotEqual(t, selection, actual)
}

func TestSelectionString(t *testing.T) {
	selection := NewSelection(
		Command{Name: ""},
		Command{Name: "foo"},
		Command{Name: ""},
		Command{Name: "bar"},
		Command{Name: ""},
		Command{Name: "baz"},
	)
	expected := "foo bar baz"
	actual := selection.String()
	assert.Equal(t, expected, actual)
}

func TestSelectionRunnable(t *testing.T) {
	t.Run("when latest has run and no subcommands", func(t *testing.T) {
		selection := NewSelection(Command{Run: "echo foobar"})
		assert.True(t, selection.Runnable())
	})
	t.Run("when latest has run and subcommands", func(t *testing.T) {
		selection := NewSelection(Command{
			Run: "echo foobar", Commands: SubCommands{SubCommand{}},
		})
		assert.False(t, selection.Runnable())
	})
	t.Run("when latest has subcommands", func(t *testing.T) {
		selection := NewSelection(Command{
			Commands: SubCommands{SubCommand{}},
		})
		assert.False(t, selection.Runnable())
	})
}

func TestSelectionDescription(t *testing.T) {
	selection := NewSelection(
		Command{Description: "foobar"},
		Command{Description: "foobaz"},
	)
	expected := "foobaz"
	actual := selection.Description()
	assert.Equal(t, expected, actual)
}

func TestSelectionRun(t *testing.T) {
	t.Run("gets latest", func(t *testing.T) {
		selection := NewSelection(
			Command{Run: "foobar"},
			Command{Run: "foobaz"},
		)
		expected := "foobaz"
		actual := selection.Run()
		assert.Equal(t, expected, actual)
	})
	t.Run("gets latest even when blank", func(t *testing.T) {
		selection := NewSelection(
			Command{Run: "foobar"},
			Command{Run: ""},
		)
		expected := ""
		actual := selection.Run()
		assert.Equal(t, expected, actual)
	})
}

func TestSelectionShell(t *testing.T) {
	t.Run("uses latest", func(t *testing.T) {
		selection := NewSelection(
			Command{Shell: []string{"/bin/bash"}},
			Command{Shell: []string{"/bin/zsh"}},
		)
		expected := []string{"/bin/zsh"}
		actual := selection.Shell()
		assert.Equal(t, expected, actual)
	})
	t.Run("fallsback on preceding command", func(t *testing.T) {
		selection := NewSelection(
			Command{Shell: []string{"/bin/zsh"}},
			Command{Shell: []string{"/bin/bash"}},
			Command{},
		)
		expected := []string{"/bin/bash"}
		actual := selection.Shell()
		assert.Equal(t, expected, actual)
	})
	t.Run("default when none defined", func(t *testing.T) {
		selection := NewSelection(
			Command{},
			Command{},
		)
		expected := DefaultShell
		actual := selection.Shell()
		assert.Equal(t, expected, actual)
	})
	t.Run("default when empty", func(t *testing.T) {
		selection := NewSelection(Command{})
		expected := DefaultShell
		actual := selection.Shell()
		assert.Equal(t, expected, actual)
	})
}

func TestSelectionEnv(t *testing.T) {
	t.Run("none", func(t *testing.T) {
		selection := NewSelection(
			Command{},
			Command{},
		)
		expected := EnvMap{}
		actual := selection.Env()
		assert.Equal(t, expected, actual)
	})
	t.Run("merged env", func(t *testing.T) {
		selection := NewSelection(
			Command{
				Env: EnvMap{
					"A": "a",
					"B": "b",
				},
			},
			Command{
				Env: EnvMap{
					"A": "aa",
					"C": "c",
				},
			},
		)
		expected := EnvMap{
			"A": "aa",
			"B": "b",
			"C": "c",
		}
		actual := selection.Env()
		assert.Equal(t, expected, actual)
	})
}

func TestSelectionPure(t *testing.T) {
	t.Run("latest is set", func(t *testing.T) {
		selection := NewSelection(
			Command{},
			Command{Pure: true},
		)
		assert.True(t, selection.Pure())
	})
	t.Run("does not inherit", func(t *testing.T) {
		selection := NewSelection(
			Command{Pure: true},
			Command{},
		)
		assert.False(t, selection.Pure())
	})
}

func TestSelectionInputs(t *testing.T) {
	t.Run("no inputs", func(t *testing.T) {
		selection := NewSelection(
			Command{},
			Command{},
		)
		expected := Inputs{}
		actual := selection.Inputs()
		assert.Equal(t, expected, actual)
	})
	t.Run("merged inputs", func(t *testing.T) {
		selection := NewSelection(
			Command{
				Inputs: Inputs{
					Input{Name: "a", Value: &StringValue{}},
					Input{Name: "b", Value: &NumberValue{}},
				},
			},
			Command{
				Inputs: Inputs{
					Input{Name: "a", Value: &NumberValue{}},
					Input{Name: "c", Value: &NumberValue{}},
				},
			},
		)
		expected := Inputs{
			Input{Name: "a", Value: &NumberValue{}},
			Input{Name: "b", Value: &NumberValue{}},
			Input{Name: "c", Value: &NumberValue{}},
		}
		actual := selection.Inputs()
		assert.Equal(t, expected, actual)
	})
}

func TestSelectionRenderScript(t *testing.T) {
	t.Run("template error", func(t *testing.T) {
		data := TemplateData{
			Input: map[string]any{
				"A": "a",
			},
		}
		selection := NewSelection(
			Command{
				Name: "foobar",
				Run:  "{{.Input.A}",
			},
		)
		_, err := selection.RenderScript(data)
		assert.ErrorContains(t, err, "template error: template: foobar:1: bad character U+007D '}'")
	})
	t.Run("script error", func(t *testing.T) {
		data := TemplateData{
			Input: map[string]any{
				"A": "a",
			},
		}
		selection := NewSelection(
			Command{
				Name: "foobar",
				Run:  "{{template \"foobaz\"}}",
			},
		)
		_, err := selection.RenderScript(data)
		assert.ErrorContains(t, err, "script error: template: foobar:1:11: executing ")
	})
	t.Run("render single template", func(t *testing.T) {
		data := TemplateData{
			Input: map[string]any{
				"A": "a",
				"B": "b",
			},
		}
		selection := NewSelection(
			Command{
				Name: "foobaz",
				Run:  "echo {{.Input.B}}",
			},
			Command{
				Name: "foobar",
				Run:  "echo {{.Input.A}}",
			},
		)
		expected := "echo a"
		actual, err := selection.RenderScript(data)
		assert.NoError(t, err)
		assert.Equal(t, expected, actual)
	})
	t.Run("render multiple templates", func(t *testing.T) {
		data := TemplateData{
			Input: map[string]any{
				"A": "a",
				"B": "b",
			},
		}
		selection := NewSelection(
			Command{
				Name: "foobaz",
				Run:  "echo {{.Input.B}}",
			},
			Command{
				Name: "foobar",
				Run:  "echo {{.Input.A}} {{template \"foobaz\" .}}",
			},
		)
		expected := "echo a echo b"
		actual, err := selection.RenderScript(data)
		assert.NoError(t, err)
		assert.Equal(t, expected, actual, "CommandSet.RenderScript() returned unexpected result")
	})
	t.Run("latest command overrides existing template", func(t *testing.T) {
		data := TemplateData{
			Input: map[string]any{
				"A": "a",
				"B": "b",
			},
		}
		selection := NewSelection(
			Command{
				Name: "foobar",
				Run:  "echo {{.Input.A}}",
			},
			Command{
				Name: "foobaz",
				Run:  "echo {{.Input.B}}",
			},
			Command{
				Name: "foobar",
				Run:  "echo {{.Input.A}} {{template \"foobaz\" .}}",
			},
		)
		expected := "echo a echo b"
		actual, err := selection.RenderScript(data)
		assert.NoError(t, err)
		assert.Equal(t, expected, actual)
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
		selection := NewSelection(
			Command{
				Name: "foobar",
				Run:  "echo {{input \"A\"}} {{env \"C\"}}",
			},
		)
		expected := "echo a c"
		actual, err := selection.RenderScript(data)
		assert.NoError(t, err)
		assert.Equal(t, expected, actual)
	})
}

func TestSelectionRenderScriptToTemp(t *testing.T) {
	data := TemplateData{
		Input: map[string]any{
			"A": "a",
		},
	}
	selection := NewSelection(
		Command{
			Name: "foobar",
			Run:  "echo {{.Input.A}}",
		},
	)
	expected := "echo a"
	file, err := selection.RenderScriptToTemp(data)
	assert.NoError(t, err)
	actual, err := readTextFile(file)
	assert.NoError(t, err)
	assert.Equal(t, expected, actual)
}

func TestSelectionCmd(t *testing.T) {
	t.Run("is pure", func(t *testing.T) {
		data := TemplateData{
			Input: map[string]any{
				"A": "a",
				"B": "b",
				"C": "c",
			},
		}
		moreEnviron := map[string]string{"D": "d"}
		selection := NewSelection(
			Command{
				Shell: []string{"/bin/sh"},
				Run:   "foobaz",
				Env: map[string]string{
					"A": "{{.Input.A}}",
					"B": "{{.Input.B}}",
				},
				Inputs: Inputs{
					{Name: "A"},
					{Name: "B"},
				},
			},
			Command{
				Shell: []string{"/bin/bash", "-x"},
				Run:   "foobar",
				Env: map[string]string{
					"A": "aa",
					"C": "{{.Input.C}}",
				},
				Inputs: Inputs{
					{Name: "C"},
				},
				Pure: true,
			},
		)
		cmd, err := selection.Cmd(data, moreEnviron)
		assert.NoError(t, err)
		assert.ElementsMatch(
			t,
			[]string{"A=aa", "B=b", "C=c"},
			cmd.Env,
			"CommandSet.Cmd() did not set cmd.Env with correct values",
		)
		assert.Equal(t, "/bin/bash", cmd.Path)
		assert.Equal(t, []string{"/bin/bash", "-x"}, cmd.Args[:len(cmd.Args)-1])
	})
	t.Run("is not pure", func(t *testing.T) {
		data := TemplateData{
			Input: map[string]any{
				"A": "a",
				"B": "b",
				"C": "c",
			},
		}
		moreEnviron := map[string]string{
			"C": "x",
			"D": "d",
		}
		selection := NewSelection(
			Command{
				Shell: []string{"/bin/sh"},
				Run:   "foobaz",
				Env: map[string]string{
					"A": "{{.Input.A}}",
					"B": "{{.Input.B}}",
				},
				Inputs: Inputs{
					{Name: "A"},
					{Name: "B"},
				},
			},
			Command{
				Shell: []string{"/bin/bash", "-x"},
				Run:   "foobar",
				Env: map[string]string{
					"A": "aa",
					"C": "{{.Input.C}}",
				},
				Inputs: Inputs{
					{Name: "C"},
				},
			},
		)
		cmd, err := selection.Cmd(data, moreEnviron)
		assert.NoError(t, err)
		assert.ElementsMatch(t, []string{"A=aa", "B=b", "C=c", "D=d"}, cmd.Env)
		assert.Equal(t, "/bin/bash", cmd.Path)
		assert.Equal(t, []string{"/bin/bash", "-x"}, cmd.Args[:len(cmd.Args)-1])
	})
}

func TestSelectionToArgs(t *testing.T) {
	selection := NewSelection(
		Command{
			Name: "",
			Inputs: Inputs{
				{Name: "arg1", Value: &StringValue{Value: "foobar"}},
				{Name: "arg2", Value: &NumberValue{Value: 123}},
			},
		},
		Command{
			Name: "command1",
			Inputs: Inputs{
				{Name: "arg1", Value: &StringValue{Value: "foobar"}},
				{Name: "arg2", Value: &NumberValue{Value: 123}},
			},
		},
		Command{
			Name: "command2",
			Inputs: Inputs{
				{Name: "arg3", Value: &BooleanValue{Value: true}},
			},
		},
	)
	expected := []string{"command1", "command2", "-arg1", "foobar", "-arg2", "123", "-arg3"}
	actual := selection.ToArgs()
	assert.Equal(t, expected, actual)
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
