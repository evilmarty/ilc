package main

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSelectionString(t *testing.T) {
	t.Run("empty", func(t *testing.T) {
		commands := Selection{}
		expected := ""
		actual := commands.String()
		assert.Equal(t, expected, actual)
	})
	commands := Selection{
		Command{Name: ""},
		Command{Name: "foo"},
		Command{Name: ""},
		Command{Name: "bar"},
		Command{Name: ""},
		Command{Name: "baz"},
	}
	expected := "foo bar baz"
	actual := commands.String()
	assert.Equal(t, expected, actual)
}

func TestSelectionRunnable(t *testing.T) {
	t.Run("when empty", func(t *testing.T) {
		commands := Selection{}
		assert.False(t, commands.Runnable())
	})
	t.Run("when latest has run and no subcommands", func(t *testing.T) {
		commands := Selection{{Run: "echo foobar"}}
		assert.True(t, commands.Runnable())
	})
	t.Run("when latest has run and subcommands", func(t *testing.T) {
		commands := Selection{{Run: "echo foobar", Commands: SubCommands{SubCommand{}}}}
		assert.False(t, commands.Runnable())
	})
	t.Run("when latest has subcommands", func(t *testing.T) {
		commands := Selection{{Commands: SubCommands{SubCommand{}}}}
		assert.False(t, commands.Runnable())
	})
}

func TestSelectionDescription(t *testing.T) {
	commands := Selection{
		Command{Description: "foobar"},
		Command{Description: "foobaz"},
	}
	expected := "foobaz"
	actual := commands.Description()
	assert.Equal(t, expected, actual)
}

func TestSelectionRun(t *testing.T) {
	t.Run("empty", func(t *testing.T) {
		commands := Selection{}
		expected := ""
		actual := commands.Run()
		assert.Equal(t, expected, actual)
	})
	t.Run("gets latest", func(t *testing.T) {
		commands := Selection{
			Command{Run: "foobar"},
			Command{Run: "foobaz"},
		}
		expected := "foobaz"
		actual := commands.Run()
		assert.Equal(t, expected, actual)
	})
	t.Run("gets latest even when blank", func(t *testing.T) {
		commands := Selection{
			Command{Run: "foobar"},
			Command{Run: ""},
		}
		expected := ""
		actual := commands.Run()
		assert.Equal(t, expected, actual)
	})
}

func TestSelectionShell(t *testing.T) {
	t.Run("uses latest", func(t *testing.T) {
		commands := Selection{
			Command{Shell: []string{"/bin/bash"}},
			Command{Shell: []string{"/bin/zsh"}},
		}
		expected := []string{"/bin/zsh"}
		actual := commands.Shell()
		assert.Equal(t, expected, actual)
	})
	t.Run("fallsback on preceding command", func(t *testing.T) {
		commands := Selection{
			Command{Shell: []string{"/bin/zsh"}},
			Command{Shell: []string{"/bin/bash"}},
			Command{},
		}
		expected := []string{"/bin/bash"}
		actual := commands.Shell()
		assert.Equal(t, expected, actual)
	})
	t.Run("default when none defined", func(t *testing.T) {
		commands := Selection{
			Command{},
			Command{},
		}
		expected := DefaultShell
		actual := commands.Shell()
		assert.Equal(t, expected, actual)
	})
	t.Run("default when empty", func(t *testing.T) {
		commands := Selection{}
		expected := DefaultShell
		actual := commands.Shell()
		assert.Equal(t, expected, actual)
	})
}

func TestSelectionEnv(t *testing.T) {
	t.Run("empty", func(t *testing.T) {
		commands := Selection{}
		expected := EnvMap{}
		actual := commands.Env()
		assert.Equal(t, expected, actual)
	})
	t.Run("none", func(t *testing.T) {
		commands := Selection{
			Command{},
			Command{},
		}
		expected := EnvMap{}
		actual := commands.Env()
		assert.Equal(t, expected, actual)
	})
	t.Run("merged env", func(t *testing.T) {
		commands := Selection{
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
		}
		expected := EnvMap{
			"A": "aa",
			"B": "b",
			"C": "c",
		}
		actual := commands.Env()
		assert.Equal(t, expected, actual)
	})
}

