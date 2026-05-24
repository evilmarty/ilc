package ilc

import (
	"errors"
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/evilmarty/ilc/internal/inputs"
)

var (
	titleStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("32")).Bold(true)
	dimStyle        = lipgloss.NewStyle().Foreground(lipgloss.Color("244"))
	accentStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("81")).Bold(true)
	descDimStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
	descActiveStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("244"))
	helpStyle       = lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Italic(true)
	progressStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
	errorStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Bold(true)
	cmdPathStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("255"))
	neutralPromptStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	validPromptStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("150"))
	invalidPromptStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("203"))
)

type commandMode int

const (
	modeCommandSelect commandMode = iota
	modeInputPrompt
)

type commandModel struct {
	title         string
	history       []Selection
	selectedIndex int
	aborted       bool
	done          bool

	// Inputs wizard integrated fields
	mode         commandMode
	missing      []*inputs.Input
	inputIndex   int
	textInput    textinput.Model
	optionsIndex int
	inputErr     error
	env          map[string]string
}

func (m *commandModel) currentSelection() Selection {
	return m.history[len(m.history)-1]
}

func (m *commandModel) currentSubcommands() SubCommands {
	return m.currentSelection().Commands()
}

func (m *commandModel) initCurrentInput() {
	m.inputErr = nil
	if len(m.missing) == 0 {
		return
	}
	current := m.missing[m.inputIndex]
	if !current.Selectable() {
		m.textInput = textinput.New()
		m.textInput.Placeholder = current.Value.String()
		m.textInput.PlaceholderStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("244"))
		m.textInput.Focus()
	} else {
		m.optionsIndex = 0
	}
}

func (m *commandModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m *commandModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	if m.mode == modeInputPrompt {
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.Type {
			case tea.KeyCtrlC:
				m.aborted = true
				return m, tea.Quit

			case tea.KeyEsc:
				if m.inputIndex > 0 {
					m.inputIndex--
					m.initCurrentInput()
					return m, nil
				}
				// Go back to command selection mode
				m.mode = modeCommandSelect
				m.selectedIndex = 0
				if len(m.history) > 1 {
					m.history = m.history[:len(m.history)-1]
				}
				return m, nil

			case tea.KeyEnter:
				current := m.missing[m.inputIndex]
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
					m.inputErr = err
					return m, nil
				}

				m.inputErr = nil
				if m.inputIndex < len(m.missing)-1 {
					m.inputIndex++
					m.initCurrentInput()
					return m, nil
				}
				m.done = true
				return m, tea.Quit

			case tea.KeyUp:
				current := m.missing[m.inputIndex]
				if current.Selectable() {
					if m.optionsIndex > 0 {
						m.optionsIndex--
					} else {
						m.optionsIndex = len(current.Options) - 1
					}
				} else if adjustable, ok := current.Value.(inputs.AdjustableValue); ok {
					newStr, err := adjustable.Adjust(m.textInput.Value(), 1)
					m.textInput.SetValue(newStr)
					m.inputErr = err
				}
				return m, nil

			case tea.KeyDown:
				current := m.missing[m.inputIndex]
				if current.Selectable() {
					if m.optionsIndex < len(current.Options)-1 {
						m.optionsIndex++
					} else {
						m.optionsIndex = 0
					}
				} else if adjustable, ok := current.Value.(inputs.AdjustableValue); ok {
					newStr, err := adjustable.Adjust(m.textInput.Value(), -1)
					m.textInput.SetValue(newStr)
					m.inputErr = err
				}
				return m, nil
			}
		}

		current := m.missing[m.inputIndex]
		if !current.Selectable() {
			m.textInput, cmd = m.textInput.Update(msg)

			// Live validation dry-run
			val := m.textInput.Value()
			if val == "" {
				val = current.Value.String()
			}

			var err error
			if validator, ok := current.Value.(inputs.LiveValidator); ok {
				err = validator.ValidateLive(val)
			} else {
				err = current.Value.Set(val)
			}

			m.inputErr = err
		}
		return m, cmd
	}

	// modeCommandSelect
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC:
			m.aborted = true
			return m, tea.Quit

		case tea.KeyEsc:
			if len(m.history) > 1 {
				m.history = m.history[:len(m.history)-1]
				m.selectedIndex = 0
				return m, nil
			}
			return m, nil

		case tea.KeyEnter:
			subs := m.currentSubcommands()
			if len(subs) == 0 {
				return m, nil
			}
			choice := subs[m.selectedIndex]
			nextSel := m.currentSelection().SelectCommand(choice.Command, m.currentSelection().Args)

			if nextSel.Runnable() {
				inps := nextSel.Inputs()
				missing, err := inps.ParseEnvAndArgs(nextSel.Args, m.env)
				if err != nil {
					m.inputErr = err
					return m, nil
				}

				if len(missing) == 0 {
					m.history = append(m.history, nextSel)
					m.done = true
					return m, tea.Quit
				}

				m.history = append(m.history, nextSel)
				m.missing = missing
				m.inputIndex = 0
				m.mode = modeInputPrompt
				m.initCurrentInput()
				return m, nil
			}

			m.history = append(m.history, nextSel)
			m.selectedIndex = 0
			return m, nil

		case tea.KeyUp:
			subs := m.currentSubcommands()
			if len(subs) > 0 {
				if m.selectedIndex > 0 {
					m.selectedIndex--
				} else {
					m.selectedIndex = len(subs) - 1
				}
			}
			return m, nil

		case tea.KeyDown:
			subs := m.currentSubcommands()
			if len(subs) > 0 {
				if m.selectedIndex < len(subs)-1 {
					m.selectedIndex++
				} else {
					m.selectedIndex = 0
				}
			}
			return m, nil
		}
	}
	return m, nil
}

