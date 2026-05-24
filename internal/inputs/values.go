package inputs

import (
	"errors"
	"math"
	"regexp"
	"strconv"
	"strings"
)

var (
	ErrInvalidValue = errors.New("invalid value")
)

type Value interface {
	String() string
	Set(string) error
	Get() any
}

type LiveValidator interface {
	ValidateLive(string) error
}

type AdjustableValue interface {
	Adjust(currentVal string, delta float64) (string, error)
}

func isIncompleteNumber(s string) bool {
	if s == "" || s == "-" || s == "+" || s == "." || s == "-." || s == "+." {
		return true
	}
	if strings.HasSuffix(s, "e") || strings.HasSuffix(s, "E") ||
		strings.HasSuffix(s, "e-") || strings.HasSuffix(s, "E-") ||
		strings.HasSuffix(s, "e+") || strings.HasSuffix(s, "E+") {
		return true
	}
	if strings.HasSuffix(s, ".") {
		return true
	}
	return false
}

type StringValue struct {
	Value   string `yaml:"default"`
	Pattern string
}

func (v StringValue) String() string {
	return v.Value
}

func (v StringValue) Get() any {
	return v.Value
}

func (v *StringValue) Set(s string) error {
	if v.Pattern != "" {
		matched, err := regexp.MatchString(v.Pattern, s)
		if err != nil {
			return err
		}
		if !matched {
			return ErrInvalidValue
		}
	}
	v.Value = s
	return nil
}

func (v StringValue) ValidateLive(s string) error {
	temp := &StringValue{Pattern: v.Pattern}
	return temp.Set(s)
}

type NumberValue struct {
	Value    float64 `yaml:"default"`
	MinValue float64 `yaml:"min"`
	MaxValue float64 `yaml:"max"`
}

func (v NumberValue) String() string {
	prec := 5
	if v.Value == math.Round(v.Value) {
		prec = 0
	}
	return strconv.FormatFloat(v.Value, 'f', prec, 64)
}

func (v NumberValue) Get() any {
	return v.Value
}

func (v *NumberValue) Set(s string) error {
	n, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return err
	}
	if v.MinValue < v.MaxValue {
		if n < v.MinValue || n > v.MaxValue {
			return ErrInvalidValue
		}
	} else if v.MinValue > v.MaxValue && v.MaxValue == 0.0 {
		if n < v.MinValue {
			return ErrInvalidValue
		}
	}
	v.Value = n
	return nil
}

func (v NumberValue) ValidateLive(s string) error {
	if isIncompleteNumber(s) {
		return nil
	}
	temp := &NumberValue{MinValue: v.MinValue, MaxValue: v.MaxValue}
	return temp.Set(s)
}

func (v *NumberValue) Adjust(currentVal string, delta float64) (string, error) {
	if currentVal == "" {
		currentVal = v.String()
	}

	n, err := strconv.ParseFloat(currentVal, 64)
	if err != nil {
		n = v.Value
	}

	newVal := n + delta

	// Clamping to min/max bounds if defined
	if v.MinValue < v.MaxValue {
		if newVal < v.MinValue {
			newVal = v.MinValue
		} else if newVal > v.MaxValue {
			newVal = v.MaxValue
		}
	} else if v.MinValue > v.MaxValue && v.MaxValue == 0.0 {
		if newVal < v.MinValue {
			newVal = v.MinValue
		}
	}

	// Format back cleanly (avoiding .00000 decimals for integers)
	prec := 5
	if newVal == math.Round(newVal) {
		prec = 0
	}
	newStr := strconv.FormatFloat(newVal, 'f', prec, 64)

	// Dry-run check
	temp := &NumberValue{MinValue: v.MinValue, MaxValue: v.MaxValue}
	return newStr, temp.Set(newStr)
}

type BooleanValue struct {
	Value bool `yaml:"default"`
}

func (v BooleanValue) String() string {
	return strconv.FormatBool(v.Value)
}

func (v BooleanValue) Get() any {
	return v.Value
}

func (v *BooleanValue) Set(s string) error {
	b, err := strconv.ParseBool(s)
	if err != nil {
		return err
	}
	v.Value = b
	return nil
}

func (v BooleanValue) ValidateLive(s string) error {
	temp := &BooleanValue{}
	return temp.Set(s)
}
