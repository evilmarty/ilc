package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

var defaultShell = []string{"/bin/sh"}

type CommandChain []ConfigCommand

func (cc CommandChain) Name() string {
	var names []string
	for _, c := range cc {
		if c.Name != "" {
			names = append(names, c.Name)
		}
	}
	return strings.Join(names, " ")
}

func (cc CommandChain) Pure() bool {
	for i := len(cc) - 1; i >= 0; {
		command := cc[i]
		return command.Pure
	}
	return false
}

func (cc CommandChain) Shell() []string {
	for i := len(cc) - 1; i >= 0; i-- {
		shell := cc[i].Shell
		if len(shell) > 0 {
			return shell
		}
	}
	return []string{}
}

func (cc CommandChain) Inputs() []ConfigInput {
	inputs := make([]ConfigInput, 0, len(cc))
	for _, command := range cc {
		inputs = append(inputs, command.Inputs...)
	}
	return inputs
}

func (cc CommandChain) Env() map[string]string {
	envSize := 0
	for _, command := range cc {
		envSize = envSize + len(command.Env)
	}
	envs := make(map[string]string, envSize)
	for _, command := range cc {
		for name, value := range command.Env {
			envs[name] = value
		}
	}
	return envs
}

func (cc CommandChain) RenderEnv(data map[string]any) ([]string, error) {
	var renderedEnvs []string
	for name, template := range cc.Env() {
		if value, err := RenderTemplate(template, data); err != nil {
			return renderedEnvs, fmt.Errorf("template error for environment variable: '%s' - %v", name, err)
		} else {
			renderedEnvs = append(renderedEnvs, fmt.Sprintf("%s=%s", name, value))
		}
	}
	return renderedEnvs, nil
}

func (cc CommandChain) RenderScript(data map[string]any) (string, error) {
	for i := len(cc) - 1; i >= 0; {
		command := cc[i]
		if script, err := RenderTemplate(command.Run, data); err != nil {
			return script, fmt.Errorf("template error for command: '%s' - %v", command.Name, err)
		} else {
			return script, nil
		}
	}
	return "", fmt.Errorf("no script present")
}

func (cc CommandChain) RenderScriptToTemp(data map[string]any) (string, error) {
	var file *os.File
	script, err := cc.RenderScript(data)
	if err != nil {
		return "", err
	}
	file, err = os.CreateTemp("", "*")
	if err != nil {
		return "", fmt.Errorf("failed to create temporary script file: %v", err)
	}
	if _, err := file.Write([]byte(script)); err != nil {
		return "", fmt.Errorf("failed to write to temporary script file: %v", err)
	}
	if err := file.Close(); err != nil {
		return "", fmt.Errorf("failed to close temporary script file: %v", err)
	}
	return file.Name(), nil
}

func (cc CommandChain) Cmd(data map[string]any) (*exec.Cmd, error) {
	var scriptFile string
	var env []string
	var err error
	shell := cc.Shell()
	if len(shell) == 0 {
		shell = defaultShell[:]
	}
	scriptFile, err = cc.RenderScriptToTemp(data)
	if err != nil {
		return nil, err
	}
	env, err = cc.RenderEnv(data)
	if err != nil {
		return nil, err
	}
	if !cc.Pure() {
		env = append(os.Environ(), env...)
	}
	shell = append(shell, scriptFile)
	cmd := exec.Command(shell[0], shell[1:]...)
	cmd.Env = env
	return cmd, nil
}

func (cc CommandChain) Run(data map[string]any) error {
	if cmd, err := cc.Cmd(data); err != nil {
		return err
	} else {
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		return cmd.Run()
	}
}