func (m *commandModel) View() string {
	if m.aborted {
		return ""
	}

	var bcStyled strings.Builder
	var bcPlain strings.Builder
	for i, sel := range m.history {
		if i == 0 {
			continue
		}
		cmd := sel.commands[len(sel.commands)-1]
		if i > 1 {
			bcStyled.WriteString(titleStyle.Render(" ❯ "))
			bcPlain.WriteString(" ❯ ")
		} else {
			bcStyled.WriteString(" ")
			bcPlain.WriteString(" ")
		}
		bcStyled.WriteString(cmdPathStyle.Render(cmd.Name))
		bcPlain.WriteString(cmd.Name)
	}

	if m.done {
		var exitSb strings.Builder
		exitSb.WriteString(fmt.Sprintf("%s%s\n", titleStyle.Render("Command:"), bcStyled.String()))
		for _, completed := range m.missing {
			exitSb.WriteString(titleStyle.Render(completed.Name+":") + " " + cmdPathStyle.Render(completed.Value.String()) + "\n")
		}
		return exitSb.String()
	}

	var sb strings.Builder

	sb.WriteString(titleStyle.Render(fmt.Sprintf("── %s ──", m.title)) + "\n\n")

	sb.WriteString(titleStyle.Render("Command:"))
	if bcPlain.Len() > 0 {
		sb.WriteString(bcStyled.String())
		if m.mode == modeCommandSelect {
			sb.WriteString(titleStyle.Render(" ❯"))
		}
	}
	if m.mode == modeInputPrompt {
		sb.WriteString("\n")
	} else {
		sb.WriteString("\n\n")
	}

	if m.mode == modeCommandSelect {
		subs := m.currentSubcommands()
		maxLen := 0
		for _, sub := range subs {
			maxLen = max(maxLen, utf8.RuneCountInString(sub.Name))
		}

		for i, sub := range subs {
			padLen := maxLen - utf8.RuneCountInString(sub.Name)
			if padLen < 0 {
				padLen = 0
			}
			padding := strings.Repeat(" ", padLen+5)

			var nameStr string
			var descStr string

			if i == m.selectedIndex {
				nameStr = accentStyle.Render(sub.Name)
				if sub.Description != "" {
					descStr = descActiveStyle.Render(padding + sub.Description)
				}
				sb.WriteString(fmt.Sprintf("  ❯ %s%s\n", nameStr, descStr))
			} else {
				nameStr = dimStyle.Render(sub.Name)
				if sub.Description != "" {
					descStr = descDimStyle.Render(padding + sub.Description)
				}
				sb.WriteString(fmt.Sprintf("    %s%s\n", nameStr, descStr))
			}
		}

		sb.WriteString("\n" + helpStyle.Render("  [Enter] Select/Confirm  •  [Esc] Back  •  [Ctrl+C] Abort") + "\n")
	} else {
		// modeInputPrompt
		// Render completed inputs in progressive/condensed form
		for i := 0; i < m.inputIndex; i++ {
			completed := m.missing[i]
			sb.WriteString(titleStyle.Render(completed.Name+":") + " " + cmdPathStyle.Render(completed.Value.String()) + "\n")
		}

		current := m.missing[m.inputIndex]

		// Active input prompt name
		sb.WriteString(titleStyle.Render(current.Name + ":"))

		// Active input description next to the name if available
		if current.Description != "" {
			sb.WriteString(descActiveStyle.Render("   " + current.Description))
		}
		sb.WriteString("\n")

		// Render active input control
		if current.Selectable() {
			sb.WriteString("\n")
			for i, option := range current.Options {
				if i == m.optionsIndex {
					sb.WriteString(fmt.Sprintf("    ❯ %s\n", accentStyle.Render(option.Label)))
				} else {
					sb.WriteString(fmt.Sprintf("      %s\n", dimStyle.Render(option.Label)))
				}
			}
		} else {
			// Dynamically style the text input prompt based on validation state
			if m.textInput.Value() == "" {
				m.textInput.Prompt = neutralPromptStyle.Render("> ")
			} else if m.inputErr != nil {
				m.textInput.Prompt = invalidPromptStyle.Render("✘ ")
			} else {
				m.textInput.Prompt = validPromptStyle.Render("✔ ")
			}
			sb.WriteString("\n    " + m.textInput.View() + "\n")
		}

		// Help Guidelines (Hide confirm instruction if active input is invalid)
		var helpParts []string
		if _, isAdjustable := current.Value.(inputs.AdjustableValue); isAdjustable {
			helpParts = append(helpParts, "[Up/Down] +/-")
		}
		if m.inputErr == nil {
			helpParts = append(helpParts, "[Enter] Confirm")
		}
		helpParts = append(helpParts, "[Esc] Back", "[Ctrl+C] Abort")
		sb.WriteString("\n" + helpStyle.Render("  "+strings.Join(helpParts, "  •  ")) + "\n")
	}

	return sb.String()
}