func TestSelectionPure(t *testing.T) {
	t.Run("empty", func(t *testing.T) {
		commands := Selection{}
		assert.False(t, commands.Pure())
	})
	t.Run("latest is set", func(t *testing.T) {
		commands := Selection{
			Command{},
			Command{Pure: true},
		}
		assert.True(t, commands.Pure())
	})
	t.Run("does not inherit", func(t *testing.T) {
		commands := Selection{
			Command{Pure: true},
			Command{},
		}
		assert.False(t, commands.Pure())
	})
}

func TestSelectionInputs(t *testing.T) {
	t.Run("empty", func(t *testing.T) {
		commands := Selection{}
		expected := Inputs{}
		actual := commands.Inputs()
		assert.Equal(t, expected, actual)
	})
	t.Run("no inputs", func(t *testing.T) {
		commands := Selection{
			Command{},
			Command{},
		}
		expected := Inputs{}
		actual := commands.Inputs()
		assert.Equal(t, expected, actual)
	})
	t.Run("merged inputs", func(t *testing.T) {
		commands := Selection{
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
		}
		expected := Inputs{
			Input{Name: "a", Value: &NumberValue{}},
			Input{Name: "b", Value: &NumberValue{}},
			Input{Name: "c", Value: &NumberValue{}},
		}
		actual := commands.Inputs()
		assert.Equal(t, expected, actual)
	})
}

func TestSelectionRenderScript(t *testing.T) {
	t.Run("empty", func(t *testing.T) {
		data := TemplateData{}
		commands := Selection{}
		_, err := commands.RenderScript(data)
		expected := "no script present"
		actual := fmt.Sprintf("%s", err)
		assert.Equal(t, expected, actual)
	})
	t.Run("template error", func(t *testing.T) {
		data := TemplateData{
			Input: map[string]any{
				"A": "a",
			},
		}
		commands := Selection{
			{
				Name: "foobar",
				Run:  "{{.Input.A}",
			},
		}
		_, err := commands.RenderScript(data)
		assert.ErrorContains(t, err, "template error: template: foobar:1: bad character U+007D '}'")
	})
	t.Run("script error", func(t *testing.T) {
		data := TemplateData{
			Input: map[string]any{
				"A": "a",
			},
		}
		commands := Selection{
			{
				Name: "foobar",
				Run:  "{{template \"foobaz\"}}",
			},
		}
		_, err := commands.RenderScript(data)
		assert.ErrorContains(t, err, "script error: template: foobar:1:11: executing ")
	})
	t.Run("render single template", func(t *testing.T) {
		data := TemplateData{
			Input: map[string]any{
				"A": "a",
				"B": "b",
			},
		}
		commands := Selection{
			{
				Name: "foobaz",
				Run:  "echo {{.Input.B}}",
			},
			{
				Name: "foobar",
				Run:  "echo {{.Input.A}}",
			},
		}
		expected := "echo a"
		actual, err := commands.RenderScript(data)
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
		commands := Selection{
			{
				Name: "foobaz",
				Run:  "echo {{.Input.B}}",
			},
			{
				Name: "foobar",
				Run:  "echo {{.Input.A}} {{template \"foobaz\" .}}",
			},
		}
		expected := "echo a echo b"
		actual, err := commands.RenderScript(data)
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
		commands := Selection{
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
		}
		expected := "echo a echo b"
		actual, err := commands.RenderScript(data)
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
		commands := Selection{
			{
				Name: "foobar",
				Run:  "echo {{input \"A\"}} {{env \"C\"}}",
			},
		}
		expected := "echo a c"
		actual, err := commands.RenderScript(data)
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
	commands := Selection{
		{
			Name: "foobar",
			Run:  "echo {{.Input.A}}",
		},
	}
	expected := "echo a"
	file, err := commands.RenderScriptToTemp(data)
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
		commands := Selection{
			{
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
			{
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
		}
		cmd, err := commands.Cmd(data, moreEnviron)
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
		commands := Selection{
			{
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
			{
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
		}
		cmd, err := commands.Cmd(data, moreEnviron)
		assert.NoError(t, err)
		assert.ElementsMatch(t, []string{"A=aa", "B=b", "C=c", "D=d"}, cmd.Env)
		assert.Equal(t, "/bin/bash", cmd.Path)
		assert.Equal(t, []string{"/bin/bash", "-x"}, cmd.Args[:len(cmd.Args)-1])
	})
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
