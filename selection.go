package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"text/template"
)

type Selection struct {
	commands []Command
	Args     []string
}

func (selection Selection) Select(args []string) Selection {
	if len(args) > 0 {
		commands := selection.commands
		if subcommand, found := commands[len(commands)-1].Get(args[0]); found {
			selection = selection.SelectCommand(subcommand.Command, args[1:])
			return selection.Select(selection.Args)
		}
	}
	newSelection := NewSelection(selection.commands[0], selection.commands[1:]...)
	newSelection.Args = args
	return newSelection
}

func (selection Selection) SelectCommand(command Command, args []string) Selection {
	commands := selection.commands
	newSelection := NewSelection(commands[0], append(commands[1:], command)...)
	newSelection.Args = args
	return newSelection
}

func (selection Selection) String() string {
	var names []string
	for _, command := range selection.commands {
		if command.Name != "" {
			names = append(names, command.Name)
		}
	}
	return strings.Join(names, " ")
}

func (selection Selection) Runnable() bool {
	return selection.commands[len(selection.commands)-1].Runnable()
}

func (selection Selection) Description() string {
	return selection.commands[len(selection.commands)-1].Description
}

func (selection Selection) Run() string {
	return selection.commands[len(selection.commands)-1].Run
}

func (selection Selection) Shell() []string {
	for i := len(selection.commands) - 1; i >= 0; i-- {
		command := selection.commands[i]
		if len(command.Shell) > 0 {
			return command.Shell
		}
	}
	return DefaultShell
}

func (selection Selection) Env() EnvMap {
	env := EnvMap{}
	for _, command := range selection.commands {
		env = env.Merge(command.Env)
	}
	return env
}

func (selection Selection) Pure() bool {
	return selection.commands[len(selection.commands)-1].Pure
}

func (selection Selection) Inputs() Inputs {
	inputs := Inputs{}
	for _, command := range selection.commands {
		inputs = inputs.Merge(command.Inputs)
	}
	return inputs
}

func (selection Selection) Commands() SubCommands {
	return selection.commands[len(selection.commands)-1].Commands
}

func (selection Selection) RenderScript(data TemplateData) (string, error) {
	var tmpl *template.Template
	for _, command := range selection.commands {
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

func (selection Selection) RenderScriptToTemp(data TemplateData) (string, error) {
	var file *os.File
	script, err := selection.RenderScript(data)
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

func (selection Selection) RenderEnv(data TemplateData) (EnvMap, error) {
	env := selection.Env()
	for name, template := range env {
		if value, err := RenderTemplate(template, data); err != nil {
			return env, fmt.Errorf("template error for environment variable: '%s' - %v", name, err)
		} else {
			env[name] = value
		}
	}
	return env, nil
}

func (selection Selection) Cmd(data TemplateData, moreEnv EnvMap) (*exec.Cmd, error) {
	var scriptFile string
	var env EnvMap
	var err error
	shell := selection.Shell()
	scriptFile, err = selection.RenderScriptToTemp(data)
	if err != nil {
		return nil, err
	}
	env, err = selection.RenderEnv(data)
	if err != nil {
		return nil, err
	}
	if !selection.Pure() {
		env = moreEnv.Merge(env)
	}
	shell = append(shell, scriptFile)
	cmd := exec.Command(shell[0], shell[1:]...)
	cmd.Env = env.ToList()
	return cmd, nil
}

func (selection Selection) ToArgs() []string {
	inputArgs := selection.Inputs().ToArgs()
	args := make([]string, 0, len(selection.commands)+len(inputArgs))
	for _, command := range selection.commands {
		if command.Name != "" {
			args = append(args, command.Name)
		}
	}
	return append(args, inputArgs...)
}

func NewSelection(command Command, commands ...Command) Selection {
	return Selection{commands: append([]Command{command}, commands...), Args: []string{}}
}
