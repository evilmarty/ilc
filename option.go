package main

import (
	"strings"

	"gopkg.in/yaml.v3"
)

type Option struct {
	Label string
	Value string
}

type Options struct {
	Items  []Option `yaml:"-"`
	Script string   `yaml:"-"`
	Prefix string   `yaml:"-"`
	Suffix string   `yaml:"-"`
}

func (options Options) Empty() bool {
	return len(options.Items) == 0 && options.Script == ""
}

func (options Options) Contains(value any) bool {
	for _, item := range options.Items {
		if item.Value == value {
			return true
		}
	}

	return false
}

func (options *Options) populate() error {
	if options.Script == "" {
		return nil
	}
	bgCmd := BgCommand(ScriptCommand(options.Script))
	bgCmd.Prefix = options.Prefix
	bgCmd.Suffix = options.Suffix
	output, err := bgCmd.Output()
	if err != nil {
		return err
	}
	values := strings.Split(strings.TrimSpace(string(output)), "\n")
	options.Items = newOptions(values)
	return nil
}

func (options Options) Get() ([]Option, error) {
	err := options.populate()
	return options.Items, err
}

func (x *Options) UnmarshalYAML(node *yaml.Node) error {
	var items []Option
	var script string

	switch node.Kind {
	case yaml.SequenceNode:
		var seqValue []string
		if err := node.Decode(&seqValue); err != nil {
			return err
		}
		items = newOptions(seqValue)
	case yaml.MappingNode:
		var mapValue map[string]string
		if err := node.Decode(&mapValue); err != nil {
			return err
		}
		items = make([]Option, 0, len(mapValue))
		for label, value := range mapValue {
			items = append(items, Option{Label: label, Value: value})
		}
	case yaml.ScalarNode:
		script = node.Value
	}

	*x = Options{Items: items, Script: script}

	return nil
}

func newOptions(values []string) []Option {
	items := make([]Option, len(values))
	for i, item := range values {
		items[i].Label = item
		items[i].Value = item
	}
	return items
}
