package main

import (
	"bytes"
	"os"
	"testing"
	"text/template"
)

func TestExecuteCommand(t *testing.T) {
	cfg := &Config{
		Commands: map[string]Command{
			"test": {
				Run: "echo {{ .Input.name }}",
			},
		},
	}
	cmd := cfg.Commands["test"]
	inputs := map[string]any{
		"name": "World",
	}

	// Redirect stdout to a buffer to capture the output
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := executeCommand(cfg, &cmd, inputs)

	w.Close()
	os.Stdout = oldStdout

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	if output != "World\n" {
		t.Errorf("expected output 'World\\n', got '%s'", output)
	}
}

func TestBuildTemplate(t *testing.T) {
	cfg := &Config{
		Commands: map[string]Command{
			"parent": {
				Run: "parent_run",
			},
			"child": {
				Run: "{{ template \"parent\" }} child_run",
			},
		},
	}
	cmd := cfg.Commands["child"]
	inputs := map[string]any{}

	tmpl, err := buildTemplate(cfg, &cmd, inputs)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	var buf bytes.Buffer
	err = tmpl.Execute(&buf, nil)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	expected := "parent_run child_run"
	if buf.String() != expected {
		t.Errorf("expected '%s', got '%s'", expected, buf.String())
	}
}

func TestTemplateFuncs(t *testing.T) {
	inputs := map[string]any{
		"myinput": "myvalue",
	}
	os.Setenv("MYENV", "myenvvalue")
	defer os.Unsetenv("MYENV")

	tmpl := template.New("test").Funcs(template.FuncMap{
		"input": func(name string) (any, error) {
			return inputs[name], nil
		},
		"env": func(name string) string {
			return os.Getenv(name)
		},
	})

	// Test input function
	result := bytes.Buffer{}
	tmpl.Parse(`{{ input "myinput" }}`)
	tmpl.Execute(&result, nil)
	if result.String() != "myvalue" {
		t.Errorf("expected 'myvalue', got '%s'", result.String())
	}

	// Test env function
	result.Reset()
	tmpl.Parse(`{{ env "MYENV" }}`)
	tmpl.Execute(&result, nil)
	if result.String() != "myenvvalue" {
		t.Errorf("expected 'myenvvalue', got '%s'", result.String())
	}
}
