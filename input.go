package main

import (
	"errors"
	"flag"
	"math"
	"reflect"
	"regexp"
	"slices"
	"strconv"
	"strings"

	orderedmap "github.com/wk8/go-ordered-map/v2"
)

const (
	NumberPrecision = 5
)

var (
	ErrInvalidValue  = errors.New("invalid value")
	ErrInvalidOption = errors.New("invalid option")
)

type InputOption struct {
	Label string
	Value string
}

func (option InputOption) String() string {
	return option.Label
}

type InputOptions []InputOption

func (options InputOptions) Contains(value string) bool {
	for _, item := range options {
		if item.Value == value {
			return true
		}
	}
	return false
}

type Value interface {
	Get() any
	Set(string) error
	String() string
	Kind() reflect.Kind
}

type StringValue struct {
	Value   string `yaml:"default"`
	Pattern string
}

func (v StringValue) Kind() reflect.Kind {
	return reflect.String
}

func (v StringValue) String() string {
	return v.Value
}

func (v StringValue) Get() any {
	return v.Value
}

func (v *StringValue) Set(s string) error {
	if v.Pattern != "" {
		if matched, err := regexp.MatchString(v.Pattern, s); err != nil {
			return err
		} else if !matched {
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

func (v NumberValue) Kind() reflect.Kind {
	return reflect.Float64
}

func (v NumberValue) String() string {
	prec := NumberPrecision
	if v.Value == math.Round(v.Value) {
		prec = 0
	}
	return strconv.FormatFloat(v.Value, 'f', prec, 64)
}

func (v NumberValue) Get() any {
	return any(v.Value)
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

// See https://pkg.go.dev/flag#Value about IsBoolFlag
func (v BooleanValue) IsBoolFlag() bool {
	return true
}

func (v BooleanValue) Kind() reflect.Kind {
	return reflect.Bool
}

func (v BooleanValue) String() string {
	return strconv.FormatBool(v.Value)
}

func (v BooleanValue) Get() any {
	return any(v.Value)
}

func (v *BooleanValue) Set(s string) error {
	b, err := strconv.ParseBool(s)
	if err != nil {
		return err
	}
	v.Value = b
	return nil
}

type Input struct {
	Name        string `yaml:"-"`
	Description string
	Options     InputOptions `yaml:",flow"`
	Value       Value
}

func (input Input) EnvName() string {
	return strings.ReplaceAll(input.Name, "-", "_")
}

func (input Input) Selectable() bool {
	return len(input.Options) > 0
}

type Inputs []Input

func (inputs Inputs) Get(name string) any {
	for _, input := range inputs {
		for input.Name == name {
			return input.Value.Get()
		}
	}
	return nil
}

func (inputs Inputs) GetAll() map[string]any {
	values := map[string]any{}
	for _, input := range inputs {
		values[input.Name] = input.Value.Get()
	}
	return values
}

func (inputs Inputs) Set(name string, value string) error {
	for _, input := range inputs {
		for input.Name == name {
			return input.Value.Set(value)
		}
	}
	return nil
}

func (inputs Inputs) SetAll(values map[string]string) error {
	var errs []error
	for name, value := range values {
		if err := inputs.Set(name, value); err != nil {
			errs = append(errs, err)
		}
	}
	return errors.Join(errs...)
}

func (inputs Inputs) Has(name string) bool {
	return slices.ContainsFunc(inputs, func(input Input) bool {
		return input.Name == name
	})
}

func (inputs Inputs) Merge(other Inputs) Inputs {
	omap := orderedmap.New[string, Input]()
	for _, input := range inputs {
		omap.Set(input.Name, input)
	}
	for _, input := range other {
		omap.Set(input.Name, input)
	}
	result := make(Inputs, 0, omap.Len())
	for pair := omap.Oldest(); pair != nil; pair = pair.Next() {
		result = append(result, pair.Value)
	}
	return result
}

func (inputs Inputs) FlagSet() *flag.FlagSet {
	flag := flag.NewFlagSet("", flag.ContinueOnError)
	for _, input := range inputs {
		flag.Var(input.Value, input.Name, input.Description)
	}
	return flag
}

func (inputs Inputs) ToEnvMap() EnvMap {
	em := EnvMap{}
	for _, input := range inputs {
		em[input.EnvName()] = input.Value.String()
	}
	return em
}
