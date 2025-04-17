package utils

import (
	"bytes"
	"text/template"
)

// RenderTemplate renders a template with the given data
func RenderTemplate(tmpl string, data any) (string, error) {
	t, err := template.New("").Parse(tmpl)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	if err := t.Execute(&buf, data); err != nil {
		return "", err
	}

	return buf.String(), nil
}

// RenderTemplateWithFuncs renders a template with custom functions
func RenderTemplateWithFuncs(tmpl string, data any, funcs template.FuncMap) (string, error) {
	t, err := template.New("").Funcs(funcs).Parse(tmpl)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	if err := t.Execute(&buf, data); err != nil {
		return "", err
	}

	return buf.String(), nil
}

// ValidateTemplate validates a template
func ValidateTemplate(tmpl string) error {
	_, err := template.New("").Parse(tmpl)
	return err
}

// ValidateTemplateWithFuncs validates a template with custom functions
func ValidateTemplateWithFuncs(tmpl string, funcs template.FuncMap) error {
	_, err := template.New("").Funcs(funcs).Parse(tmpl)
	return err
}
