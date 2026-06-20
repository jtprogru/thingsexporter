// Package export provides the Writer interface for serializing Export to io.Writer
// and a registry of registered formats.
package export

import (
	"fmt"
	"io"
	"sort"
	"strings"

	"github.com/jtprogru/thingsexporter/internal/things"
)

// Options holds serialization parameters common to all formats.
type Options struct {
	Indent int
}

// Writer is an implementation of a specific format (json, markdown).
type Writer interface {
	Format() string
	Write(out io.Writer, data things.Export, opts Options) error
}

// Registry is a thread-safe (after construction) registry of formats.
type Registry struct {
	writers map[string]Writer
}

// NewRegistry creates a registry and registers the given writers.
func NewRegistry(ws ...Writer) *Registry {
	r := &Registry{writers: make(map[string]Writer, len(ws))}
	for _, w := range ws {
		r.Register(w)
	}
	return r
}

// Register adds a writer (overwrites if the format already exists).
func (r *Registry) Register(w Writer) {
	r.writers[w.Format()] = w
}

// Lookup returns the writer for the given format name.
// For an unknown format it returns an error with a hint.
func (r *Registry) Lookup(format string) (Writer, error) {
	if w, ok := r.writers[format]; ok {
		return w, nil
	}
	return nil, fmt.Errorf("unknown format %q (supported: %s)", format, strings.Join(r.Formats(), ", "))
}

// Formats returns an alphabetically sorted list of registered formats.
func (r *Registry) Formats() []string {
	out := make([]string, 0, len(r.writers))
	for k := range r.writers {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}
