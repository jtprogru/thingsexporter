package preset_test

import (
	"fmt"
	"testing"

	"github.com/jtprogru/thingsexporter/internal/export/preset"
	"github.com/jtprogru/thingsexporter/internal/things"
	"pgregory.net/rapid"
)

// genExport — генератор произвольного Export со случайным количеством areas/tags/tasks,
// каждый task — со всеми обогащёнными полями (Tags/Checklist/*Title/ContactName),
// чтобы preset.Apply имел что вырезать.
func genExport() *rapid.Generator[things.Export] {
	return rapid.Custom(func(t *rapid.T) things.Export {
		nAreas := rapid.IntRange(0, 4).Draw(t, "nAreas")
		nTags := rapid.IntRange(0, 4).Draw(t, "nTags")
		nTasks := rapid.IntRange(0, 6).Draw(t, "nTasks")
		nCL := rapid.IntRange(0, 4).Draw(t, "nCL")
		nContacts := rapid.IntRange(0, 3).Draw(t, "nContacts")
		nTomb := rapid.IntRange(0, 3).Draw(t, "nTomb")
		nTTL := rapid.IntRange(0, 5).Draw(t, "nTTL")

		areas := make([]things.Area, nAreas)
		for i := 0; i < nAreas; i++ {
			title := fmt.Sprintf("area-%d", i)
			areas[i] = things.Area{UUID: fmt.Sprintf("A-%d", i), Title: &title}
		}
		tags := make([]things.Tag, nTags)
		for i := 0; i < nTags; i++ {
			title := fmt.Sprintf("tag-%d", i)
			tags[i] = things.Tag{UUID: fmt.Sprintf("T-%d", i), Title: &title}
		}
		tasks := make([]things.Task, nTasks)
		for i := 0; i < nTasks; i++ {
			title := fmt.Sprintf("task-%d", i)
			area := fmt.Sprintf("area-%d", i%maxInt(nAreas, 1))
			proj := fmt.Sprintf("proj-%d", i)
			head := fmt.Sprintf("head-%d", i)
			contact := fmt.Sprintf("c-%d", i)
			tasks[i] = things.Task{
				UUID:         fmt.Sprintf("task-%d", i),
				Title:        &title,
				AreaTitle:    &area,
				ProjectTitle: &proj,
				HeadingTitle: &head,
				ContactName:  &contact,
				Tags:         []things.TagRef{{UUID: "T-0", Title: &area}},
				Checklist:    []things.ChecklistItem{{UUID: fmt.Sprintf("cl-%d", i)}},
			}
		}
		cl := make([]things.ChecklistItem, nCL)
		for i := 0; i < nCL; i++ {
			cl[i] = things.ChecklistItem{UUID: fmt.Sprintf("cl-%d", i)}
		}
		contacts := make([]things.Contact, nContacts)
		for i := 0; i < nContacts; i++ {
			contacts[i] = things.Contact{UUID: fmt.Sprintf("C-%d", i)}
		}
		tombs := make([]things.Tombstone, nTomb)
		for i := 0; i < nTomb; i++ {
			tombs[i] = things.Tombstone{UUID: fmt.Sprintf("tomb-%d", i)}
		}
		links := &things.Links{}
		for i := 0; i < nTTL; i++ {
			links.TaskTag = append(links.TaskTag,
				things.TaskTagLink{Task: fmt.Sprintf("task-%d", i), Tag: "T-0"})
		}

		return things.Export{
			Schema:         things.SchemaVersion,
			Meta:           things.Meta{Source: "test"},
			Areas:          areas,
			Tags:           tags,
			Tasks:          tasks,
			ChecklistItems: cl,
			Contacts:       contacts,
			Tombstones:     tombs,
			Links:          links,
			Hierarchy:      &things.Hierarchy{},
		}
	})
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// PropPresetExclusions (CP-11): для любого Export каждый пресет вырезает
// ровно те поля, что заявлены в его контракте, и сохраняет ровно те, что заявлены.
func PropPresetExclusions(t *rapid.T) {
	in := genExport().Draw(t, "in")

	// All — identity по коллекциям.
	all := preset.All{}.Apply(in)
	if len(all.Areas) != len(in.Areas) || len(all.Tasks) != len(in.Tasks) ||
		len(all.Tags) != len(in.Tags) {
		t.Fatalf("All preset must preserve all collections")
	}
	if all.Hierarchy == nil || all.Links == nil {
		t.Fatalf("All preset must preserve Hierarchy and Links")
	}

	// Tasks — только Tasks; на каждой задаче — никаких enrich-полей.
	tasks := preset.Tasks{}.Apply(in)
	if tasks.Areas != nil || tasks.Tags != nil || tasks.ChecklistItems != nil ||
		tasks.Contacts != nil || tasks.Tombstones != nil ||
		tasks.Links != nil || tasks.Hierarchy != nil {
		t.Fatalf("Tasks preset must strip all collections except tasks")
	}
	if len(tasks.Tasks) != len(in.Tasks) {
		t.Fatalf("Tasks preset must preserve task count: got %d want %d",
			len(tasks.Tasks), len(in.Tasks))
	}
	for i, tk := range tasks.Tasks {
		if tk.Tags != nil || tk.Checklist != nil ||
			tk.AreaTitle != nil || tk.ProjectTitle != nil ||
			tk.HeadingTitle != nil || tk.ContactName != nil {
			t.Fatalf("Tasks preset must strip enrich-fields on task[%d]", i)
		}
	}

	// TasksTags — Tasks с Tags + коллекция Tags; всё остальное nil.
	tt := preset.TasksTags{}.Apply(in)
	if tt.Areas != nil || tt.ChecklistItems != nil || tt.Contacts != nil ||
		tt.Tombstones != nil || tt.Links != nil || tt.Hierarchy != nil {
		t.Fatalf("TasksTags preset must strip all collections except tasks+tags")
	}
	if len(tt.Tags) != len(in.Tags) {
		t.Fatalf("TasksTags must preserve tags")
	}
	for i, tk := range tt.Tasks {
		if tk.Checklist != nil || tk.AreaTitle != nil ||
			tk.ProjectTitle != nil || tk.HeadingTitle != nil ||
			tk.ContactName != nil {
			t.Fatalf("TasksTags must strip non-tag enrich-fields on task[%d]", i)
		}
		// Tags должны сохраниться, если они были в исходном.
		if len(in.Tasks[i].Tags) > 0 && tk.Tags == nil {
			t.Fatalf("TasksTags must preserve task[%d].Tags", i)
		}
	}

	// Structure — Areas + Tags + Hierarchy без Tasks и связанных коллекций.
	st := preset.Structure{}.Apply(in)
	if st.Tasks != nil || st.ChecklistItems != nil || st.Contacts != nil ||
		st.Tombstones != nil || st.Links != nil {
		t.Fatalf("Structure preset must drop Tasks and all task-related collections")
	}
	if len(st.Areas) != len(in.Areas) {
		t.Fatalf("Structure must preserve areas: got %d want %d", len(st.Areas), len(in.Areas))
	}
	if len(st.Tags) != len(in.Tags) {
		t.Fatalf("Structure must preserve tags: got %d want %d", len(st.Tags), len(in.Tags))
	}
	if in.Hierarchy != nil && st.Hierarchy == nil {
		t.Fatalf("Structure must preserve Hierarchy when input has one")
	}
	if st.Meta.Counts.Tasks != nil {
		t.Fatalf("Structure must not set Counts.Tasks")
	}
	if st.Meta.Counts.Areas == nil || *st.Meta.Counts.Areas != len(in.Areas) {
		t.Fatalf("Structure Counts.Areas must equal len(Areas)")
	}
	if st.Meta.Counts.Tags == nil || *st.Meta.Counts.Tags != len(in.Tags) {
		t.Fatalf("Structure Counts.Tags must equal len(Tags)")
	}

	// TasksProjects — Tasks с *Title + коллекция Areas; tags/checklist на задачах nil.
	tp := preset.TasksProjects{}.Apply(in)
	if tp.Tags != nil || tp.ChecklistItems != nil || tp.Contacts != nil ||
		tp.Tombstones != nil || tp.Links != nil || tp.Hierarchy != nil {
		t.Fatalf("TasksProjects preset must strip all collections except tasks+areas")
	}
	if len(tp.Areas) != len(in.Areas) {
		t.Fatalf("TasksProjects must preserve areas")
	}
	for i, tk := range tp.Tasks {
		if tk.Tags != nil || tk.Checklist != nil {
			t.Fatalf("TasksProjects must strip Tags/Checklist on task[%d]", i)
		}
		if in.Tasks[i].AreaTitle != nil && tk.AreaTitle == nil {
			t.Fatalf("TasksProjects must preserve task[%d].AreaTitle", i)
		}
		if in.Tasks[i].ProjectTitle != nil && tk.ProjectTitle == nil {
			t.Fatalf("TasksProjects must preserve task[%d].ProjectTitle", i)
		}
	}
}

func TestPropPresetExclusions(t *testing.T) {
	t.Parallel()
	rapid.Check(t, PropPresetExclusions)
}
