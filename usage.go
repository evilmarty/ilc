package main

import (
	"flag"
	"fmt"
	"io"
	"strings"
)

type Usage struct {
	Title        string
	Description  string
	commandNames []string
	commandDescs []string
	inputNames   []string
	inputDescs   []string
	flagNames    []string
	flagDescs    []string
	Entrypoint   []string
}

func (u Usage) printSection(b io.Writer, header, content string) {
	content = strings.ReplaceAll(content, "\n", "\n  ")
	content = strings.TrimSuffix(content, "  ")
	fmt.Fprintf(b, "%s\n  %s\n", header, content)
}

func (u Usage) printInstructions(b io.Writer, names, descs []string, header, prefix string) {
	var s strings.Builder
	col := 15
	for _, name := range names {
		col = max(col, len([]rune(name)))
	}
	col = col + 5
	format := fmt.Sprintf("%%-%ds %%s\n", col)
	for i, name := range names {
		desc := descs[i]
		name = fmt.Sprintf("%s%s", prefix, name)
		fmt.Fprintf(&s, format, name, desc)
	}
	u.printSection(b, header, s.String())
}

func (u Usage) usage() string {
	params := []string{}
	if len(u.Entrypoint) > 0 {
		params = append(params, u.Entrypoint[0])
	}
	if len(u.flagNames) > 0 {
		params = append(params, "[flags]")
	}
	if len(u.Entrypoint) > 1 {
		params = append(params, u.Entrypoint[1:]...)
	} else {
		params = append(params, "<config>")
	}
	if len(u.commandNames) > 0 {
		params = append(params, "<commands>")
	}
	if len(u.inputNames) > 0 {
		params = append(params, "[inputs]")
	}
	return strings.Join(params, " ")
}

func (u Usage) String() string {
	var b strings.Builder
	if s := u.Title; s != "" {
		fmt.Fprintf(&b, "%s\n\n", s)
	}
	if s := u.Description; s != "" {
		fmt.Fprintf(&b, "%s\n\n", s)
	}
	if s := u.usage(); s != "" {
		u.printSection(&b, "USAGE", s)
		b.WriteString("\n")
	}
	if len(u.commandNames) > 0 {
		u.printInstructions(&b, u.commandNames, u.commandDescs, "COMMANDS", "")
	}
	if len(u.inputNames) > 0 {
		u.printInstructions(&b, u.inputNames, u.inputDescs, "INPUTS", "-")
	}
	if len(u.flagNames) > 0 {
		u.printInstructions(&b, u.flagNames, u.flagDescs, "FLAGS", "-")
	}
	b.WriteString("\n")
	return b.String()
}

func (u Usage) ImportCommands(commands []ConfigCommand) Usage {
	for _, command := range commands {
		u.commandNames = append(u.commandNames, command.Name)
		u.commandDescs = append(u.commandDescs, command.Description)
	}
	return u
}

func (u Usage) ImportInputs(inputs []ConfigInput) Usage {
	for _, input := range inputs {
		u.inputNames = append(u.inputNames, input.Name)
		u.inputDescs = append(u.inputDescs, input.Description)
	}
	return u
}

func (u Usage) ImportCommandSet(cs CommandSet) Usage {
	return u.ImportCommands(cs.Subcommands()).ImportInputs(cs.Inputs())
}

func NewUsage(entrypoint []string, title, desc string) Usage {
	u := Usage{
		Title:       title,
		Description: desc,
		Entrypoint:  entrypoint,
	}
	mainFlagSet.VisitAll(func(f *flag.Flag) {
		u.flagNames = append(u.flagNames, f.Name)
		u.flagDescs = append(u.flagDescs, f.Usage)
	})
	return u
}
