package main

import (
	"flag"
	"fmt"
	"os"
	"sort"
	"strings"
	"testing"
)

func TestCommandSetString_Empty(t *testing.T) {
	cs := CommandSet{}
	assertEqual(t, "", cs.String(), "CommandSet.String() to return an empty string")
}

func TestCommandSetString_One(t *testing.T) {
	cs := CommandSet{
		Commands: []ConfigCommand{
			{Name: "foobar"},
		},
	}
	assertEqual(t, "foobar", cs.String(), "CommandSet.String() to return expected value")
}

func TestCommandSetString_Many(t *testing.T) {
	cs := CommandSet{
		Commands: []ConfigCommand{
			{Name: ""},
			{Name: "foobar"},
			{Name: "foobaz"},
		},
	}
	assertEqual(t, "foobar foobaz", cs.String(), "CommandSet.String() to return expected value")
}

func TestCommandSetPure_Empty(t *testing.T) {
	cs := CommandSet{
		Commands: []ConfigCommand{},
	}
	assertEqual(t, false, cs.Pure(), "CommandSet.Pure() to return false")
}

func TestCommandSetPure_One(t *testing.T) {
	cs := CommandSet{
		Commands: []ConfigCommand{
			{Pure: true},
		},
	}
	assertEqual(t, true, cs.Pure(), "CommandSet.Pure() to return true")
}

func TestCommandSetPure_Many(t *testing.T) {
	cs := CommandSet{
		Commands: []ConfigCommand{
			{Pure: true},
			{Pure: false},
		},
	}
	assertEqual(t, false, cs.Pure(), "CommandSet.Pure() to return false")
}

func TestCommandSetShell_Default(t *testing.T) {
	cs := CommandSet{
		Commands: []ConfigCommand{
			{Shell: []string{}},
			{Shell: []string{}},
		},
	}
	assertDeepEqual(t, DefaultShell, cs.Shell(), "CommandSet.Shell() to return default")
}

func TestCommandSetShell_Latest(t *testing.T) {
	cs := CommandSet{
		Commands: []ConfigCommand{
			{Shell: []string{"foobaz"}},
			{Shell: []string{"foobar"}},
		},
	}
	assertDeepEqual(t, []string{"foobar"}, cs.Shell(), "CommandSet.Shell() to return the latest entry")
}

func TestCommandSetShell_Parent(t *testing.T) {
	cs := CommandSet{
		Commands: []ConfigCommand{
			{Shell: []string{"foobaz"}},
			{Shell: []string{}},
		},
	}
	assertDeepEqual(t, []string{"foobaz"}, cs.Shell(), "CommandSet.Shell() to return the parent's entry")
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
	assertDeepEqual(t, expected, cs.Inputs(), "CommandSet.Inputs() returned unexpected results")
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
	expected := map[string]string{
		"A": "aa",
		"B": "b",
		"C": "c",
	}
	assertDeepEqual(t, expected, cs.Env(), "CommandSet.Env() returned unexpected results")
}

func TestCommandSetRenderEnv_NonError(t *testing.T) {
	data := map[string]any{
		"A": "a",
		"B": "b",
		"C": "c",
	}
	cs := CommandSet{
		Commands: []ConfigCommand{
			{
				Env: map[string]string{
					"A": "{{.A}}",
					"B": "{{.B}}",
				},
			},
			{
				Env: map[string]string{
					"A": "aa",
					"C": "{{.C}}",
				},
			},
		},
	}
	expected := []string{
		"A=aa",
		"B=b",
		"C=c",
	}
	actual, err := cs.RenderEnv(data)
	if err != nil {
		t.Fatalf("CommandSet.RenderEnv() returned an unexpected error: %v", err)
	}
	sort.Strings(actual)
	assertDeepEqual(t, expected, actual, "CommandSet.RenderEnv() returned unexpected results")
}

func TestCommandSetRenderEnv_TemplateError(t *testing.T) {
	data := map[string]any{
		"A": "a",
	}
	cs := CommandSet{
		Commands: []ConfigCommand{
			{
				Env: map[string]string{
					"A": "{{.A}",
				},
			},
		},
	}
	_, err := cs.RenderEnv(data)
	expected := "template error for environment variable: 'A' - template: :1: bad character U+007D '}'"
	actual := fmt.Sprintf("%s", err)
	assertEqual(t, expected, actual, "CommandSet.RenderEnv() returned unexpected error")
}

func TestCommandSetRenderScript_Empty(t *testing.T) {
	data := map[string]any{}
	cs := CommandSet{}
	_, err := cs.RenderScript(data)
	expected := "no script present"
	actual := fmt.Sprintf("%s", err)
	assertEqual(t, expected, actual, "CommandSet.RenderScript() returned unexpected error")
}

func TestCommandSetRenderScript_TemplateError(t *testing.T) {
	data := map[string]any{
		"A": "a",
	}
	cs := CommandSet{
		Commands: []ConfigCommand{
			{
				Name: "foobar",
				Run:  "{{.A}",
			},
		},
	}
	_, err := cs.RenderScript(data)
	expected := "template error for command: 'foobar' - template: :1: bad character U+007D '}'"
	actual := fmt.Sprintf("%s", err)
	assertEqual(t, expected, actual, "CommandSet.RenderScript() returned unexpected error")
}

