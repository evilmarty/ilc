package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"text/template"
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

func (cs CommandSet) Env() EnvMap {
	em := NewEnvMap([]string{})
	for _, command := range cs.Commands {
		em = em.Merge(command.Env)
	}
	return em
}

func (cs CommandSet) RenderEnv(data TemplateData) (EnvMap, error) {
	env := cs.Env()
	for name, template := range env {
		if value, err := RenderTemplate(template, data); err != nil {
			return env, fmt.Errorf("template error for environment variable: '%s' - %v", name, err)
		} else {
			env[name] = value
		}
	}
	return env, nil
}

func (cs CommandSet) RenderScript(data TemplateData) (string, error) {
	var tmpl *template.Template
	for _, command := range cs.Commands {
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
		// Set all default values
		if input.DefaultValue != nil {
			(*values)[input.Name] = input.DefaultValue
		}
		switch input.Type {
		case "number":
			fs.Float64(input.Name, 0.0, input.Description)
		case "boolean":
			fs.Bool(input.Name, false, input.Description)
		default:
			fs.String(input.Name, "", input.Description)
		}
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

func (cs CommandSet) Validate(values map[string]any) error {
	for _, input := range cs.Inputs() {
		found := false
		for k, value := range values {
			if input.Name == k {
				found = true
				if !input.Valid(value) {
					return fmt.Errorf("invalid input: %s", input.Name)
				}
				break
			}
		}
		if !found {
			return fmt.Errorf("missing input: %s", input.Name)
		}
	}
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

func (cs CommandSet) ParseEnv(values *map[string]any, env EnvMap) {
	inputs := cs.Inputs()
	inputsMap := make(map[string]*ConfigInput, len(inputs))
	for _, input := range inputs {
		inputsMap[input.Name] = &input
	}
	for key, val := range env {
		if !strings.HasPrefix(key, EnvVarPrefix) {
			continue
		}
		name := strings.TrimPrefix(key, EnvVarPrefix)
		if input, ok := inputsMap[name]; !ok {
			continue
		} else if val, ok := input.Parse(val); !ok {
			logger.Printf("Invalid value for input in environment: %s\n", input.Name)
			continue
		} else {
			logger.Printf("Found value for input in environment: %s\n", input.Name)
			(*values)[input.Name] = val
		}
	}
}

func (cs CommandSet) getInputsEnv(data TemplateData) EnvMap {
	inputs := cs.Inputs()
	env := make(EnvMap, len(inputs))
	for _, input := range inputs {
		value := data.getInput(input.Name)
		key := fmt.Sprintf("%s%s", EnvVarPrefix, input.SafeName())
		env[key] = fmt.Sprintf("%s", value)
	}
	return env
}

func (cs CommandSet) Cmd(data TemplateData, moreEnv EnvMap) (*exec.Cmd, error) {
	var scriptFile string
	var env EnvMap
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
	env = cs.getInputsEnv(data).Merge(env)
	if !cs.Pure() {
		env = moreEnv.Merge(env)
	}
	shell = append(shell, scriptFile)
	cmd := exec.Command(shell[0], shell[1:]...)
	cmd.Env = env.ToList()
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
