# Inputs Library (`internal/inputs`)

A generic, Bubble Tea-powered input parsing, validation, and terminal form-rendering library. It is designed to be completely decoupled from the `ilc` core application package.

---

## Architecture Overview

The package is structured around three core concepts:

```
┌────────────────────────────────────────────────────────┐
│                        FlagSet                         │
│  (Manages parsing, env merging, and TUI prompts)       │
└───────────────────────────┬────────────────────────────┘
                            │ contains
                            ▼
┌────────────────────────────────────────────────────────┐
│                         Input                          │
│  (Wraps Name, Description, Options, and active Value)  │
└───────────────────────────┬────────────────────────────┘
                            │ implements
                            ▼
┌────────────────────────────────────────────────────────┐
│                    Value (Interface)                   │
│  (String(), Get(), Set(string) error)                  │
│                                                        │
│  Implementations:                                      │
│  • StringValue (with regex Pattern verification)       │
│  • NumberValue (with Min/Max clamping)                 │
│  • BooleanValue                                        │
└────────────────────────────────────────────────────────┘
```

### 1. The `Value` Interface
Any input type must implement the `Value` interface to support parsing and validation:
```go
type Value interface {
    String() string   // Return current clean text representation
    Set(string) error // Parse and validate string input, or return error
    Get() any         // Return raw typed value (e.g. float64, string, bool)
}
```

### 2. The `Input` Model
Defines the input's metadata, whether it represents a list of static selectable `Options`, and its associated `Value` backend:
```go
type Input struct {
    Name        string
    Description string
    Options     InputOptions
    Value       Value
}
```

### 3. The `FlagSet` Controller
Exposes CLI argument parsing (`flag.FlagSet` wrapping), environment variable overrides, and the interactive Bubble Tea terminal wizard.

---

## Interactive TUI Implementation Guidelines

When extending or maintaining the Bubble Tea views (`tui.go` or core application prompts in `prompt.go`), adhere to the following best practices:

### 1. Side-Effect-Free Dry-Run Validation
Always run dry-run validation on every keystroke (`Update(msg)` cycle) without mutating the active, persistent model state. Replicate the validation logic on a temporary struct:
```go
// Example dry-run validation
temp := &inputs.NumberValue{MinValue: min, MaxValue: max}
err := temp.Set(textInput.Value())
m.inputErr = err // Track temporarily in Bubble Tea model
```
This ensures that if the user backs out (`Esc`) or aborts, no invalid parameters are retained.

### 2. Keyboard Number Adjustments (Up/Down Arrow Keys)
For inputs containing numeric values (`NumberValue`), intercept `KeyUp` and `KeyDown` keystrokes to increment or decrement the active value by `1`:
- **Clamping**: Automatically clamp the resulting number within the `MinValue` and `MaxValue` bounds.
- **Incomplete Typings**: Allow users to type intermediate numeric formats (such as single negative signs `"-"` or trailing decimal points `"1."`) without throwing premature validation errors. Use `isIncompleteNumber()` checking:
  ```go
  func isIncompleteNumber(s string) bool {
      return s == "" || s == "-" || s == "+" || s == "." || strings.HasSuffix(s, ".")
  }
  ```
- **Integer Formatting**: Ensure whole integers are cleanly formatted to omit trailing decimals (e.g. rendering `4` instead of `4.00000`).

### 3. Progressive Visual Validation States
Render clear, distinct validation prompts on every frame:
- **Neutral/Empty**: Dark gray prompt symbol (`> `).
- **Valid Entry**: Sage green checkmark (`✔ `) immediately when valid.
- **Invalid Entry**: Coral red cross (`✘ `) immediately when invalid.
- **Help Guidelines**: Dynamically hide the `[Enter] Confirm` and include `[Up/Down] +/-` shortcuts in the footer depending on validation states and value types.
