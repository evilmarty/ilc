package inputs

import (
	"fmt"
	"math"
	"strconv"
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
	if !current.Selectable() {
		m.textInput = textinput.New()
		m.textInput.Placeholder = current.Value.String()
		m.textInput.Focus()
	} else {
		m.optionsIndex = 0
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
			if current.Selectable() {
				val = current.Options[m.optionsIndex].Value
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
			if current.Selectable() {
				if m.optionsIndex > 0 {
					m.optionsIndex--
				} else {
					m.optionsIndex = len(current.Options) - 1
				}
			} else if _, isNumber := current.Value.(*NumberValue); isNumber {
				m.adjustNumberInput(1)
			}
			return m, nil

		case tea.KeyDown:
			current := m.inputs[m.currentIndex]
			if current.Selectable() {
				if m.optionsIndex < len(current.Options)-1 {
					m.optionsIndex++
				} else {
					m.optionsIndex = 0
				}
			} else if _, isNumber := current.Value.(*NumberValue); isNumber {
				m.adjustNumberInput(-1)
			}
			return m, nil
		}
	}

	current := m.inputs[m.currentIndex]
	if !current.Selectable() {
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
		if current.Selectable() {
			prompt = fmt.Sprintf("Choose a %s", current.Name)
		} else {
			prompt = fmt.Sprintf("Please specify a %s", current.Name)
		}
	}

	sb.WriteString(fmt.Sprintf("%s %s\n", progressStyle.Render(progress), prompt))

	// Render specific control
	if current.Selectable() {
		sb.WriteString("\n")
		for i, option := range current.Options {
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
	if _, isNumber := current.Value.(*NumberValue); isNumber {
		helpParts = append(helpParts, "[Up/Down] +/-")
	}
	helpParts = append(helpParts, "[Enter] Confirm", "[Ctrl+C / Esc] Abort")
	sb.WriteString("\n" + helpStyle.Render("  "+strings.Join(helpParts, "  •  ")) + "\n")

	return sb.String()
}

func (fs *FlagSet) promptTUI(missing []*Input) error {
	m := tuiModel{
		title:        fs.name,
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

func (m *tuiModel) adjustNumberInput(delta float64) {
	current := m.inputs[m.currentIndex]
	numVal, ok := current.Value.(*NumberValue)
	if !ok {
		return
	}

	val := m.textInput.Value()
	if val == "" {
		val = current.Value.String()
	}

	n, err := strconv.ParseFloat(val, 64)
	if err != nil {
		n = numVal.Value
	}

	newVal := n + delta

	// Clamping to min/max bounds if defined
	if numVal.MinValue < numVal.MaxValue {
		if newVal < numVal.MinValue {
			newVal = numVal.MinValue
		} else if newVal > numVal.MaxValue {
			newVal = numVal.MaxValue
		}
	} else if numVal.MinValue > numVal.MaxValue && numVal.MaxValue == 0.0 {
		if newVal < numVal.MinValue {
			newVal = numVal.MinValue
		}
	}

	// Format back cleanly (avoiding .00000 decimals for integers)
	prec := 5
	if newVal == math.Round(newVal) {
		prec = 0
	}
	newStr := strconv.FormatFloat(newVal, 'f', prec, 64)

	m.textInput.SetValue(newStr)

	// Update validation error
	m.err = numVal.Set(newStr)
}
