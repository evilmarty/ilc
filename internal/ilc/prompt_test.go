package ilc

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/evilmarty/ilc/internal/inputs"
	"github.com/stretchr/testify/assert"
)

func TestCommandModel_NumberInputAdjustment(t *testing.T) {
	numVal := &inputs.NumberValue{Value: 3, MinValue: 1, MaxValue: 5}
	input := &inputs.Input{
		Name:  "rating",
		Value: numVal,
	}
	m := &commandModel{
		mode:       modeInputPrompt,
		missing:    []*inputs.Input{input},
		inputIndex: 0,
	}
	m.initCurrentInput()

	// Initially, textInput value is pre-populated with default
	if m.textInput.Value() != "3" {
		t.Errorf("expected textInput to be initially '3', got %q", m.textInput.Value())
	}

	// Send KeyUp message (should increment to 4, since 3 + 1 = 4)
	_, _ = m.Update(tea.KeyMsg{Type: tea.KeyUp})
	if m.textInput.Value() != "4" {
		t.Errorf("expected textInput value after KeyUp to be 4, got %q", m.textInput.Value())
	}

	// Send KeyUp message (should increment to 5)
	_, _ = m.Update(tea.KeyMsg{Type: tea.KeyUp})
	if m.textInput.Value() != "5" {
		t.Errorf("expected textInput value after second KeyUp to be 5, got %q", m.textInput.Value())
	}

	// Send KeyUp message (should clamp to 5)
	_, _ = m.Update(tea.KeyMsg{Type: tea.KeyUp})
	if m.textInput.Value() != "5" {
		t.Errorf("expected textInput value to clamp to MaxValue 5, got %q", m.textInput.Value())
	}

	// Send KeyDown message (should decrement to 4)
	_, _ = m.Update(tea.KeyMsg{Type: tea.KeyDown})
	if m.textInput.Value() != "4" {
		t.Errorf("expected textInput value after KeyDown to be 4, got %q", m.textInput.Value())
	}

	// Send KeyDown multiple times to test lower clamp
	_, _ = m.Update(tea.KeyMsg{Type: tea.KeyDown}) // 3
	_, _ = m.Update(tea.KeyMsg{Type: tea.KeyDown}) // 2
	_, _ = m.Update(tea.KeyMsg{Type: tea.KeyDown}) // 1
	_, _ = m.Update(tea.KeyMsg{Type: tea.KeyDown}) // 1 (clamped)
	if m.textInput.Value() != "1" {
		t.Errorf("expected textInput value to clamp to MinValue 1, got %q", m.textInput.Value())
	}
}

func TestCommandModel_NumberInputAdjustment_NoMax(t *testing.T) {
	// A number input with MinValue = 1, MaxValue = 0 means minimum is 1, no upper bound
	numVal := &inputs.NumberValue{Value: 1, MinValue: 1, MaxValue: 0}
	input := &inputs.Input{
		Name:  "quantity",
		Value: numVal,
	}
	m := &commandModel{
		mode:       modeInputPrompt,
		missing:    []*inputs.Input{input},
		inputIndex: 0,
	}
	m.initCurrentInput()

	// Send KeyUp (should increment to 2)
	_, _ = m.Update(tea.KeyMsg{Type: tea.KeyUp})
	if m.textInput.Value() != "2" {
		t.Errorf("expected textInput value to be 2, got %q", m.textInput.Value())
	}

	// Send KeyDown (should decrement to 1)
	_, _ = m.Update(tea.KeyMsg{Type: tea.KeyDown})
	if m.textInput.Value() != "1" {
		t.Errorf("expected textInput value to be 1, got %q", m.textInput.Value())
	}

	// Send KeyDown (should clamp to 1 as MinValue is 1)
	_, _ = m.Update(tea.KeyMsg{Type: tea.KeyDown})
	if m.textInput.Value() != "1" {
		t.Errorf("expected textInput value to clamp to 1, got %q", m.textInput.Value())
	}
}

