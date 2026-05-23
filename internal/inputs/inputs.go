package inputs

import (
	"errors"
	"flag"
	"fmt"
	"slices"
	"strings"
)

var (
	ErrAborted = errors.New("aborted")
)

type InputOption struct {
	Label string `yaml:"label"`
	Value string `yaml:"value"`
}

func (o InputOption) String() string {
	return o.Label
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

type Input struct {
	Name        string       `yaml:"-"`
	Description string       `yaml:"description"`
	Options     InputOptions `yaml:"options,flow"`
	Value       Value        `yaml:"value"`
}

func (input Input) EnvName() string {
	return strings.ReplaceAll(strings.ToUpper(input.Name), "-", "_")
}

func (input Input) Selectable() bool {
	return len(input.Options) > 0
}

type trackingValue struct {
	val      Value
	provided *bool
}

func (t trackingValue) String() string {
	return t.val.String()
}

func (t trackingValue) Set(s string) error {
	*t.provided = true
	return t.val.Set(s)
}

type FlagSet struct {
	name           string
	envPrefix      string
	inputs         []*Input
	provided       map[string]*bool
}

func NewFlagSet(name string, envPrefix string) *FlagSet {
	return &FlagSet{
		name:      name,
		envPrefix: envPrefix,
		provided:  make(map[string]*bool),
	}
}

func (fs *FlagSet) Var(input *Input) {
	fs.inputs = append(fs.inputs, input)
	provided := false
	fs.provided[input.Name] = &provided
}

func (fs *FlagSet) Inputs() []*Input {
	return fs.inputs
}

func (fs *FlagSet) ParseEnvAndArgs(args []string, envs map[string]string) ([]*Input, error) {
	// Reset provided status
	for _, provided := range fs.provided {
		*provided = false
	}

	stdFs := flag.NewFlagSet("", flag.ContinueOnError)
	stdFs.Usage = func() {} // Silent usage printout

	// Register variables with tracking wrapper
	for _, input := range fs.inputs {
		provided := fs.provided[input.Name]
		stdFs.Var(trackingValue{val: input.Value, provided: provided}, input.Name, input.Description)
	}

	// 1. Process environment variables
	for _, input := range fs.inputs {
		envName := fs.envPrefix + input.EnvName()
		if envVal, found := envs[envName]; found {
			if err := stdFs.Set(input.Name, envVal); err != nil {
				return nil, fmt.Errorf("invalid environment variable %s: %w", envName, err)
			}
		}
	}

	// 2. Parse command-line flags
	if err := stdFs.Parse(args); err != nil {
		return nil, err
	}

	// 3. Collect missing inputs
	var missing []*Input
	for _, input := range fs.inputs {
		if !*fs.provided[input.Name] {
			missing = append(missing, input)
		}
	}

	return missing, nil
}

func (fs *FlagSet) Parse(args []string, envs map[string]string, nonInteractive bool) error {
	missing, err := fs.ParseEnvAndArgs(args, envs)
	if err != nil {
		return err
	}

	// 4. Prompt for missing inputs if any
	if len(missing) > 0 {
		if nonInteractive {
			var missingNames []string
			for _, input := range missing {
				missingNames = append(missingNames, input.Name)
			}
			return fmt.Errorf("missing inputs: %s", strings.Join(missingNames, ", "))
		}

		// Prompt using Bubble Tea TUI
		if err := fs.promptTUI(missing); err != nil {
			return err
		}
	}

	return nil
}

func (fs *FlagSet) Values() map[string]any {
	values := make(map[string]any, len(fs.inputs))
	for _, input := range fs.inputs {
		values[input.Name] = input.Value.Get()
	}
	return values
}

func (fs *FlagSet) ToEnvMap() map[string]string {
	em := make(map[string]string, len(fs.inputs))
	for _, input := range fs.inputs {
		envName := fs.envPrefix + input.EnvName()
		em[envName] = input.Value.String()
	}
	return em
}

func (fs *FlagSet) ToArgs() []string {
	args := make([]string, 0, len(fs.inputs))
	for _, input := range fs.inputs {
		if v, ok := input.Value.(*BooleanValue); ok {
			if v.Value {
				args = append(args, fmt.Sprintf("-%s", input.Name))
			} else {
				args = append(args, fmt.Sprintf("-%s=%s", input.Name, v.String()))
			}
		} else {
			args = append(args, fmt.Sprintf("-%s", input.Name))
			args = append(args, input.Value.String())
		}
	}
	return args
}

func (fs *FlagSet) Merge(other *FlagSet) *FlagSet {
	merged := NewFlagSet(fs.name, fs.envPrefix)

	for _, input := range fs.inputs {
		merged.Var(input)
	}
	for _, input := range other.inputs {
		found := false
		for i, existing := range merged.inputs {
			if existing.Name == input.Name {
				merged.inputs[i] = input
				found = true
				break
			}
		}
		if !found {
			merged.Var(input)
		}
	}
	return merged
}

func (fs *FlagSet) Has(name string) bool {
	return slices.ContainsFunc(fs.inputs, func(input *Input) bool {
		return input.Name == name
	})
}
