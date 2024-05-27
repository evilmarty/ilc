package main

import (
	"fmt"
	"os"
	"sort"
	"strings"
	"testing"
)

func TestCommandChainPure_Empty(t *testing.T) {
	commandChain := CommandChain{}
	assertEqual(t, false, commandChain.Pure(), "CommandChain.Pure() to return false")
}

func TestCommandChainPure_NotEmpty(t *testing.T) {
	commandChain := CommandChain{
		ConfigCommand{Pure: true},
	}
	assertEqual(t, true, commandChain.Pure(), "CommandChain.Pure() to return true")
}

func TestCommandChainShell_Latest(t *testing.T) {
	commandChain := CommandChain{
		ConfigCommand{Shell: []string{"foobaz"}},
		ConfigCommand{Shell: []string{"foobar"}},
	}
	assertDeepEqual(t, []string{"foobar"}, commandChain.Shell(), "CommandChain.Shell() to return the latest entry")
}

func TestCommandChainShell_Parent(t *testing.T) {
	commandChain := CommandChain{
		ConfigCommand{Shell: []string{"foobaz"}},
		ConfigCommand{Shell: []string{}},
	}
	assertDeepEqual(t, []string{"foobaz"}, commandChain.Shell(), "CommandChain.Shell() to return the parent's entry")
}

func TestCommandChainInputs(t *testing.T) {
	commandChain := CommandChain{
		ConfigCommand{
			Inputs: ConfigInputs{
				ConfigInput{Name: "A"},
				ConfigInput{Name: "B"},
			},
		},
		ConfigCommand{
			Inputs: ConfigInputs{
				ConfigInput{Name: "C"},
			},
		},
	}
	expected := []ConfigInput{
		commandChain[0].Inputs[0],
		commandChain[0].Inputs[1],
		commandChain[1].Inputs[0],
	}
	assertDeepEqual(t, expected, commandChain.Inputs(), "CommandChain.Inputs() returned unexpected results")
}

func TestCommandChainEnv(t *testing.T) {
	commandChain := CommandChain{
		ConfigCommand{
			Env: map[string]string{
				"A": "a",
				"B": "b",
			},
		},
		ConfigCommand{
			Env: map[string]string{
				"A": "aa",
				"C": "c",
			},
		},
	}
	expected := map[string]string{
		"A": "aa",
		"B": "b",
		"C": "c",
	}
	assertDeepEqual(t, expected, commandChain.Env(), "CommandChain.Env() returned unexpected results")
}

func TestCommandChainRenderEnv_NonError(t *testing.T) {
	data := map[string]any{
		"A": "a",
		"B": "b",
		"C": "c",
	}
	commandChain := CommandChain{
		ConfigCommand{
			Env: map[string]string{
				"A": "{{.A}}",
				"B": "{{.B}}",
			},
		},
		ConfigCommand{
			Env: map[string]string{
				"A": "aa",
				"C": "{{.C}}",
			},
		},
	}
	expected := []string{
		"A=aa",
		"B=b",
		"C=c",
	}
	actual, err := commandChain.RenderEnv(data)
	if err != nil {
		t.Fatalf("CommandChain.RenderEnv() returned an unexpected error: %v", err)
	}
	sort.Strings(actual)
	assertDeepEqual(t, expected, actual, "CommandChain.RenderEnv() returned unexpected results")
}

func TestCommandChainRenderEnv_TemplateError(t *testing.T) {
	data := map[string]any{
		"A": "a",
	}
	commandChain := CommandChain{
		ConfigCommand{
			Env: map[string]string{
				"A": "{{.A}",
			},
		},
	}
	_, err := commandChain.RenderEnv(data)
	expected := "template error for environment variable: 'A' - template: :1: bad character U+007D '}'"
	actual := fmt.Sprintf("%s", err)
	assertEqual(t, expected, actual, "CommandChain.RenderEnv() returned unexpected error")
}

func TestCommandChainRenderScript_Empty(t *testing.T) {
	data := map[string]any{}
	commandChain := CommandChain{}
	_, err := commandChain.RenderScript(data)
	expected := "no script present"
	actual := fmt.Sprintf("%s", err)
	assertEqual(t, expected, actual, "CommandChain.RenderScript() returned unexpected error")
}

func TestCommandChainRenderScript_TemplateError(t *testing.T) {
	data := map[string]any{
		"A": "a",
	}
	commandChain := CommandChain{
		ConfigCommand{
			Name: "foobar",
			Run:  "{{.A}",
		},
	}
	_, err := commandChain.RenderScript(data)
	expected := "template error for command: 'foobar' - template: :1: bad character U+007D '}'"
	actual := fmt.Sprintf("%s", err)
	assertEqual(t, expected, actual, "CommandChain.RenderScript() returned unexpected error")
}

func TestCommandChainRenderScript_NonError(t *testing.T) {
	data := map[string]any{
		"A": "a",
		"B": "b",
	}
	commandChain := CommandChain{
		ConfigCommand{
			Name: "foobaz",
			Run:  "echo {{.B}}",
		},
		ConfigCommand{
			Name: "foobar",
			Run:  "echo {{.A}}",
		},
	}
	expected := "echo a"
	actual, err := commandChain.RenderScript(data)
	if err != nil {
		t.Fatalf("CommandChain.RenderScript() returned an unexpected error: %v", err)
	}
	assertEqual(t, expected, actual, "CommandChain.RenderScript() returned unexpected result")
}

func TestCommandChainRenderScriptToTemp(t *testing.T) {
	data := map[string]any{
		"A": "a",
	}
	commandChain := CommandChain{
		ConfigCommand{
			Name: "foobar",
			Run:  "echo {{.A}}",
		},
	}
	expected := "echo a"
	file, err := commandChain.RenderScriptToTemp(data)
	if err != nil {
		t.Fatalf("CommandChain.RenderScriptToTemp() returned an unexpected error: %v", err)
	}
	actual, err := readTextFile(file)
	if err != nil {
		t.Fatalf("Could not read file containing rendered script: %v", err)
	}
	assertEqual(t, expected, actual, "CommandChain.RenderScriptToTemp() returned unexpected result")
}

func TestCommandChainCmd(t *testing.T) {
	data := map[string]any{
		"A": "a",
		"B": "b",
		"C": "c",
	}
	commandChain := CommandChain{
		ConfigCommand{
			Shell: []string{"/bin/sh"},
			Run:   "foobaz",
			Env: map[string]string{
				"A": "{{.A}}",
				"B": "{{.B}}",
			},
		},
		ConfigCommand{
			Shell: []string{"/bin/bash", "-x"},
			Run:   "foobar",
			Env: map[string]string{
				"A": "aa",
				"C": "{{.C}}",
			},
			Pure: true,
		},
	}
	cmd, err := commandChain.Cmd(data)
	if err != nil {
		t.Fatalf("CommandChain.Cmd() returned an unexpected error: %v", err)
	}
	env := cmd.Env
	sort.Strings(env)
	assertDeepEqual(t, []string{"A=aa", "B=b", "C=c"}, env, "CommandChain.Cmd() did not set cmd.Env with correct values")
	assertEqual(t, "/bin/bash", cmd.Path, "CommandChain.Cmd() did not set cmd.Path to the shell path")
	assertDeepEqual(t, []string{"/bin/bash", "-x"}, cmd.Args[:len(cmd.Args)-1], "CommandChain.Cmd() did not set cmd.Args with the correct values")
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