func TestCommandModel_NumberInputAdjustment_IncompleteNumber(t *testing.T) {
	numVal := &inputs.NumberValue{Value: 5, MinValue: 1, MaxValue: 10}
	input := &inputs.Input{
		Name:  "rating",
		Value: numVal,
	}
	m := &commandModel{
		mode:       modeInputPrompt,
		missing:    []*inputs.Input{input},
		inputIndex: 0,
	}
	m.initCurrentInput()

	// Manually set an incomplete/invalid string in textInput
	m.textInput.SetValue("abc")

	// Trigger Up/Down adjustment. It should fall back to current Value (5) and add/subtract delta.
	_, _ = m.Update(tea.KeyMsg{Type: tea.KeyUp})
	if m.textInput.Value() != "6" {
		t.Errorf("expected fallback to Value 5 + delta = 6 on invalid input, got %q", m.textInput.Value())
	}
}

func TestCommandModel_HelpTextForNumberInput(t *testing.T) {
	numVal := &inputs.NumberValue{Value: 5, MinValue: 1, MaxValue: 10}
	input := &inputs.Input{
		Name:  "rating",
		Value: numVal,
	}
	m := &commandModel{
		mode:       modeInputPrompt,
		missing:    []*inputs.Input{input},
		inputIndex: 0,
	}
	m.initCurrentInput()

	viewStr := m.View()
	if !strings.Contains(viewStr, "[Up/Down] +/-") {
		t.Errorf("expected help text to contain '[Up/Down] +/-', got %q", viewStr)
	}

	// For a string input it should NOT contain '[Up/Down] +/-'
	strInput := &inputs.Input{
		Name:  "name",
		Value: &inputs.StringValue{Value: "test"},
	}
	mStr := &commandModel{
		mode:       modeInputPrompt,
		missing:    []*inputs.Input{strInput},
		inputIndex: 0,
	}
	mStr.initCurrentInput()
	viewStrStr := mStr.View()
	if strings.Contains(viewStrStr, "[Up/Down] +/-") {
		t.Errorf("expected string input help text NOT to contain '[Up/Down] +/-', got %q", viewStrStr)
	}
}

func TestCommandModel_BooleanInputToggle(t *testing.T) {
	boolVal := &inputs.BooleanValue{Value: true}
	input := &inputs.Input{
		Name:  "confirm",
		Value: boolVal,
	}
	m := &commandModel{
		mode:       modeInputPrompt,
		missing:    []*inputs.Input{input},
		inputIndex: 0,
	}
	m.initCurrentInput()

	// Initial selection should be "true" (index 0) because default is true
	if m.optionsIndex != 0 {
		t.Errorf("expected optionsIndex to be 0 for default true, got %d", m.optionsIndex)
	}

	// Verify TUI view rendering for boolean input
	viewStr := m.View()
	if !strings.Contains(viewStr, "confirm:") {
		t.Errorf("expected view to contain 'confirm:', got %q", viewStr)
	}
	if !strings.Contains(viewStr, "❯ true") {
		t.Errorf("expected view to highlight 'true' option, got %q", viewStr)
	}

	// Press KeyDown to select "false" (index 1)
	_, _ = m.Update(tea.KeyMsg{Type: tea.KeyDown})
	if m.optionsIndex != 1 {
		t.Errorf("expected optionsIndex to be 1 after KeyDown, got %d", m.optionsIndex)
	}

	// Press KeyDown again to cycle back to "true" (index 0)
	_, _ = m.Update(tea.KeyMsg{Type: tea.KeyDown})
	if m.optionsIndex != 0 {
		t.Errorf("expected optionsIndex to cycle back to 0, got %d", m.optionsIndex)
	}

	// Press KeyUp to go to "false" (index 1)
	_, _ = m.Update(tea.KeyMsg{Type: tea.KeyUp})
	if m.optionsIndex != 1 {
		t.Errorf("expected optionsIndex to be 1 after KeyUp, got %d", m.optionsIndex)
	}

	// Confirm selection by pressing KeyEnter
	_, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if boolVal.Value != false {
		t.Errorf("expected boolVal to be set to false after confirmation, got %v", boolVal.Value)
	}
}

func TestCommandModel_Init(t *testing.T) {
	m := &commandModel{}
	cmd := m.Init()
	if cmd == nil {
		t.Error("expected Init to return a non-nil cmd")
	}
}

