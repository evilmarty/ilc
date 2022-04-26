package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"text/template"

	"github.com/charmbracelet/lipgloss"
	"github.com/erikgeiser/promptkit/selection"
	"github.com/erikgeiser/promptkit/textinput"
)

const (
	defaultConfigFile  = "ilc.yml"
	minChoiceFiltering = 5
)

var (
	promptStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("#666"))
	inputNameStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#0cc")).Bold(true)
)

func renderTemplate(name, text string, data map[string]any) (string, error) {
	b := strings.Builder{}
	if tmpl, err := template.New(name).Parse(text); err != nil {
		return "", err
	} else if tmpl.Execute(&b, data); err != nil {
		return "", err
	} else {
		return b.String(), nil
	}
}

func askInputChoice(input *ConfigCommandInput) (any, error) {
	var choices = make([]*selection.Choice, 0, len(input.Options))
	prompt := fmt.Sprintf("%s %s", promptStyle.Render("Choose a"), inputNameStyle.Render(input.Name))
	for label, value := range input.Options {
		choices = append(choices, &selection.Choice{String: label, Value: value})
	}
	sp := selection.New(prompt, choices)

	if len(choices) <= minChoiceFiltering {
		sp.Filter = nil
	}

	if choice, err := sp.RunPrompt(); err != nil {
		return choice, err
	} else {
		return choice.Value, err
	}
}

func askInputText(input *ConfigCommandInput) (any, error) {
	prompt := fmt.Sprintf("%s %s", promptStyle.Render("Please specify a"), inputNameStyle.Render(input.Name))
	ti := textinput.New(prompt)
	ti.InitialValue = input.Default
	ti.Validate = input.Validate
	return ti.RunPrompt()
}

func askInput(input *ConfigCommandInput) (any, error) {
	if len(input.Options) > 0 {
		return askInputChoice(input)
	} else {
		return askInputText(input)
	}
}

type model struct {
	config  *Config
	command *ConfigCommand
	values  map[string]any
	flagSet *flag.FlagSet
}

func (m *model) setCommand(command *ConfigCommand) {
	m.command = command
	m.values = make(map[string]any, len(command.Inputs))
}

func (m *model) populate(args []string) error {
	if len(args) == 0 {
		return nil
	}

	for _, command := range m.config.Commands {
		if command.Name == args[0] {
			m.setCommand(&command)
		}
	}

	if m.command == nil {
		return fmt.Errorf("Unknown command: %s", args[0])
	}

	values := make(map[string]*string)
	m.flagSet = flag.NewFlagSet(m.config.Description, flag.ExitOnError)
	for _, input := range m.command.Inputs {
		values[input.Name] = m.flagSet.String(input.Name, input.Default, "")
	}

	if err := m.flagSet.Parse(args[1:]); err != nil {
		return err
	}

	for name, value := range values {
		if value != nil && (*value) != "" {
			m.values[name] = value
		}
	}

	return nil
}

func (m *model) askCommand() error {
	if m.command != nil {
		return nil
	}

	var choices = make([]*selection.Choice, len(m.config.Commands))

	for i, command := range m.config.Commands {
		choices[i] = &selection.Choice{String: command.Name, Value: command}
	}

	prompt := promptStyle.Render("Choose command")
	sp := selection.New(prompt, choices)

	if len(choices) <= minChoiceFiltering {
		sp.Filter = nil
	}

	choice, err := sp.RunPrompt()
	if err != nil {
		return err
	}
	if command, ok := choice.Value.(ConfigCommand); ok {
		m.setCommand(&command)
		return nil
	} else {
		return fmt.Errorf("Failed to cast choice: %s", choice.String)
	}
}

func (m *model) askInputs() error {
	if m.command == nil {
		return nil
	}

	for _, input := range m.command.Inputs {
		if _, ok := m.values[input.Name]; ok {
			continue
		}

		value, err := askInput(&input)
		if err != nil {
			return err
		} else {
			m.values[input.Name] = value
		}
	}

	return nil
}

func (m *model) ask() error {
	if err := m.askCommand(); err != nil {
		return err
	} else if err := m.askInputs(); err != nil {
		return nil
	} else {
		return nil
	}
}

func (m *model) env() []string {
	if m.command == nil {
		return []string{}
	}

	var env = make([]string, len(m.command.Env))

	for name, rawValue := range m.command.Env {
		if value, err := renderTemplate(name, rawValue, m.values); err == nil {
			env = append(env, fmt.Sprintf("%s=%s", name, value))
		}
	}

	return env
}

func (m *model) exec() error {
	if m.command == nil {
		return fmt.Errorf("No command specified")
	}

	if script, err := renderTemplate(m.command.Name, m.command.Run, m.values); err != nil {
		return err
	} else {
		cmd := exec.Command("sh", "-c", script)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Env = m.env()
		return cmd.Run()
	}
}

func newModel(defaultConfigFile string) *model {
	var configFile = flag.String("f", defaultConfigFile, "Config file to load")
	flag.Parse()

	config, err := LoadConfig(*configFile)
	if err != nil {
		panic(err)
	}

	m := model{config: config}
	if err := m.populate(flag.Args()); err != nil {
		panic(err)
	}

	return &m
}

func main() {
	m := newModel(defaultConfigFile)

	if err := m.ask(); err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}

	if err := m.exec(); err != nil {
		fmt.Println("Error with template:", err)
		os.Exit(1)
	}
}
