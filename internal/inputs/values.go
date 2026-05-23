package inputs

import (
	"errors"
	"math"
	"regexp"
	"strconv"
)

var (
	ErrInvalidValue = errors.New("invalid value")
)

type Value interface {
	String() string
	Set(string) error
	Get() any
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