func TestCommandModel_CommandSelectMode(t *testing.T) {
	rootCmd := Command{
		Name:        "root",
		Description: "root command",
		Commands: SubCommands{
			{Command: Command{Name: "sub1", Description: "sub1 desc"}},
			{Command: Command{Name: "sub2", Description: "sub2 desc"}},
		},
	}
	history := []Selection{
		{
			commands: []Command{rootCmd},
		},
	}
	m := &commandModel{
		title:         "Choose command",
		mode:          modeCommandSelect,
		history:       history,
		selectedIndex: 0,
	}

	// View rendering check
	viewStr := m.View()
	if !strings.Contains(viewStr, "Choose command") {
		t.Errorf("expected view to contain title, got %q", viewStr)
	}
	if !strings.Contains(viewStr, "❯ sub1") {
		t.Errorf("expected view to highlight active sub1, got %q", viewStr)
	}

	// Update arrow down -> select sub2
	_, _ = m.Update(tea.KeyMsg{Type: tea.KeyDown})
	if m.selectedIndex != 1 {
		t.Errorf("expected selectedIndex to be 1, got %d", m.selectedIndex)
	}

	// KeyDown wraps cyclic
	_, _ = m.Update(tea.KeyMsg{Type: tea.KeyDown})
	if m.selectedIndex != 0 {
		t.Errorf("expected selectedIndex to wrap back to 0, got %d", m.selectedIndex)
	}

	// KeyUp wraps cyclic
	_, _ = m.Update(tea.KeyMsg{Type: tea.KeyUp})
	if m.selectedIndex != 1 {
		t.Errorf("expected selectedIndex to wrap to 1, got %d", m.selectedIndex)
	}

	// Esc pops history (none left since we are at root, should not crash/change history length)
	_, _ = m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	if len(m.history) != 1 {
		t.Errorf("expected history length to remain 1, got %d", len(m.history))
	}

	// Ctrl+C aborts
	_, _ = m.Update(tea.KeyMsg{Type: tea.KeyCtrlC})
	if !m.aborted {
		t.Error("expected aborted to be true")
	}
	if m.View() != "" {
		t.Error("expected aborted TUI to render empty view")
	}
}

func TestCommandModel_InputPromptModeTransition(t *testing.T) {
	input1 := &inputs.Input{Name: "user", Value: &inputs.StringValue{Value: "marty"}}
	input2 := &inputs.Input{Name: "pass", Value: &inputs.StringValue{}}
	m := &commandModel{
		mode:       modeInputPrompt,
		missing:    []*inputs.Input{input1, input2},
		inputIndex: 0,
	}
	m.initCurrentInput()

	// Initial value of input1 is pre-populated
	if m.textInput.Value() != "marty" {
		t.Errorf("expected pre-populated 'marty', got %q", m.textInput.Value())
	}

	// Press Enter to confirm first input and transition
	_, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if m.inputIndex != 1 {
		t.Errorf("expected inputIndex to transition to 1, got %d", m.inputIndex)
	}

	// Press Esc to return to previous input index 0
	_, _ = m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	if m.inputIndex != 0 {
		t.Errorf("expected Esc to navigate back to 0, got %d", m.inputIndex)
	}

	// Press Esc at index 0 to return to command selection mode
	m.history = []Selection{{}} // dummy history
	_, _ = m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	if m.mode != modeCommandSelect {
		t.Error("expected Esc at first input to swap mode back to command select")
	}
}

func TestCommandModel_CommandSelectionTransitions(t *testing.T) {
	rootCmd := Command{
		Name:        "root",
		Description: "root command",
		Commands: SubCommands{
			{Command: Command{Name: "runnable-leaf", Run: "echo RunnableLeaf"}},
			{Command: Command{Name: "nested-parent", Commands: SubCommands{
				{Command: Command{Name: "leaf", Run: "echo Leaf"}},
			}}},
		},
	}
	history := []Selection{
		{
			commands: []Command{rootCmd},
		},
	}
	m := &commandModel{
		title:         "Choose",
		mode:          modeCommandSelect,
		history:       history,
		selectedIndex: 0,
	}

	// 1. Enter on a runnable leaf command directly (should set done = true and exit)
	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if !m.done {
		t.Error("expected done to be true after selecting runnable leaf")
	}
	if cmd == nil {
		t.Error("expected tea.Quit command, got nil")
	}

	// Verify done exit TUI view rendering
	viewStrDone := m.View()
	if !strings.Contains(viewStrDone, "Command: runnable-leaf") {
		t.Errorf("expected done TUI to print chosen command path, got %q", viewStrDone)
	}

	// 2. Select a parent subcommand (should not run/finish, but push history selection)
	m.done = false
	m.selectedIndex = 1 // select nested-parent
	_, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if m.done {
		t.Error("expected done to be false after selecting non-runnable nested-parent")
	}
	if len(m.history) != 2 {
		t.Errorf("expected history length to be 2, got %d", len(m.history))
	}
}

