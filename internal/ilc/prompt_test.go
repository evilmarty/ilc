package ilc

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/evilmarty/ilc/internal/inputs"
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

	// Initially, textInput value is empty, so it uses default/placeholder
	if m.textInput.Value() != "" {
		t.Errorf("expected textInput to be initially empty, got %q", m.textInput.Value())
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
