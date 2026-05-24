package ilc

import "fmt"

// TemplateError represents a syntax or parsing error in a command's run script or environment template.
type TemplateError struct {
	Type      string // "run" or "env"
	Command   string
	FieldName string
	Err       error
}

func (e *TemplateError) Error() string {
	if e.Type == "run" {
		return fmt.Sprintf("invalid run template in command %q: %v", e.Command, e.Err)
	}
	return fmt.Sprintf("invalid env template %q in command %q: %v", e.FieldName, e.Command, e.Err)
}

func (e *TemplateError) Unwrap() error {
	return e.Err
}