type mockProgram struct {
	RunFunc func() (tea.Model, error)
}

func (mp mockProgram) Run() (tea.Model, error) {
	if mp.RunFunc != nil {
		return mp.RunFunc()
	}
	return nil, nil
}

func TestAskCommands_Mocked(t *testing.T) {
	oldNewProgram := newProgram
	defer func() { newProgram = oldNewProgram }()

	rootCmd := Command{
		Name:        "root",
		Description: "root command",
		Commands: SubCommands{
			{Command: Command{Name: "sub", Run: "echo sub"}},
		},
	}
	sel := Selection{
		commands: []Command{rootCmd},
	}

	newProgram = func(m tea.Model) programRunner {
		return mockProgram{
			RunFunc: func() (tea.Model, error) {
				// Mutate model to simulate selection
				cm := m.(*commandModel)
				cm.history = append(cm.history, cm.currentSelection().SelectCommand(rootCmd.Commands[0].Command, nil))
				return cm, nil
			},
		}
	}

	res, err := askCommands(sel, nil)
	assert.NoError(t, err)
	assert.True(t, res.Runnable())
	assert.Equal(t, "sub", res.commands[len(res.commands)-1].Name)
}

func TestCommandModel_SelectableBooleanInput(t *testing.T) {
	input := &inputs.Input{
		Name:  "selectable-bool",
		Value: &inputs.BooleanValue{},
		Options: inputs.InputOptions{
			{Label: "Yes Please", Value: "true"},
			{Label: "No Thanks", Value: "false"},
		},
	}
	m := &commandModel{}
	opts := m.getBooleanOptions(input)
	assert.Len(t, opts, 2)
	assert.Equal(t, "Yes Please", opts[0].Label)
}

func TestCommandModel_CommandSelectMode_Truncation(t *testing.T) {
	rootCmd := Command{
		Name:        "root",
		Description: "root command",
		Commands: SubCommands{
			{Command: Command{
				Name:        "dock",
				Description: "The Dock is a prominent feature of macOS. It is used to launch applications.",
			}},
		},
	}
	history := []Selection{
		{
			commands: []Command{rootCmd},
		},
	}
	m := &commandModel{
		title:         "Choose command",
		mode:          modeCommandSelect,
		history:       history,
		selectedIndex: 0,
		width:         40, // narrow width to force truncation
	}

	viewStr := m.View()
	// dock name length is 4. maxLen is 4. maxLen + 9 is 13.
	// wrapWidth is 40 - 13 = 27.
	// The truncated description should be 27 runes total including ellipsis.
	// "The Dock is a prominent ..." (27 runes)
	assert.Contains(t, viewStr, "The Dock is a prominent ...")
}

func TestCommandModel_CommandSelectMode_Pagination(t *testing.T) {
	rootCmd := Command{
		Name:        "root",
		Description: "root command",
		Commands: SubCommands{
			{Command: Command{Name: "sub1", Description: "sub1 desc"}},
			{Command: Command{Name: "sub2", Description: "sub2 desc"}},
			{Command: Command{Name: "sub3", Description: "sub3 desc"}},
			{Command: Command{Name: "sub4", Description: "sub4 desc"}},
			{Command: Command{Name: "sub5", Description: "sub5 desc"}},
			{Command: Command{Name: "sub6", Description: "sub6 desc"}},
			{Command: Command{Name: "sub7", Description: "sub7 desc"}},
		},
	}
	history := []Selection{
		{
			commands: []Command{rootCmd},
		},
	}
	m := &commandModel{
		title:         "Choose command",
		mode:          modeCommandSelect,
		history:       history,
		selectedIndex: 5,
		width:         80,
		height:        24,
	}

	viewStr := m.View()
	// Should contain scroll up indicator
	assert.Contains(t, viewStr, "▲  (more above)")
	// Should contain visible items
	assert.Contains(t, viewStr, "sub6")
	assert.Contains(t, viewStr, "sub7")
	// Should NOT contain hidden items
	assert.NotContains(t, viewStr, "sub1")
	assert.NotContains(t, viewStr, "sub2")
}





