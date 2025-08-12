package main

import (
	"flag"
	"os"

	tea "github.com/charmbracelet/bubbletea"
)

func collectInputs(cmd *Command, cmdArgs []string) (map[string]any, error) {
	values := make(map[string]any)
	args := flag.Args()[len(cmdArgs):]
	parseInputArgs(args, values)
	getInitialInputValues(cmd.Inputs, values)

	missingInputs := getMissingInputs(cmd.Inputs, values)
	if len(missingInputs) == 0 {
		return values, nil
	}

	return runInputCollector(cmd, values, missingInputs)
}

func runInputCollector(cmd *Command, values map[string]any, missingInputs []string) (map[string]any, error) {
	initialModel := inputModel{
		inputs:  cmd.Inputs,
		values:  values,
		current: missingInputs[0],
	}
	initialModel.options = getOptions(cmd.Inputs[initialModel.current])

	p := tea.NewProgram(initialModel)
	m, err := p.Run()
	if err != nil {
		return nil, err
	}

	return m.(inputModel).values, nil
}

func parseInputArgs(args []string, values map[string]any) {
	for i := 0; i < len(args); i++ {
		if (args[i][0] == '-') && i+1 < len(args) {
			name := args[i][1:]
			if args[i][1] == '-' {
				name = args[i][2:]
			}
			values[name] = args[i+1]
			i++
		}
	}
}

func getInitialInputValues(inputs map[string]Input, values map[string]any) {
	for name, input := range inputs {
		if _, ok := values[name]; !ok {
			envVar := "ILC_INPUT_" + name
			if val := os.Getenv(envVar); val != "" {
				values[name] = val
			} else if input.Default != nil {
				values[name] = input.Default
			}
		}
	}
}

func getMissingInputs(inputs map[string]Input, values map[string]any) []string {
	var missing []string
	for name := range inputs {
		if _, ok := values[name]; !ok {
			missing = append(missing, name)
		}
	}
	return missing
}

func getOptions(input Input) []string {
	var options []string
	if len(input.Options) > 0 {
		for _, opt := range input.Options {
			switch v := opt.(type) {
			case string:
				options = append(options, v)
			case map[string]any:
				for k := range v {
					options = append(options, k)
				}
			}
		}
	}
	return options
}