func TestCommandSetRenderScript_NonError(t *testing.T) {
	data := map[string]any{
		"A": "a",
		"B": "b",
	}
	cs := CommandSet{
		Commands: []ConfigCommand{
			{
				Name: "foobaz",
				Run:  "echo {{.B}}",
			},
			{
				Name: "foobar",
				Run:  "echo {{.A}}",
			},
		},
	}
	expected := "echo a"
	actual, err := cs.RenderScript(data)
	if err != nil {
		t.Fatalf("CommandSet.RenderScript() returned an unexpected error: %v", err)
	}
	assertEqual(t, expected, actual, "CommandSet.RenderScript() returned unexpected result")
}

func TestCommandSetRenderScriptToTemp(t *testing.T) {
	data := map[string]any{
		"A": "a",
	}
	cs := CommandSet{
		Commands: []ConfigCommand{
			{
				Name: "foobar",
				Run:  "echo {{.A}}",
			},
		},
	}
	expected := "echo a"
	file, err := cs.RenderScriptToTemp(data)
	if err != nil {
		t.Fatalf("CommandSet.RenderScriptToTemp() returned an unexpected error: %v", err)
	}
	actual, err := readTextFile(file)
	if err != nil {
		t.Fatalf("Could not read file containing rendered script: %v", err)
	}
	assertEqual(t, expected, actual, "CommandSet.RenderScriptToTemp() returned unexpected result")
}

func TestCommandSetCmd_IsPure(t *testing.T) {
	data := map[string]any{
		"A": "a",
		"B": "b",
		"C": "c",
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
					"A": "{{.A}}",
					"B": "{{.B}}",
				},
			},
			{
				Shell: []string{"/bin/bash", "-x"},
				Run:   "foobar",
				Env: map[string]string{
					"A": "aa",
					"C": "{{.C}}",
				},
				Pure: true,
			},
		},
	}
	cmd, err := cs.Cmd(data, moreEnviron)
	if err != nil {
		t.Fatalf("CommandSet.Cmd() returned an unexpected error: %v", err)
	}
	env := cmd.Env
	sort.Strings(env)
	assertDeepEqual(t, []string{"A=aa", "B=b", "C=c"}, env, "CommandSet.Cmd() did not set cmd.Env with correct values")
	assertEqual(t, "/bin/bash", cmd.Path, "CommandSet.Cmd() did not set cmd.Path to the shell path")
	assertDeepEqual(t, []string{"/bin/bash", "-x"}, cmd.Args[:len(cmd.Args)-1], "CommandSet.Cmd() did not set cmd.Args with the correct values")
}

func TestCommandSetCmd_NotPure(t *testing.T) {
	data := map[string]any{
		"A": "a",
		"B": "b",
		"C": "c",
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
					"A": "{{.A}}",
					"B": "{{.B}}",
				},
			},
			{
				Shell: []string{"/bin/bash", "-x"},
				Run:   "foobar",
				Env: map[string]string{
					"A": "aa",
					"C": "{{.C}}",
				},
			},
		},
	}
	cmd, err := cs.Cmd(data, moreEnviron)
	if err != nil {
		t.Fatalf("CommandSet.Cmd() returned an unexpected error: %v", err)
	}
	env := cmd.Env
	sort.Strings(env)
	assertDeepEqual(t, []string{"A=aa", "B=b", "C=c", "D=d"}, env, "CommandSet.Cmd() did not set cmd.Env with correct values")
	assertEqual(t, "/bin/bash", cmd.Path, "CommandSet.Cmd() did not set cmd.Path to the shell path")
	assertDeepEqual(t, []string{"/bin/bash", "-x"}, cmd.Args[:len(cmd.Args)-1], "CommandSet.Cmd() did not set cmd.Args with the correct values")
}

func TestCommandSetParseArgs_Valid(t *testing.T) {
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
	if err := cs.ParseArgs(&actual); err != nil {
		t.Fatalf("CommandSet.ParseArgs() returned unexpected error: %v", err)
	}
	assertDeepEqual(t, expected, actual, "CommandSet.ParseArgs() returned unexpected results")
}

func TestCommandSetParseArgs_Help(t *testing.T) {
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
	assertDeepEqual(t, flag.ErrHelp, actual, "CommandSet.ParseArgs() did not acknowledge help")
}

func TestNewCommandSet_PreselectedFromArgs(t *testing.T) {
	config := Config{
		Commands: ConfigCommands{
			ConfigCommand{Name: "foobar", Run: "true"},
		},
	}
	cs, err := NewCommandSet(config, []string{"foobar", "-a", "1"})
	if err != nil {
		t.Fatalf("NewCommandSet() returned unexpected error: %v", err)
	}
	assertDeepEqual(t, cs.Config, config, "NewCommandSet() did not set config")
	assertDeepEqual(t, cs.Args, []string{"-a", "1"}, "NewCommandSet() has unexpected args")
}

func TestNewCommandSet_Invalid(t *testing.T) {
	config := Config{
		Commands: ConfigCommands{
			ConfigCommand{Name: "foobar", Run: "true"},
		},
	}
	_, err := NewCommandSet(config, []string{"foobaz", "-a", "1"})
	expected := "invalid subcommand: foobaz"
	actual := fmt.Sprintf("%s", err)
	assertEqual(t, expected, actual, "NewCommandSet() returned unexpected error")
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
