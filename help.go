package main

import (
	"flag"
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

var (
	helpHeadingStyle = lipgloss.NewStyle().
				Width(helpWidth).
				Align(lipgloss.Center).
				Foreground(lipgloss.Color("219")).
				Border(lipgloss.NormalBorder(), false, false, true, false).
				MarginBottom(1).
				Padding(1)
	helpSectionTitleStyle = lipgloss.NewStyle().
				MarginTop(1).
				Width(helpWidth).
				Bold(true)
	helpSectionBodyStyle = lipgloss.NewStyle().
				MaxWidth(helpWidth).
				PaddingLeft(2)
	helpDataListTitle = lipgloss.NewStyle().
				Width(14)
	helpDataListData = lipgloss.NewStyle()
	helpDefaultValue = lipgloss.NewStyle().
				Foreground(lipgloss.Color("8"))
)

func printHelp(m *model) {
	fmt.Println(renderHelp(m))
}

func renderHelp(m *model) string {
	sections := []string{
		helpHeadingStyle.Render("ILC"),
		renderUsage("USAGE", m),
		renderGlobalFlags("GLOBAL FLAGS"),
	}

	if m != nil {
		if m.Selected() {
			sections = append(
				sections,
				renderInputs("FLAGS", m.Inputs()),
			)
		}
		sections = append(
			sections,
			renderCommands("COMMANDS", m.AvailableCommands()),
		)
	}

	output := lipgloss.JoinVertical(0, sections...)

	return fmt.Sprintln(output)
}

func renderUsage(title string, m *model) string {
	commandNames := make([]string, 0)
	if m != nil {
		for _, command := range m.commands {
			commandNames = append(commandNames, command.Name)
		}
		if l := len(m.commands); l > 0 && m.commands[l-1].HasSubCommands() {
			commandNames = append(commandNames, "<subcommand>")
		}
	}
	if len(commandNames) == 0 {
		commandNames = append(commandNames, "<command>")
	}

	return renderSection(
		title,
		fmt.Sprintf("ilc [global flags] %s [flags]", strings.Join(commandNames, " ")),
	)
}

func renderGlobalFlags(title string) string {
	items := make([]string, 0)
	flag.VisitAll(func(f *flag.Flag) {
		items = append(
			items,
			fmt.Sprintf("-%s", f.Name),
			f.Usage,
		)
	})

	return renderSection(
		title,
		renderDataList(items...),
	)
}

func renderCommands(title string, commands ConfigCommands) string {
	commandsCount := len(commands)
	if commandsCount == 0 {
		return ""
	}

	items := make([]string, 0, commandsCount*2)
	for _, command := range commands {
		items = append(items, command.Name, command.Description)
	}

	return renderSection(
		title,
		renderDataList(items...),
	)
}

func renderInputs(title string, inputs Inputs) string {
	inputsCount := len(inputs)
	if inputsCount == 0 {
		return ""
	}

	items := make([]string, 0, inputsCount*2)
	for _, input := range inputs {
		desc := input.Description

		if input.DefaultValue != "" {
			defaultDesc := fmt.Sprintf("%s %s", helpDefaultValue.Render("Default is"), helpDefaultValue.Bold(true).Render(input.DefaultValue))
			if desc == "" {
				desc = defaultDesc
			} else {
				desc = fmt.Sprintf("%s %s", desc, defaultDesc)
			}
		}

		items = append(
			items,
			fmt.Sprintf("-%s", input.Name),
			desc,
		)
	}

	return renderSection(
		title,
		renderDataList(items...),
	)
}

func renderSection(title, body string) string {
	return lipgloss.JoinVertical(
		0,
		helpSectionTitleStyle.Render(title),
		helpSectionBodyStyle.Render(body),
	)
}

func renderDataList(items ...string) string {
	datalist := make([]string, 0, len(items)%2)
	for len(items) > 0 && len(items)%2 == 0 {
		datalist = append(
			datalist,
			lipgloss.JoinHorizontal(
				0,
				helpDataListTitle.Render(items[0]),
				helpDataListData.Render(items[1]),
			),
		)
		items = items[2:]
	}

	return lipgloss.JoinVertical(0, datalist...)
}
