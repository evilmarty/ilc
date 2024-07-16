package main

import (
	"fmt"
	"strings"
	"text/template"
)

var renderTemplate = template.New("")

type TemplateData struct {
	Input map[string]any
	Env   map[string]string
}

func (td TemplateData) getInput(name string) any {
	if v, ok := td.Input[name]; ok {
		return v
	} else {
		return nil
	}
}

func (td TemplateData) getEnv(name string) any {
	if v, ok := td.Env[name]; ok {
		return v
	} else {
		return nil
	}
}

func (td *TemplateData) Funcs() template.FuncMap {
	return template.FuncMap{
		"input": td.getInput,
		"env":   td.getEnv,
	}
}

func NewTemplateData(input map[string]any, env []string) TemplateData {
	safeInputs := make(map[string]any)
	for name, value := range input {
		safeName := strings.ReplaceAll(name, "-", "_")
		safeInputs[safeName] = value
	}
	return TemplateData{
		Input: safeInputs,
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

func RenderTemplate(text interface{}, data TemplateData) (string, error) {
	var tmpl *template.Template
	var err error
	b := strings.Builder{}
	switch t := text.(type) {
	case *template.Template:
		tmpl = t
	case string:
		tmpl, err = renderTemplate.Parse(t)
		if err != nil {
			return "", err
		}
	default:
		return "", fmt.Errorf("unsupported type: %v", t)
	}
	if err = tmpl.Execute(&b, data); err != nil {
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
