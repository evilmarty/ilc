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

func TestTuiModel_BooleanInputToggle_CustomOptions(t *testing.T) {
	t.Run("array custom options", func(t *testing.T) {
		boolVal := &BooleanValue{Value: true}
		input := &Input{
			Name:  "confirm",
			Value: boolVal,
			Options: InputOptions{
				{Label: "No", Value: "false"},
				{Label: "Yes", Value: "true"},
			},
		}
		m := &tuiModel{
			inputs:       []*Input{input},
			currentIndex: 0,
		}
		m.initCurrentInput()

		// Initial selection should be "true" (index 1 in our options: Yes is true)
		assert.Equal(t, 1, m.optionsIndex)

		// Press KeyUp to cycle to "false" (index 0: No is false)
		_, _ = m.Update(tea.KeyMsg{Type: tea.KeyUp})
		assert.Equal(t, 0, m.optionsIndex)

		// Confirm "false" selection by pressing KeyEnter
		_, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
		assert.False(t, boolVal.Value)
	})

	t.Run("map custom options", func(t *testing.T) {
		boolVal := &BooleanValue{Value: false}
		input := &Input{
			Name:  "confirm",
			Value: boolVal,
			Options: InputOptions{
				{Label: "Absolutely", Value: "true"},
				{Label: "No way", Value: "false"},
			},
		}
		m := &tuiModel{
			inputs:       []*Input{input},
			currentIndex: 0,
		}
		m.initCurrentInput()

		// Initial selection should be "false" (index 1 in our options: No way is false)
		assert.Equal(t, 1, m.optionsIndex)

		// Press KeyDown to cycle to "true" (index 0: Absolutely is true)
		_, _ = m.Update(tea.KeyMsg{Type: tea.KeyDown})
		assert.Equal(t, 0, m.optionsIndex)

		// Confirm "true" selection by pressing KeyEnter
		_, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
		assert.True(t, boolVal.Value)
	})
}

func TestTuiModel_Init(t *testing.T) {
	m := &tuiModel{}
	cmd := m.Init()
	assert.NotNil(t, cmd)
}

func TestTuiModel_Abort(t *testing.T) {
	input := &Input{Name: "name", Value: &StringValue{Value: "test"}}
	m := &tuiModel{
		inputs: []*Input{input},
	}
	m.initCurrentInput()

	// Press Ctrl+C -> should set aborted to true
	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyCtrlC})
	assert.True(t, m.aborted)
	assert.NotNil(t, cmd)

	m.aborted = false
	// Press Esc -> should set aborted to true
	_, cmd = m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	assert.True(t, m.aborted)
	assert.NotNil(t, cmd)

	// View should return empty string if aborted
	assert.Equal(t, "", m.View())
}

func TestTuiModel_ViewSelectable(t *testing.T) {
	input := &Input{
		Name:        "choices",
		Description: "Choose one",
		Options: InputOptions{
			{Label: "OptA", Value: "A"},
			{Label: "OptB", Value: "B"},
		},
		Value: &StringValue{},
	}
	m := &tuiModel{
		inputs: []*Input{input},
	}
	m.initCurrentInput()

	viewStr := m.View()
	assert.Contains(t, viewStr, "[1/1] Choose one")
	assert.Contains(t, viewStr, "❯ OptA")
	assert.Contains(t, viewStr, "OptB")

	// Adjust optionsIndex and check update view
	_, _ = m.Update(tea.KeyMsg{Type: tea.KeyDown})
	viewStr2 := m.View()
	assert.Contains(t, viewStr2, "❯ OptB")
}

func TestTuiModel_ViewStringInputWithError(t *testing.T) {
	input := &Input{
		Name:  "name",
		Value: &StringValue{Pattern: "^[a-z]+$"},
	}
	m := &tuiModel{
		inputs: []*Input{input},
	}
	m.initCurrentInput()

	// Initially no error
	viewStr := m.View()
	assert.Contains(t, viewStr, "Please specify a name")
	assert.NotContains(t, viewStr, "✗ Invalid input")

	// Set invalid value to trigger error
	_, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter}) // triggers empty set which fails pattern
	viewStrErr := m.View()
	assert.Contains(t, viewStrErr, "✗ Invalid input")
}

func TestTuiModel_TransitionMultipleInputs(t *testing.T) {
	input1 := &Input{Name: "first", Value: &StringValue{Value: "A"}}
	input2 := &Input{Name: "second", Value: &StringValue{Value: "B"}}
	m := &tuiModel{
		inputs: []*Input{input1, input2},
	}
	m.initCurrentInput()

	// Initially at index 0
	assert.Equal(t, 0, m.currentIndex)

	// Press Enter to confirm first and transition
	_, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	assert.Equal(t, 1, m.currentIndex)

	// Press Enter to confirm second and finish
	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	assert.NotNil(t, cmd) // Should return tea.Quit
}

func TestTuiModel_NonSelectableKeystrokes(t *testing.T) {
	input := &Input{Name: "name", Value: &StringValue{}}
	m := &tuiModel{
		inputs: []*Input{input},
	}
	m.initCurrentInput()

	// Send normal key msg to update textInput
	_, _ = m.Update(tea.KeyMsg{Runes: []rune("hello"), Type: tea.KeyRunes})
	assert.Equal(t, "hello", m.textInput.Value())
}


