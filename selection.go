package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"text/template"
)

type Selection []Command

func (commands Selection) String() string {
	var names []string
	for _, command := range commands {
		if command.Name != "" {
			names = append(names, command.Name)
		}
	}
	return strings.Join(names, " ")
}

func (commands Selection) Runnable() bool {
	for i := len(commands) - 1; i >= 0; {
		return commands[i].Runnable()
	}
	return false
}

func (commands Selection) Description() string {
	for i := len(commands) - 1; i >= 0; {
		return commands[i].Description
	}
	return ""
}

func (commands Selection) Run() string {
	for i := len(commands) - 1; i >= 0; {
		return commands[i].Run
	}
	return ""
}

func (commands Selection) Shell() []string {
	for i := len(commands) - 1; i >= 0; i-- {
		command := commands[i]
		if len(command.Shell) > 0 {
			return command.Shell
		}
	}
	return DefaultShell
}

func (commands Selection) Env() EnvMap {
	env := EnvMap{}
	for _, command := range commands {
		env = env.Merge(command.Env)
	}
	return env
}

func (commands Selection) Pure() bool {
	for i := len(commands) - 1; i >= 0; {
		return commands[i].Pure
	}
	return false
}

func (commands Selection) Inputs() Inputs {
	inputs := Inputs{}
	for _, command := range commands {
		inputs = inputs.Merge(command.Inputs)
	}
	return inputs
}

func (commands Selection) Commands() SubCommands {
	for i := len(commands) - 1; i >= 0; {
		return commands[i].Commands
	}
	return SubCommands{}
}

func (commands Selection) RenderScript(data TemplateData) (string, error) {
	var tmpl *template.Template
	for _, command := range commands {
		var err error
		if command.Run == "" {
			continue
		}
		if tmpl == nil {
			tmpl = template.New(command.Name).Funcs(defaultTemplateFuncs)
		} else {
			tmpl = tmpl.New(command.Name)
		}
		tmpl, err = tmpl.Funcs(data.Funcs()).Parse(command.Run)
		if err != nil {
			return "", fmt.Errorf("template error: %v", err)
		}
	}
	if tmpl == nil {
		return "", fmt.Errorf("no script present")
	}
	if script, err := RenderTemplate(tmpl, data); err != nil {
		return script, fmt.Errorf("script error: %v", err)
	} else {
		return script, err
	}
}

func (commands Selection) RenderScriptToTemp(data TemplateData) (string, error) {
	var file *os.File
	script, err := commands.RenderScript(data)
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
	logger.Printf("Created temporary script: %s\n", file.Name())
	return file.Name(), nil
}

func (commands Selection) RenderEnv(data TemplateData) (EnvMap, error) {
	env := commands.Env()
	for name, template := range env {
		if value, err := RenderTemplate(template, data); err != nil {
			return env, fmt.Errorf("template error for environment variable: '%s' - %v", name, err)
		} else {
			env[name] = value
		}
	}
	return env, nil
}

func (commands Selection) Cmd(data TemplateData, moreEnv EnvMap) (*exec.Cmd, error) {
	var scriptFile string
	var env EnvMap
	var err error
	shell := commands.Shell()
	scriptFile, err = commands.RenderScriptToTemp(data)
	if err != nil {
		return nil, err
	}
	env, err = commands.RenderEnv(data)
	if err != nil {
		return nil, err
	}
	if !commands.Pure() {
		env = moreEnv.Merge(env)
	}
	shell = append(shell, scriptFile)
	cmd := exec.Command(shell[0], shell[1:]...)
	cmd.Env = env.ToList()
	return cmd, nil
}

func (commands Selection) ToArgs() []string {
	inputArgs := commands.Inputs().ToArgs()
	args := make([]string, 0, len(commands)+len(inputArgs))
	for _, command := range commands {
		if command.Name != "" {
			args = append(args, command.Name)
		}
	}
	return append(args, inputArgs...)
}
