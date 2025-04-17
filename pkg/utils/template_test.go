package utils

import (
	"strings"
	"testing"
)

func TestRenderTemplate(t *testing.T) {
	tmpl := "Hello, {{.Name}}!"
	data := struct {
		Name string
	}{
		Name: "World",
	}

	result, err := RenderTemplate(tmpl, data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := "Hello, World!"
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

func TestRenderTemplateWithFuncs(t *testing.T) {
	tmpl := "Hello, {{.Name | upper}}!"
	data := struct {
		Name string
	}{
		Name: "World",
	}

	funcs := map[string]any{
		"upper": func(s string) string {
			return strings.ToUpper(s)
		},
	}

	result, err := RenderTemplateWithFuncs(tmpl, data, funcs)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := "Hello, WORLD!"
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

func TestValidateTemplate(t *testing.T) {
	tmpl := "Hello, {{.Name}}!"
	if err := ValidateTemplate(tmpl); err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	tmpl = "Hello, {{.Name" // Invalid template
	if err := ValidateTemplate(tmpl); err == nil {
		t.Error("expected error, got nil")
	}
}

func TestValidateTemplateWithFuncs(t *testing.T) {
	tmpl := "Hello, {{.Name | upper}}!"
	funcs := map[string]any{
		"upper": func(s string) string {
			return strings.ToUpper(s)
		},
	}

	if err := ValidateTemplateWithFuncs(tmpl, funcs); err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	tmpl = "Hello, {{.Name | upper" // Invalid template
	if err := ValidateTemplateWithFuncs(tmpl, funcs); err == nil {
		t.Error("expected error, got nil")
	}
}
