package templates

// BuildRegistry creates a registry with all built-in templates.
func BuildRegistry() *Registry {
	r := NewRegistry()
	r.Register(SalesCRM())
	r.Register(Recruiting())
	return r
}
