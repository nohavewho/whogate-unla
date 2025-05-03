package template

import (
	"bytes"
	"text/template"
)

// Renderer is responsible for rendering templates
type Renderer struct {
	templates map[string]*template.Template
}

// NewRenderer creates a new template renderer
func NewRenderer() *Renderer {
	return &Renderer{
		templates: make(map[string]*template.Template),
	}
}

// Render renders a template with the given context
func (r *Renderer) Render(tmpl string, ctx *Context) (string, error) {
	t, ok := r.templates[tmpl]
	if !ok {
		var err error
		t, err = template.New("").Funcs(template.FuncMap{
			"env": ctx.Env,
		}).Parse(tmpl)
		if err != nil {
			return "", err
		}
		r.templates[tmpl] = t
	}

	var buf bytes.Buffer
	if err := t.Execute(&buf, ctx); err != nil {
		return "", err
	}

	return buf.String(), nil
}
