package main

import (
	"fmt"
	"math"
	"strings"
	"text/template"
)

var defaultTemplateFuncs = template.FuncMap{
	"contains":   strings.Contains,
	"startswith": strings.HasPrefix,
	"endswith":   strings.HasSuffix,
}
var renderTemplate = template.New("").Funcs(defaultTemplateFuncs)

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

func NewTemplateData(input map[string]any, env EnvMap) TemplateData {
	safeInputs := make(map[string]any)
	for name, value := range input {
		safeName := strings.ReplaceAll(name, "-", "_")
		safeInputs[safeName] = value
	}
	return TemplateData{
		Input: safeInputs,
		Env:   env,
	}
}

type EnvMap map[string]string

func (em EnvMap) Merge(other map[string]string) EnvMap {
	for name, value := range other {
		em[name] = value
	}
	return em
}

func (em EnvMap) Prefix(prefix string) EnvMap {
	result := EnvMap{}
	for key, value := range em {
		var b strings.Builder
		b.WriteString(prefix)
		b.WriteString(key)
		result[b.String()] = value
	}
	return result
}

func (em EnvMap) TrimPrefix(prefix string) EnvMap {
	result := EnvMap{}
	for key, value := range em {
		key = strings.TrimPrefix(key, prefix)
		result[key] = value
	}
	return result
}

func (em EnvMap) FilterPrefix(prefix string) EnvMap {
	result := EnvMap{}
	for key, value := range em {
		if strings.HasPrefix(key, prefix) {
			result[key] = value
		}
	}
	return result
}

func (em EnvMap) ToList() []string {
	var env []string
	for name, value := range em {
		env = append(env, fmt.Sprintf("%s=%s", name, value))
	}
	return env
}

func NewEnvMap(env []string) EnvMap {
	m := make(EnvMap, len(env))
	for _, item := range env {
		key, val, _ := strings.Cut(item, "=")
		m[key] = val
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
		tmpl, err = renderTemplate.Funcs(data.Funcs()).Parse(t)
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

func ToFloat64(value any) float64 {
	switch v := value.(type) {
	case uint:
		return float64(v)
	case int:
		return float64(v)
	case uint16:
		return float64(v)
	case int16:
		return float64(v)
	case uint32:
		return float64(v)
	case int32:
		return float64(v)
	case uint64:
		return float64(v)
	case int64:
		return float64(v)
	case float32:
		return float64(v)
	case float64:
		return v
	default:
		return math.NaN()
	}
}
