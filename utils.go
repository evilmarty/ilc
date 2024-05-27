package main

import (
	"strings"
	"text/template"
)

var renderTemplate = template.New("")

func RenderTemplate(text string, data map[string]any) (string, error) {
	b := strings.Builder{}
	if tmpl, err := renderTemplate.Parse(text); err != nil {
		return "", err
	} else if tmpl.Execute(&b, data); err != nil {
		return "", err
	} else {
		return b.String(), nil
	}
}

func DiffStrings(a, b []string) []string {
	mb := make(map[string]struct{}, len(b))
	for _, x := range b {
		mb[x] = struct{}{}
	}
	var diff []string
	for _, x := range a {
		if _, found := mb[x]; !found {
			diff = append(diff, x)
		}
	}
	return diff
}