func askCommands(sel Selection, env map[string]string) (Selection, error) {
	title := sel.commands[0].Description
	if title == "" {
		title = sel.commands[0].Name
	}
	if title == "" {
		title = "Choose command"
	}

	var history []Selection
	for i := 1; i <= len(sel.commands); i++ {
		history = append(history, Selection{
			commands: sel.commands[:i],
			Args:     sel.Args,
		})
	}

	m := commandModel{
		title:         title,
		history:       history,
		selectedIndex: 0,
		mode:          modeCommandSelect,
		env:           env,
	}

	if sel.Runnable() {
		inps := sel.Inputs()
		missing, err := inps.ParseEnvAndArgs(sel.Args, env)
		if err == nil && len(missing) > 0 {
			m.missing = missing
			m.inputIndex = 0
			m.mode = modeInputPrompt
			m.initCurrentInput()
		}
	}

	p := tea.NewProgram(&m)
	if _, err := p.Run(); err != nil {
		return sel, err
	}

	if m.aborted {
		return sel, inputs.ErrAborted
	}

	if len(m.history) > 0 {
		finalSel := m.history[len(m.history)-1]
		if finalSel.Runnable() {
			logger.Printf("selected command: %s", finalSel.String())
			return finalSel, nil
		}
	}

	return sel, errors.New("no choice made")
}


