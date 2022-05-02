package main

import (
	"strings"
	"text/template"
)

func RenderTemplate(text string, data map[string]any) (string, error) {
	b := strings.Builder{}
	if tmpl, err := template.New("").Parse(text); err != nil {
		return "", err
	} else if tmpl.Execute(&b, data); err != nil {
		return "", err
	} else {
		return b.String(), nil
	}
}
