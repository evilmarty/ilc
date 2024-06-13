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
	_, err := cs.RenderEnv(data)
	expected := "template error for environment variable: 'A' - template: :1: bad character U+007D '}'"
	actual := fmt.Sprintf("%s", err)
	assertEqual(t, expected, actual, "CommandSet.RenderEnv() returned unexpected error")
}

func TestCommandSetRenderScript_Empty(t *testing.T) {
	data := TemplateData{}
	cs := CommandSet{}
	_, err := cs.RenderScript(data)
	expected := "no script present"
	actual := fmt.Sprintf("%s", err)
	assertEqual(t, expected, actual, "CommandSet.RenderScript() returned unexpected error")
}

func TestCommandSetRenderScript_TemplateError(t *testing.T) {
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
	expected := "template error for command: 'foobar' - template: :1: bad character U+007D '}'"
	actual := fmt.Sprintf("%s", err)
	assertEqual(t, expected, actual, "CommandSet.RenderScript() returned unexpected error")
}

func TestCommandSetRenderScript_NonError(t *testing.T) {
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
	assertEqual(t, expected, actual, "CommandSet.RenderScript() returned unexpected result")
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
	if err != nil {
		t.Fatalf("Could not read file containing rendered script: %v", err)
	}
	assertEqual(t, expected, actual, "CommandSet.RenderScriptToTemp() returned unexpected result")
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
			},
			{
				Shell: []string{"/bin/bash", "-x"},
				Run:   "foobar",
				Env: map[string]string{
					"A": "aa",
					"C": "{{.Input.C}}",
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
			},
			{
				Shell: []string{"/bin/bash", "-x"},
				Run:   "foobar",
				Env: map[string]string{
					"A": "aa",
					"C": "{{.Input.C}}",
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
	assertDeepEqual(t, expected, actual, "CommandSet.ParseEnv() returned unexpected results")
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
	assertEqual(t, nil, actual, "CommandSet.Validate() returned unexpected error")
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
	err := cs.Validate(map[string]any{})
	expected := "missing input: A"
	actual := fmt.Sprintf("%s", err)
	assertEqual(t, expected, actual, "CommandSet.Validate() returned unexpected error")
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
	err := cs.Validate(map[string]any{"A": "123"})
	expected := "invalid input: A"
	actual := fmt.Sprintf("%s", err)
	assertEqual(t, expected, actual, "CommandSet.Validate() returned unexpected error")
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
	assertEqual(t, nil, actual, "CommandSet.Validate() returned unexpected error")
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
	assertEqual(t, true, cs.Runnable(), "CommandSet.Runnable() expected to return true")
}

func TestCommandSetRunnable_False(t *testing.T) {
	cs := CommandSet{
		Commands: []ConfigCommand{
			{
				Run: "",
			},
		},
	}
	assertEqual(t, false, cs.Runnable(), "CommandSet.Runnable() expected to return false")
}

func TestCommandSetSelected_True(t *testing.T) {
	cs := CommandSet{
		Commands: []ConfigCommand{
			{},
		},
	}
	assertEqual(t, true, cs.Selected(), "CommandSet.Selected() expected to return true")
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
	assertEqual(t, false, cs.Selected(), "CommandSet.Selected() expected to return false")
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
