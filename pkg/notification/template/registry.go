package template

import (
	"fmt"
	"sync"
)

// Registry holds all registered Go-defined notification event templates.
// Templates are registered at startup via init() functions in each template file.
type Registry struct {
	mu        sync.RWMutex
	templates map[string]EventTemplate // keyed by slug
}

// GlobalRegistry is the singleton registry used by the application.
// Register templates via GlobalRegistry.Register() in builtin/*.go init() functions.
var GlobalRegistry = &Registry{
	templates: make(map[string]EventTemplate),
}

// Register adds an EventTemplate to the registry.
// Panics if a template with the same slug is already registered — duplicate slugs
// are a programming error and should be caught at startup, not silently ignored.
func (r *Registry) Register(t EventTemplate) {
	r.mu.Lock()
	defer r.mu.Unlock()

	slug := t.Slug()
	if _, exists := r.templates[slug]; exists {
		panic(fmt.Sprintf("notification template: duplicate slug %q registered", slug))
	}
	r.templates[slug] = t
}

// Get retrieves a registered template by slug.
// Returns the template and true if found, nil and false if not registered.
func (r *Registry) Get(slug string) (EventTemplate, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	t, ok := r.templates[slug]
	return t, ok
}

// MustGet retrieves a registered template by slug or returns an error.
func (r *Registry) MustGet(slug string) (EventTemplate, error) {
	t, ok := r.Get(slug)
	if !ok {
		return nil, fmt.Errorf("notification template: event %q is not registered; add a Go template struct and register it via GlobalRegistry.Register()", slug)
	}
	return t, nil
}

// IsRegistered reports whether an event slug has a registered template.
func (r *Registry) IsRegistered(slug string) bool {
	_, ok := r.Get(slug)
	return ok
}

// Slugs returns all registered event slugs (for admin/debug endpoints).
func (r *Registry) Slugs() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	slugs := make([]string, 0, len(r.templates))
	for slug := range r.templates {
		slugs = append(slugs, slug)
	}
	return slugs
}
