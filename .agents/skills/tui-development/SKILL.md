---
name: TUI Development
description: Guidelines for building, styling, and debugging Bubble Tea and Lipgloss components inside ILC.
---

# TUI Development Skill

This skill guides AI coding assistants through extending, styling, and refining interactive terminal prompts inside the `ilc` codebase using the **Bubble Tea** (`github.com/charmbracelet/bubbletea`) and **Lipgloss** (`github.com/charmbracelet/lipgloss`) libraries.

---

## 1. Architecture Constraints
* **Generic Isolation**: Keep `internal/inputs` generic and free from application-specific dependencies in `internal/ilc`. It must function as a standalone command line inputs library.
* **Unified UI**: Do not import third-party prompt libraries (e.g. `promptkit`, `survey`). All interactive dialogs must be rendered using unified Lipgloss styles.

---

## 2. Bubble Tea Best Practices

### Use Pointer Receivers
To propagate selection states, view indices, or errors correctly back to the caller when `p.Run()` returns, **always** implement the model's Bubble Tea methods with pointer receivers:
```go
// CORRECT
func (m *tuiModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) { ... }

// INCORRECT (Value receiver modifies a copy of the model, breaking output state)
func (m tuiModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) { ... }
```

### Side-Effect-Free Dry-Run Validation
Always run dry-run validation on every keystroke (`Update(msg)` cycle) without mutating the active, persistent model state. Perform validation checks on a temporary copy of the input type:
```go
// Replicate target type validation without side effects
temp := &inputs.NumberValue{MinValue: min, MaxValue: max}
err := temp.Set(textInput.Value())
m.inputErr = err // Store validation error on model
```

### Tolerating Incomplete Numeric Input
Do not show premature validation errors when the user is in the middle of typing a decimal, scientific notation, or negative number. Use `isIncompleteNumber()` checking to suppress errors:
```go
func isIncompleteNumber(s string) bool {
    if s == "" || s == "-" || s == "+" || s == "." || s == "-." || s == "+." {
        return true
    }
    if strings.HasSuffix(s, "e") || strings.HasSuffix(s, "E") ||
       strings.HasSuffix(s, "e-") || strings.HasSuffix(s, "E-") {
        return true
    }
    return false
}
```

---

## 3. Styling & Layout Alignment

### Preventing Lipgloss Offset Padding
Lipgloss treats multi-line styled blocks dynamically. When formatting layouts, **never** include newline characters (`\n` or `\n\n`) inside style render calls. Newlines inside a style block trigger Lipgloss to pad all empty lines with spaces to match the full rendering block width, throwing off left-border alignments of list items:
```go
// CORRECT
sb.WriteString(titleStyle.Render("── Choose command ──") + "\n\n")

// INCORRECT (causes subsequent lines to offset to the right)
sb.WriteString(titleStyle.Render("── Choose command ──\n\n"))
```

### Color Palette Consistency
Always adhere to the established premium dark mode design tokens:
* **Headers & Prompt Labels**: Bold Blue (`Color("32")` or `Color("81")`)
* **Neutral Icons & Placeholders**: Dark Gray (`Color("240")` or `Color("244")`)
* **Valid/Sage Green Icons**: Soft Sage Green (`Color("150")`)
* **Invalid/Coral Red Icons**: Soft Coral Red (`Color("203")`)
