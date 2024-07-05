package main

import (
	"flag"
	"fmt"
	"io"
	"strings"

	"github.com/muesli/termenv"
)

type Usage struct {
	Title       string
	Description string
	commands    [][]string
	inputs      [][]string
	flags       [][]string
	Entrypoint  []string
	output      *termenv.Output
}

func (u Usage) printSection(b io.Writer, header, content string) {
	o := u.output
	content = strings.ReplaceAll(content, "\n", "\n  ")
	content = strings.TrimSuffix(content, "  ")
	fmt.Fprintf(b, "%s\n  %s\n",
		o.String(header).Bold(),
		content,
	)
}

func (u Usage) printInstructions(b io.Writer, entries [][]string, header, prefix string) {
	var s strings.Builder
	col := 15
	for _, entry := range entries {
		col = max(col, len([]rune(entry[0])))
	}
	col = col + 5
	format := fmt.Sprintf("%%-%ds %%s\n", col)
	for _, entry := range entries {
		desc := entry[1]
		name := fmt.Sprintf("%s%s", prefix, entry[0])
		fmt.Fprintf(&s, format, name, desc)
	}
	u.printSection(b, header, s.String())
}

func (u Usage) usage() string {
	params := []string{}
	if len(u.Entrypoint) > 0 {
		params = append(params, u.Entrypoint[0])
	}
	if len(u.flags) > 0 {
		params = append(params, "[flags]")
	}
	if len(u.Entrypoint) > 1 {
		params = append(params, u.Entrypoint[1:]...)
	} else {
		params = append(params, "<config>")
	}
	if len(u.commands) > 0 {
		params = append(params, "<commands>")
	}
	if len(u.inputs) > 0 {
		params = append(params, "[inputs]")
	}
	return strings.Join(params, " ")
}

func (u Usage) String() string {
	var b strings.Builder
	o := u.output
	fmt.Fprintf(&b, "\n")
	if s := u.Title; s != "" {
		fmt.Fprintf(&b, "%s\n\n",
			o.String(s).Underline(),
		)
	}
	if s := u.Description; s != "" {
		fmt.Fprintf(&b, "%s\n\n", s)
	}
	if b.Len() > 0 {
		fmt.Fprintf(&b, "\n")
	}
	if s := u.usage(); s != "" {
		u.printSection(&b, "USAGE", s)
		b.WriteString("\n")
	}
	if len(u.commands) > 0 {
		u.printInstructions(&b, u.commands, "COMMANDS", "")
	}
	if len(u.inputs) > 0 {
		u.printInstructions(&b, u.inputs, "INPUTS", "-")
	}
	if len(u.flags) > 0 {
		u.printInstructions(&b, u.flags, "FLAGS", "-")
	}
	b.WriteString("\n")
	return b.String()
}

func (u Usage) Print() error {
	_, err := u.output.WriteString(u.String())
	return err
}

func (u Usage) ImportCommands(commands []ConfigCommand) Usage {
	for _, command := range commands {
		u.commands = append(u.commands, append([]string{command.Name, command.Description}, command.Aliases...))
	}
	return u
}

func (u Usage) ImportInputs(inputs []ConfigInput) Usage {
	for _, input := range inputs {
		u.inputs = append(u.inputs, []string{input.Name, input.Description})
	}
	return u
}

func (u Usage) ImportCommandSet(cs CommandSet) Usage {
	return u.ImportCommands(cs.Subcommands()).ImportInputs(cs.Inputs())
}

func NewUsage(tty io.Writer) Usage {
	u := Usage{
		Title:  "ILC",
		output: termenv.NewOutput(tty),
	}
	mainFlagSet.VisitAll(func(f *flag.Flag) {
		u.flags = append(u.flags, []string{f.Name, f.Usage})
	})
	return u
}
