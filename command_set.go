package main

import (
	"flag"
	"fmt"
	"os/exec"
)

type CommandSet struct {
	Config   Config
	Commands CommandChain
	Args     []string
}

func (cs CommandSet) parseArgs(values *map[string]any) error {
	fs := flag.NewFlagSet(cs.Commands.Name(), flag.ExitOnError)
	for _, input := range cs.Commands.Inputs() {
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

func (cs CommandSet) askInputs(values *map[string]any) error {
	for _, input := range cs.Commands.Inputs() {
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

func (cs CommandSet) Values() (map[string]any, error) {
	var err error
	values := make(map[string]any)
	err = cs.parseArgs(&values)
	if err == nil {
		err = cs.askInputs(&values)
	}
	return values, err
}

func (cs CommandSet) Cmd(data map[string]any, moreEnv []string) (*exec.Cmd, error) {
	var scriptFile string
	var env []string
	var err error
	shell := cs.Commands.Shell()
	if len(shell) == 0 {
		shell = defaultShell[:]
	}
	scriptFile, err = cs.Commands.RenderScriptToTemp(data)
	if err != nil {
		return nil, err
	}
	env, err = cs.Commands.RenderEnv(data)
	if err != nil {
		return nil, err
	}
	if !cs.Commands.Pure() {
		env = append(moreEnv, env...)
	}
	shell = append(shell, scriptFile)
	cmd := exec.Command(shell[0], shell[1:]...)
	cmd.Env = env
	return cmd, nil
}

func NewCommandSet(config Config, args []string) (CommandSet, error) {
	var cursor ConfigCommand
	rootCommand := ConfigCommand{
		Name:        "",
		Description: config.Description,
		Run:         config.Run,
		Shell:       config.Shell,
		Env:         config.Env,
		Pure:        config.Pure,
		Inputs:      config.Inputs,
		Commands:    config.Commands,
	}
	cc := []ConfigCommand{rootCommand}

	for len(args) > 0 {
		cursor = cc[len(cc)-1]
		if cursor.Run != "" || len(cursor.Commands) == 0 {
			break
		}
		if args[0][0] == '-' {
			break
		}
		next := cursor.Commands.Get(args[0])
		if next == nil {
			return CommandSet{}, fmt.Errorf("invalid subcommand: %s", args[0])
		}
		cc = append(cc, *next)
		args = args[1:]
	}
	// Now we ask to select any remaining commands
	for cursor = cc[len(cc)-1]; cursor.Run == ""; cursor = cc[len(cc)-1] {
		if subcommand, err := selectCommand(cursor); err != nil {
			break
		} else {
			cc = append(cc, subcommand)
		}
	}
	cs := CommandSet{
		Config:   config,
		Commands: cc,
		Args:     args,
	}
	return cs, nil
}
