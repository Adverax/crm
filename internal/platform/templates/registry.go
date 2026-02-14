package templates

import "fmt"

// Registry holds all available application templates.
type Registry struct {
	templates map[string]Template
	order     []string
}

// NewRegistry creates an empty template registry.
func NewRegistry() *Registry {
	return &Registry{
		templates: make(map[string]Template),
	}
}

// Register adds a template to the registry.
func (r *Registry) Register(t Template) {
	if _, exists := r.templates[t.ID]; exists {
		panic(fmt.Sprintf("template %q already registered", t.ID))
	}
	r.templates[t.ID] = t
	r.order = append(r.order, t.ID)
}

// Get returns a template by ID.
func (r *Registry) Get(id string) (Template, bool) {
	t, ok := r.templates[id]
	return t, ok
}

// List returns all registered templates in registration order.
func (r *Registry) List() []Template {
	result := make([]Template, 0, len(r.order))
	for _, id := range r.order {
		result = append(result, r.templates[id])
	}
	return result
}
