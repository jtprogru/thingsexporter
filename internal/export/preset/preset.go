// Package preset implements Export filtering by a content preset (--include).
package preset

import (
	"fmt"
	"sort"
	"strings"

	"github.com/jtprogru/thingsexporter/internal/things"
)

// Preset is a strategy for filtering Export.
type Preset interface {
	Name() string
	Apply(in things.Export) things.Export
}

// Registry is a registry of presets by name.
type Registry struct {
	presets map[string]Preset
}

// NewRegistry creates a registry and registers the given presets.
func NewRegistry(ps ...Preset) *Registry {
	r := &Registry{presets: make(map[string]Preset, len(ps))}
	for _, p := range ps {
		r.Register(p)
	}
	return r
}

// Register adds a preset (overwrites an existing one).
func (r *Registry) Register(p Preset) { r.presets[p.Name()] = p }

// Lookup returns the preset for the given name.
func (r *Registry) Lookup(name string) (Preset, error) {
	if p, ok := r.presets[name]; ok {
		return p, nil
	}
	return nil, fmt.Errorf("unknown include preset %q (supported: %s)", name, strings.Join(r.Names(), ", "))
}

// Names returns an alphabetically sorted list of presets.
func (r *Registry) Names() []string {
	out := make([]string, 0, len(r.presets))
	for k := range r.presets {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

// All is the "all" preset: an identical Export.
type All struct{}

func (All) Name() string                         { return "all" }
func (All) Apply(in things.Export) things.Export { return in }

// Tasks is the "tasks" preset: only Tasks with enums/dates, without relations.
type Tasks struct{}

func (Tasks) Name() string { return "tasks" }
func (Tasks) Apply(in things.Export) things.Export {
	tasks := make([]things.Task, len(in.Tasks))
	for i, t := range in.Tasks {
		t.Tags = nil
		t.Checklist = nil
		t.AreaTitle = nil
		t.ProjectTitle = nil
		t.HeadingTitle = nil
		t.ContactName = nil
		tasks[i] = t
	}
	return things.Export{
		Schema: in.Schema,
		Meta: things.Meta{
			Source:     in.Meta.Source,
			ExportedAt: in.Meta.ExportedAt,
			Counts:     things.Counts{Tasks: intPtr(len(tasks))},
			DBMetaRows: in.Meta.DBMetaRows,
		},
		Tasks: tasks,
	}
}

// TasksTags is the "tasks+tags" preset: Tasks with tags + the Tags collection.
type TasksTags struct{}

func (TasksTags) Name() string { return "tasks+tags" }
func (TasksTags) Apply(in things.Export) things.Export {
	tasks := make([]things.Task, len(in.Tasks))
	for i, t := range in.Tasks {
		t.Checklist = nil
		t.AreaTitle = nil
		t.ProjectTitle = nil
		t.HeadingTitle = nil
		t.ContactName = nil
		tasks[i] = t
	}
	return things.Export{
		Schema: in.Schema,
		Meta: things.Meta{
			Source:     in.Meta.Source,
			ExportedAt: in.Meta.ExportedAt,
			Counts: things.Counts{
				Tasks: intPtr(len(tasks)),
				Tags:  intPtr(len(in.Tags)),
			},
			DBMetaRows: in.Meta.DBMetaRows,
		},
		Tags:  in.Tags,
		Tasks: tasks,
	}
}

// Structure is the "structure" preset: a table of contents for the export.
// Returns Areas + Tags + Hierarchy without the Tasks collection and related
// objects (ChecklistItems, Contacts, Tombstones, Links). Useful
// for a quick overview of the organizational structure of the Things 3 database
// without exporting the task bodies.
type Structure struct{}

func (Structure) Name() string { return "structure" }
func (Structure) Apply(in things.Export) things.Export {
	return things.Export{
		Schema: in.Schema,
		Meta: things.Meta{
			Source:     in.Meta.Source,
			ExportedAt: in.Meta.ExportedAt,
			Counts: things.Counts{
				Areas: intPtr(len(in.Areas)),
				Tags:  intPtr(len(in.Tags)),
			},
			DBMetaRows: in.Meta.DBMetaRows,
		},
		Areas:     in.Areas,
		Tags:      in.Tags,
		Hierarchy: in.Hierarchy,
	}
}

// TasksProjects is the "tasks+projects" preset: Tasks with *Title fields + the Areas collection.
type TasksProjects struct{}

func (TasksProjects) Name() string { return "tasks+projects" }
func (TasksProjects) Apply(in things.Export) things.Export {
	tasks := make([]things.Task, len(in.Tasks))
	for i, t := range in.Tasks {
		t.Tags = nil
		t.Checklist = nil
		tasks[i] = t
	}
	return things.Export{
		Schema: in.Schema,
		Meta: things.Meta{
			Source:     in.Meta.Source,
			ExportedAt: in.Meta.ExportedAt,
			Counts: things.Counts{
				Areas: intPtr(len(in.Areas)),
				Tasks: intPtr(len(tasks)),
			},
			DBMetaRows: in.Meta.DBMetaRows,
		},
		Areas: in.Areas,
		Tasks: tasks,
	}
}

func intPtr(v int) *int { return &v }
