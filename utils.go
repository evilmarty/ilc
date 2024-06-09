package main

import (
	"strings"
	"text/template"
)

var renderTemplate = template.New("")

type TemplateData struct {
	Input map[string]any
	Env   map[string]string
}

func NewTemplateData(input map[string]any, env []string) TemplateData {
	return TemplateData{
		Input: input,
		Env:   EnvMap(env),
	}
}

func EnvMap(env []string) map[string]string {
	m := make(map[string]string, len(env))
	for _, item := range env {
		entry := strings.SplitN(item, "=", 2)
		if len(entry) > 1 {
			m[entry[0]] = entry[1]
		} else {
			m[entry[0]] = ""
		}
	}
	return m
}

func RenderTemplate(text string, data TemplateData) (string, error) {
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
