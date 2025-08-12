package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"text/template"
)

func executeCommand(cfg *Config, cmd *Command, inputs map[string]any) error {
	tmpl, err := buildTemplate(cfg, cmd, inputs)
	if err != nil {
		return err
	}

	env := getEnv()
	for k, v := range cmd.Env {
		tmpl, err := template.New(k).Parse(v)
		if err != nil {
			return fmt.Errorf("error parsing env template: %w", err)
		}
		data := struct {
			Input map[string]any
			Env   map[string]string
		}{
			Input: inputs,
			Env:   env,
		}
		var buf bytes.Buffer
		if err := tmpl.Execute(&buf, data); err != nil {
			return fmt.Errorf("error executing env template: %w", err)
		}
		env[k] = buf.String()
	}

	data := struct {
		Input map[string]any
		Env   map[string]string
	}{
		Input: inputs,
		Env:   env,
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return fmt.Errorf("error executing template: %w", err)
	}

	command := buf.String()

	shell := cfg.Shell
	if len(shell) == 0 {
		shell = []string{"/bin/sh"}
	}

	c := exec.Command(shell[0], append(shell[1:], "-c", command)...)

	if !cmd.Pure {
		c.Env = os.Environ()
	}

	for k, v := range env {
		c.Env = append(c.Env, fmt.Sprintf("%s=%s", k, v))
	}

	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	c.Stdin = os.Stdin

	if err := c.Run(); err != nil {
		return err
	}

	return recordHistory(cfg.FilePath, command)
}

func buildTemplate(cfg *Config, cmd *Command, inputs map[string]any) (*template.Template, error) {
	tmpl := template.New("main")

	funcs := template.FuncMap{
		"input": func(name string) (any, error) {
			if val, ok := inputs[name]; ok {
				return val, nil
			}
			return nil, fmt.Errorf("input not found: %s", name)
		},
		"env": func(name string) string {
			return os.Getenv(name)
		},
	}

	tmpl.Funcs(funcs)

	// Add parent templates
	var addTemplates func(commands map[string]Command)
	addTemplates = func(commands map[string]Command) {
		for name, command := range commands {
			if command.Run != "" {
				_, err := tmpl.New(name).Parse(command.Run)
				if err != nil {
					// panic to avoid complexity of returning error
					panic(err)
				}
			}
			if len(command.Commands) > 0 {
				addTemplates(command.Commands)
			}
		}
	}

	addTemplates(cfg.Commands)

	// Parse the final command
	_, err := tmpl.Parse(cmd.Run)
	if err != nil {
		return nil, fmt.Errorf("error parsing template: %w", err)
	}

	return tmpl, nil
}

func getEnv() map[string]string {
	envs := make(map[string]string)
	for _, env := range os.Environ() {
		pair := strings.SplitN(env, "=", 2)
		envs[pair[0]] = pair[1]
	}
	return envs
}
