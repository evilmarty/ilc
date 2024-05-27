package main

import (
	"strings"
)

type ArgSet struct {
	Commands []string
	Params   map[string]string
}

func (argset *ArgSet) ParamNames() []string {
	var names []string
	for name := range argset.Params {
		names = append(names, name)
	}
	return names
}

func ParseArgSet(args []string) ArgSet {
	argsLength := len(args)
	commands := make([]string, 0, argsLength)
	params := make(map[string]string, argsLength)
	paramKey := ""
	for _, arg := range args {
		if strings.HasPrefix(arg, "-") {
			paramKey = strings.TrimLeft(arg, "-")
			if strings.Contains(paramKey, "=") {
				parts := strings.SplitN(paramKey, "=", 2)
				params[parts[0]] = parts[1]
				paramKey = ""
			}
			continue
		}
		if paramKey != "" {
			params[paramKey] = arg
			paramKey = ""
			continue
		} else {
			commands = append(commands, arg)
		}
	}
	return ArgSet{Commands: commands, Params: params}
}
