package main

import (
	"bytes"
	"os/exec"
	"strings"

	spinner "github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	bgCmdSpinnerStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
)

type BgCmd struct {
	Prefix  string
	Suffix  string
	spinner spinner.Model
	cmd     *exec.Cmd
	err     error
}

type bgCmdFinishedMsg struct {
	Err error
}

func (c BgCmd) Init() tea.Cmd {
	return tea.Batch(c.spinner.Tick, c.exec)
}

func (c BgCmd) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	cmd := tea.Quit
	switch msg := msg.(type) {
	case bgCmdFinishedMsg:
		c.err = msg.Err
	default:
		c.spinner, cmd = c.spinner.Update(msg)
	}
	return c, cmd
}

func (c BgCmd) View() string {
	return strings.Join([]string{c.Prefix, c.spinner.View(), c.Suffix}, " ")
}

func (c *BgCmd) Run() error {
	if err := tea.NewProgram(c).Start(); err != nil {
		return err
	}
	return c.err
}

func (c BgCmd) Output() ([]byte, error) {
	var stdout bytes.Buffer
	c.cmd.Stdout = &stdout
	err := c.Run()
	return stdout.Bytes(), err
}

func (c BgCmd) exec() tea.Msg {
	err := c.cmd.Run()
	return bgCmdFinishedMsg{Err: err}
}

func BgCommand(cmd *exec.Cmd) BgCmd {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = bgCmdSpinnerStyle
	return BgCmd{
		cmd:     cmd,
		spinner: s,
	}
}
