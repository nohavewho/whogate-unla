package template

// Context represents the template context
type (
	Context struct {
		Args    map[string]any `json:"args"`
		Request struct {
			Headers map[string]string `json:"headers"`
			Query   map[string]string `json:"query"`
			Path    map[string]string `json:"path"`
			Body    map[string]any    `json:"body"`
		} `json:"request"`
		Response ResponseWrapper `json:"response"`
	}
	ResponseWrapper struct {
		Data any `json:"data"`
	}
)

// NewContext creates a new template context
func NewContext() *Context {
	return &Context{
		Args: make(map[string]any),
		Request: struct {
			Headers map[string]string `json:"headers"`
			Query   map[string]string `json:"query"`
			Path    map[string]string `json:"path"`
			Body    map[string]any    `json:"body"`
		}{
			Headers: make(map[string]string),
			Query:   make(map[string]string),
			Path:    make(map[string]string),
			Body:    make(map[string]any),
		},
		Response: ResponseWrapper{},
	}
}
