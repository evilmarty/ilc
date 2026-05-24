package inputs

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	titleStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("32")).Bold(true)
	progressStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
	accentStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("81")).Bold(true)
	dimStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("244"))
	errorStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Bold(true)
	helpStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Italic(true)
)

type tuiModel struct {
	title        string
	inputs       []*Input
	currentIndex int
	textInput    textinput.Model
	optionsIndex int
	err          error
	aborted      bool
}

func (m *tuiModel) initCurrentInput() {
	m.err = nil
	current := m.inputs[m.currentIndex]
	if !current.Selectable() && !m.isBooleanInput(current) {
		m.textInput = textinput.New()
		m.textInput.SetValue(current.Value.String())
		m.textInput.Placeholder = current.Value.String()
		m.textInput.Focus()
	} else {
		m.optionsIndex = 0
		if m.isBooleanInput(current) {
			opts := m.getBooleanOptions(current)
			targetVal := "false"
			if boolVal, ok := current.Value.(*BooleanValue); ok && boolVal.Value {
				targetVal = "true"
			}
			m.optionsIndex = 0
			for i, opt := range opts {
				if opt.Value == targetVal {
					m.optionsIndex = i
					break
				}
			}
		}
	}
}

func (m *tuiModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m *tuiModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			m.aborted = true
			return m, tea.Quit

		case tea.KeyEnter:
			current := m.inputs[m.currentIndex]
			var val string
			if current.Selectable() || m.isBooleanInput(current) {
				opts := m.getBooleanOptions(current)
				val = opts[m.optionsIndex].Value
			} else {
				val = m.textInput.Value()
				if val == "" {
					val = current.Value.String()
				}
			}

			if err := current.Value.Set(val); err != nil {
				m.err = err
				return m, nil
			}

			m.err = nil
			if m.currentIndex < len(m.inputs)-1 {
				m.currentIndex++
				m.initCurrentInput()
				return m, nil
			}
			return m, tea.Quit

		case tea.KeyUp:
			current := m.inputs[m.currentIndex]
			if current.Selectable() || m.isBooleanInput(current) {
				opts := m.getBooleanOptions(current)
				if m.optionsIndex > 0 {
					m.optionsIndex--
				} else {
					m.optionsIndex = len(opts) - 1
				}
			} else if adjustable, ok := current.Value.(AdjustableValue); ok {
				newStr, err := adjustable.Adjust(m.textInput.Value(), 1)
				m.textInput.SetValue(newStr)
				m.err = err
			}
			return m, nil

		case tea.KeyDown:
			current := m.inputs[m.currentIndex]
			if current.Selectable() || m.isBooleanInput(current) {
				opts := m.getBooleanOptions(current)
				if m.optionsIndex < len(opts)-1 {
					m.optionsIndex++
				} else {
					m.optionsIndex = 0
				}
			} else if adjustable, ok := current.Value.(AdjustableValue); ok {
				newStr, err := adjustable.Adjust(m.textInput.Value(), -1)
				m.textInput.SetValue(newStr)
				m.err = err
			}
			return m, nil
		}
	}

	current := m.inputs[m.currentIndex]
	if !current.Selectable() && !m.isBooleanInput(current) {
		m.textInput, cmd = m.textInput.Update(msg)
	}

	return m, cmd
}

func (m *tuiModel) View() string {
	if m.aborted {
		return ""
	}

	var sb strings.Builder



	current := m.inputs[m.currentIndex]

	// Progress and Label
	progress := fmt.Sprintf("[%d/%d]", m.currentIndex+1, len(m.inputs))
	prompt := current.Description
	if prompt == "" {
		if current.Selectable() || m.isBooleanInput(current) {
			prompt = fmt.Sprintf("Choose a %s", current.Name)
		} else {
			prompt = fmt.Sprintf("Please specify a %s", current.Name)
		}
	}

	sb.WriteString(fmt.Sprintf("%s %s\n", progressStyle.Render(progress), prompt))

	// Render specific control
	if current.Selectable() || m.isBooleanInput(current) {
		sb.WriteString("\n")
		opts := m.getBooleanOptions(current)
		for i, option := range opts {
			if i == m.optionsIndex {
				sb.WriteString(fmt.Sprintf("  ❯ %s\n", accentStyle.Render(option.Label)))
			} else {
				sb.WriteString(fmt.Sprintf("    %s\n", dimStyle.Render(option.Label)))
			}
		}
	} else {
		sb.WriteString("\n  " + m.textInput.View() + "\n")
	}

	// Validation Error Display
	if m.err != nil {
		sb.WriteString("\n" + errorStyle.Render(fmt.Sprintf("  ✗ Invalid input: %v", m.err)) + "\n")
	}

	// Help Guidelines
	var helpParts []string
	if _, isAdjustable := current.Value.(AdjustableValue); isAdjustable {
		helpParts = append(helpParts, "[Up/Down] +/-")
	}
	helpParts = append(helpParts, "[Enter] Confirm", "[Ctrl+C / Esc] Abort")
	sb.WriteString("\n" + helpStyle.Render("  "+strings.Join(helpParts, "  •  ")) + "\n")

	return sb.String()
}

func (m *tuiModel) isBooleanInput(current *Input) bool {
	_, isBool := current.Value.(*BooleanValue)
	return isBool
}

func (m *tuiModel) getBooleanOptions(current *Input) []InputOption {
	if current.Selectable() {
		return current.Options
	}
	return []InputOption{
		{Label: "true", Value: "true"},
		{Label: "false", Value: "false"},
	}
}

type Prompter interface {
	Prompt(title string, missing []*Input) error
}

type TuiPrompter struct{}

func (tp TuiPrompter) Prompt(title string, missing []*Input) error {
	m := tuiModel{
		title:        title,
		inputs:       missing,
		currentIndex: 0,
	}
	m.initCurrentInput()

	p := tea.NewProgram(&m)
	if _, err := p.Run(); err != nil {
		return err
	}

	if m.aborted {
		return ErrAborted
	}

	return nil
}


