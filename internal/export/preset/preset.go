// Package preset реализует фильтр Export по пресету состава (--include).
package preset

import (
	"fmt"
	"sort"
	"strings"

	"github.com/jtprogru/thingsexporter/internal/things"
)

// Preset — стратегия фильтрации Export.
type Preset interface {
	Name() string
	Apply(in things.Export) things.Export
}

// Registry — реестр пресетов по имени.
type Registry struct {
	presets map[string]Preset
}

// NewRegistry создаёт реестр и регистрирует переданные пресеты.
func NewRegistry(ps ...Preset) *Registry {
	r := &Registry{presets: make(map[string]Preset, len(ps))}
	for _, p := range ps {
		r.Register(p)
	}
	return r
}

// Register добавляет пресет (перезаписывает существующий).
func (r *Registry) Register(p Preset) { r.presets[p.Name()] = p }

// Lookup возвращает пресет по имени.
func (r *Registry) Lookup(name string) (Preset, error) {
	if p, ok := r.presets[name]; ok {
		return p, nil
	}
	return nil, fmt.Errorf("unknown include preset %q (supported: %s)", name, strings.Join(r.Names(), ", "))
}

// Names — алфавитно отсортированный список пресетов.
func (r *Registry) Names() []string {
	out := make([]string, 0, len(r.presets))
	for k := range r.presets {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

// All — пресет "all": идентичный Export.
type All struct{}

func (All) Name() string                         { return "all" }
func (All) Apply(in things.Export) things.Export { return in }

// Tasks — пресет "tasks": только Tasks с enums/dates, без связей.
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

// TasksTags — пресет "tasks+tags": Tasks с tags + коллекция Tags.
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

// Structure — пресет "structure": оглавление выгрузки.
// Возвращает Areas + Tags + Hierarchy без коллекции Tasks и связанных
// объектов (ChecklistItems, Contacts, Tombstones, Links). Полезно
// для быстрого обзора организационной структуры базы Things 3
// без выгрузки тел задач.
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

// TasksProjects — пресет "tasks+projects": Tasks с *Title-полями + коллекция Areas.
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
