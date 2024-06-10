package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

var (
	DefaultShell = []string{"/bin/sh"}
	EnvVarPrefix = "ILC_INPUT_"
)

type CommandSet struct {
	Config        Config
	Commands      []ConfigCommand
	Args          []string
	ErrorHandling flag.ErrorHandling
}

func (cs CommandSet) String() string {
	var names []string
	for _, c := range cs.Commands {
		if c.Name != "" {
			names = append(names, c.Name)
		}
	}
	return strings.Join(names, " ")
}

func (cs CommandSet) Description() string {
	for i := len(cs.Commands) - 1; i >= 0; {
		command := cs.Commands[i]
		return command.Description
	}
	return ""
}

func (cs CommandSet) Subcommands() []ConfigCommand {
	for i := len(cs.Commands) - 1; i >= 0; {
		command := cs.Commands[i]
		return command.Commands
	}
	return []ConfigCommand{}
}

func (cs CommandSet) Pure() bool {
	for i := len(cs.Commands) - 1; i >= 0; {
		command := cs.Commands[i]
		return command.Pure
	}
	return false
}

func (cs CommandSet) Shell() []string {
	for i := len(cs.Commands) - 1; i >= 0; i-- {
		shell := cs.Commands[i].Shell
		if len(shell) > 0 {
			return shell
		}
	}
	return DefaultShell
}

func (cs CommandSet) Inputs() []ConfigInput {
	inputs := make([]ConfigInput, 0, len(cs.Commands))
	for _, command := range cs.Commands {
		inputs = append(inputs, command.Inputs...)
	}
	return inputs
}

func (cs CommandSet) Env() map[string]string {
	envs := make(map[string]string)
	for _, command := range cs.Commands {
		for name, value := range command.Env {
			envs[name] = value
		}
	}
	return envs
}

func (cs CommandSet) RenderEnv(data TemplateData) ([]string, error) {
	var renderedEnvs []string
	for name, template := range cs.Env() {
		if value, err := RenderTemplate(template, data); err != nil {
			return renderedEnvs, fmt.Errorf("template error for environment variable: '%s' - %v", name, err)
		} else {
			renderedEnvs = append(renderedEnvs, fmt.Sprintf("%s=%s", name, value))
		}
	}
	return renderedEnvs, nil
}

func (cs CommandSet) RenderScript(data TemplateData) (string, error) {
	for i := len(cs.Commands) - 1; i >= 0; {
		command := cs.Commands[i]
		if script, err := RenderTemplate(command.Run, data); err != nil {
			return script, fmt.Errorf("template error for command: '%s' - %v", command.Name, err)
		} else {
			return script, nil
		}
	}
	return "", fmt.Errorf("no script present")
}

func (cs CommandSet) RenderScriptToTemp(data TemplateData) (string, error) {
	var file *os.File
	script, err := cs.RenderScript(data)
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

func (cs CommandSet) ParseArgs(values *map[string]any) error {
	fs := flag.NewFlagSet(cs.String(), flag.ContinueOnError)
	fs.Usage = func() {
		// Don't do anything here. We just want the error.
	}
	for _, input := range cs.Inputs() {
		fs.String(input.Name, input.DefaultValue, input.Description)
	}
	if err := fs.Parse(cs.Args); err != nil {
		return err
	}
	fs.Visit(func(f *flag.Flag) {
		if v, ok := f.Value.(flag.Getter); ok {
			(*values)[f.Name] = v.Get()
		}
	})
	return nil
}

func (cs CommandSet) AskInputs(values *map[string]any) error {
	for _, input := range cs.Inputs() {
		found := false
		for k := range *values {
			if input.Name == k {
				found = true
				break
			}
		}
		if found {
			continue
		}
		if val, err := askInput(input); err != nil {
			return err
		} else {
			(*values)[input.Name] = val
		}
	}
	return nil
}

func (cs CommandSet) ParseEnv(values *map[string]any, environ []string) {
	inputs := cs.Inputs()
	inputsMap := make(map[string]*ConfigInput, len(inputs))
	for _, input := range inputs {
		inputsMap[input.Name] = &input
	}
	for _, item := range environ {
		if !strings.HasPrefix(item, EnvVarPrefix) {
			continue
		}
		entry := strings.SplitN(item, "=", 2)
		name := strings.TrimPrefix(entry[0], EnvVarPrefix)
		if input, ok := inputsMap[name]; !ok {
			continue
		} else {
			logger.Printf("Found value for input in environment: %s\n", input.Name)
			(*values)[input.Name] = entry[1]
		}
	}
}

func (cs CommandSet) Cmd(data TemplateData, moreEnviron []string) (*exec.Cmd, error) {
	var scriptFile string
	var env []string
	var err error
	shell := cs.Shell()
	scriptFile, err = cs.RenderScriptToTemp(data)
	if err != nil {
		return nil, err
	}
	env, err = cs.RenderEnv(data)
	if err != nil {
		return nil, err
	}
	if !cs.Pure() {
		env = append(moreEnviron, env...)
	}
	shell = append(shell, scriptFile)
	cmd := exec.Command(shell[0], shell[1:]...)
	cmd.Env = env
	return cmd, nil
}

func (cs CommandSet) Runnable() bool {
	for i := len(cs.Commands) - 1; i >= 0; {
		command := cs.Commands[i]
		return command.Run != ""
	}
	return false
}

func (cs CommandSet) Selected() bool {
	for i := len(cs.Commands) - 1; i >= 0; {
		command := cs.Commands[i]
		return len(command.Commands) == 0
	}
	return false
}

func (cs CommandSet) AskCommands() (CommandSet, error) {
	for i := len(cs.Commands) - 1; i >= 0 && !cs.Selected(); i = len(cs.Commands) - 1 {
		command := cs.Commands[i]
		if subcommand, err := selectCommand(command); err != nil {
			return cs, err
		} else {
			cs.Commands = append(cs.Commands, subcommand)
		}
	}
	return cs, nil
}

func NewCommandSet(config Config, args []string) (CommandSet, error) {
	var cursor ConfigCommand
	cs := CommandSet{
		Config: config,
		Args:   args,
		Commands: []ConfigCommand{
			{
				Name:        "",
				Description: config.Description,
				Run:         config.Run,
				Shell:       config.Shell,
				Env:         config.Env,
				Pure:        config.Pure,
				Inputs:      config.Inputs,
				Commands:    config.Commands,
			},
		},
		ErrorHandling: flag.ContinueOnError,
	}

	for len(cs.Args) > 0 {
		cursor = cs.Commands[len(cs.Commands)-1]
		if cursor.Run != "" || len(cursor.Commands) == 0 {
			break
		}
		if cs.Args[0][0] == '-' {
			break
		}
		next := cursor.Commands.Get(cs.Args[0])
		if next == nil {
			return cs, fmt.Errorf("invalid subcommand: %s", args[0])
		}
		cs.Commands = append(cs.Commands, *next)
		cs.Args = cs.Args[1:]
	}
	if hasHelp(cs.Args) {
		logger.Println("Detected help flag whilst parsing arguments for command")
		return cs, flag.ErrHelp
	} else {
		return cs, nil
	}
}

func hasHelp(args []string) bool {
	for _, arg := range args {
		if arg == "-h" || arg == "--h" || arg == "-help" || arg == "--help" {
			return true
		}
	}
	return false
}
