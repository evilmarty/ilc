package inputs

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
)

func TestTuiModel_NumberInputAdjustment(t *testing.T) {
	numVal := &NumberValue{Value: 3, MinValue: 1, MaxValue: 5}
	input := &Input{
		Name:  "rating",
		Value: numVal,
	}
	m := &tuiModel{
		inputs:       []*Input{input},
		currentIndex: 0,
	}
	m.initCurrentInput()

	// Initially pre-populated
	assert.Equal(t, "3", m.textInput.Value())

	// Press KeyUp -> increment to 4 (default 3 + 1)
	_, _ = m.Update(tea.KeyMsg{Type: tea.KeyUp})
	assert.Equal(t, "4", m.textInput.Value())

	// Press KeyUp -> increment to 5
	_, _ = m.Update(tea.KeyMsg{Type: tea.KeyUp})
	assert.Equal(t, "5", m.textInput.Value())

	// Press KeyUp -> clamp to 5
	_, _ = m.Update(tea.KeyMsg{Type: tea.KeyUp})
	assert.Equal(t, "5", m.textInput.Value())

	// Press KeyDown -> decrement to 4
	_, _ = m.Update(tea.KeyMsg{Type: tea.KeyDown})
	assert.Equal(t, "4", m.textInput.Value())

	// Decrement down to clamp 1
	_, _ = m.Update(tea.KeyMsg{Type: tea.KeyDown}) // 3
	_, _ = m.Update(tea.KeyMsg{Type: tea.KeyDown}) // 2
	_, _ = m.Update(tea.KeyMsg{Type: tea.KeyDown}) // 1
	_, _ = m.Update(tea.KeyMsg{Type: tea.KeyDown}) // 1 (clamp)
	assert.Equal(t, "1", m.textInput.Value())
}

func TestTuiModel_NumberInputAdjustment_NoMax(t *testing.T) {
	numVal := &NumberValue{Value: 1, MinValue: 1, MaxValue: 0}
	input := &Input{
		Name:  "quantity",
		Value: numVal,
	}
	m := &tuiModel{
		inputs:       []*Input{input},
		currentIndex: 0,
	}
	m.initCurrentInput()

	// Send KeyUp (should increment to 2)
	_, _ = m.Update(tea.KeyMsg{Type: tea.KeyUp})
	assert.Equal(t, "2", m.textInput.Value())

	// Send KeyDown (should decrement to 1)
	_, _ = m.Update(tea.KeyMsg{Type: tea.KeyDown})
	assert.Equal(t, "1", m.textInput.Value())

	// Send KeyDown (should clamp to 1)
	_, _ = m.Update(tea.KeyMsg{Type: tea.KeyDown})
	assert.Equal(t, "1", m.textInput.Value())
}

func TestTuiModel_BooleanInputToggle(t *testing.T) {
	boolVal := &BooleanValue{Value: false}
	input := &Input{
		Name:  "confirm",
		Value: boolVal,
	}
	m := &tuiModel{
		inputs:       []*Input{input},
		currentIndex: 0,
	}
	m.initCurrentInput()

	// Initial selection should be "false" (index 1) because default is false
	assert.Equal(t, 1, m.optionsIndex)

	// Press KeyUp to select "true" (index 0)
	_, _ = m.Update(tea.KeyMsg{Type: tea.KeyUp})
	assert.Equal(t, 0, m.optionsIndex)

	// Press KeyUp again to cycle to "false" (index 1)
	_, _ = m.Update(tea.KeyMsg{Type: tea.KeyUp})
	assert.Equal(t, 1, m.optionsIndex)

	// Press KeyDown to select "true" (index 0)
	_, _ = m.Update(tea.KeyMsg{Type: tea.KeyDown})
	assert.Equal(t, 0, m.optionsIndex)

	// Confirm selection by pressing KeyEnter
	_, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	assert.True(t, boolVal.Value)
}

