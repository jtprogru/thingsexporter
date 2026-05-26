// Package export содержит интерфейс Writer для сериализации Export в io.Writer
// и реестр зарегистрированных форматов.
package export

import (
	"fmt"
	"io"
	"sort"
	"strings"

	"github.com/jtprogru/thingsexporter/internal/things"
)

// Options — параметры сериализации, общие для всех форматов.
type Options struct {
	Indent int
}

// Writer — реализация конкретного формата (json, markdown).
type Writer interface {
	Format() string
	Write(out io.Writer, data things.Export, opts Options) error
}

// Registry — потокобезопасный (после конструктора) реестр форматов.
type Registry struct {
	writers map[string]Writer
}

// NewRegistry создаёт реестр и регистрирует переданные writers.
func NewRegistry(ws ...Writer) *Registry {
	r := &Registry{writers: make(map[string]Writer, len(ws))}
	for _, w := range ws {
		r.Register(w)
	}
	return r
}

// Register добавляет writer (перезаписывает, если формат уже есть).
func (r *Registry) Register(w Writer) {
	r.writers[w.Format()] = w
}

// Lookup возвращает writer по имени формата.
// При неизвестном формате — ошибка с подсказкой.
func (r *Registry) Lookup(format string) (Writer, error) {
	if w, ok := r.writers[format]; ok {
		return w, nil
	}
	return nil, fmt.Errorf("unknown format %q (supported: %s)", format, strings.Join(r.Formats(), ", "))
}

// Formats возвращает алфавитно отсортированный список зарегистрированных форматов.
func (r *Registry) Formats() []string {
	out := make([]string, 0, len(r.writers))
	for k := range r.writers {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}
