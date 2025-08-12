package main

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"

	tea "github.com/charmbracelet/bubbletea"
)

type commandModel struct {
	commands []string
	cursor   int
	selected string
}

func (m commandModel) Init() tea.Cmd {
	return nil
}

func (m commandModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.commands)-1 {
				m.cursor++
			}
		case "enter":
			m.selected = m.commands[m.cursor]
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m commandModel) View() string {
	s := "Select a command:\n\n"
	for i, choice := range m.commands {
		cursor := " "
		if m.cursor == i {
			cursor = ">"
		}
		s += fmt.Sprintf("%s %s\n", cursor, choice)
	}
	return s
}

type inputModel struct {
	inputs       map[string]Input
	values       map[string]any
	current      string
	input        string
	prompt       string
	err          error
	options      []string
	optionCursor int
}

func (m inputModel) Init() tea.Cmd {
	return nil
}

func (m inputModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if len(m.options) > 0 {
			return m.updateOptions(msg)
		}
		return m.updateInput(msg)
	}
	return m, nil
}

func (m inputModel) updateOptions(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c", "q":
		return m, tea.Quit
	case "up", "k":
		if m.optionCursor > 0 {
			m.optionCursor--
		}
	case "down", "j":
		if m.optionCursor < len(m.options)-1 {
			m.optionCursor++
		}
	case "enter":
		m.values[m.current] = m.options[m.optionCursor]
		m.options = nil
		return m.nextInput()
	}
	return m, nil
}

func (m inputModel) updateInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyCtrlC, tea.KeyEsc:
		return m, tea.Quit
	case tea.KeyEnter:
		return m.validateAndSubmit()
	case tea.KeyBackspace:
		if len(m.input) > 0 {
			m.input = m.input[:len(m.input)-1]
		}
	case tea.KeyRunes:
		m.input += string(msg.Runes)
	}
	return m, nil
}

func (m inputModel) validateAndSubmit() (tea.Model, tea.Cmd) {
	input := m.inputs[m.current]
	val := m.input
	m.err = nil

	if input.Pattern != "" {
		if matched, _ := regexp.MatchString(input.Pattern, val); !matched {
			m.err = errors.New("input does not match pattern")
			return m, nil
		}
	}

	if input.Type == "number" {
		num, err := strconv.ParseFloat(val, 64)
		if err != nil {
			m.err = errors.New("invalid number")
			return m, nil
		}
		if input.Min != 0 && num < input.Min {
			m.err = fmt.Errorf("value must be greater than or equal to %f", input.Min)
			return m, nil
		}
		if input.Max != 0 && num > input.Max {
			m.err = fmt.Errorf("value must be less than or equal to %f", input.Max)
			return m, nil
		}
	}

	m.values[m.current] = val
	m.input = ""
	return m.nextInput()
}

func (m inputModel) nextInput() (tea.Model, tea.Cmd) {
	for name := range m.inputs {
		if _, ok := m.values[name]; !ok {
			m.current = name
			// Reset options
			m.options = nil
			m.optionCursor = 0
			if len(m.inputs[name].Options) > 0 {
				for _, opt := range m.inputs[name].Options {
					switch v := opt.(type) {
					case string:
						m.options = append(m.options, v)
					case map[string]any:
						for k := range v {
							m.options = append(m.options, k)
						}
					}
				}
			}
			return m, nil
		}
	}
	// All inputs collected
	return m, tea.Quit
}

func (m inputModel) View() string {
	if m.current == "" {
		return ""
	}

	if len(m.options) > 0 {
		s := m.inputs[m.current].Description + "\n\n"
		for i, choice := range m.options {
			cursor := " "
			if m.optionCursor == i {
				cursor = ">"
			}
			s += fmt.Sprintf("%s %s\n", cursor, choice)
		}
		return s
	}

	ss := fmt.Sprintf("%s\n> %s", m.inputs[m.current].Description, m.input)
	if m.err != nil {
		ss += fmt.Sprintf("\nError: %v", m.err)
	}
	return ss
}
