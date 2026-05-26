package preset_test

import (
	"testing"

	"github.com/jtprogru/thingsexporter/internal/export/preset"
	"github.com/jtprogru/thingsexporter/internal/things"
	"github.com/stretchr/testify/require"
)

func sp(s string) *string { return &s }
func si(v int64) *int64   { return &v }

func full() things.Export {
	return things.Export{
		Schema: things.SchemaVersion,
		Areas:  []things.Area{{UUID: "A1", Title: sp("Work"), Tags: []things.TagRef{{UUID: "T1"}}}},
		Tags:   []things.Tag{{UUID: "T1", Title: sp("P1")}},
		Tasks: []things.Task{
			{
				UUID:         "t1",
				Title:        sp("buy milk"),
				AreaTitle:    sp("Work"),
				ProjectTitle: sp("Errands"),
				HeadingTitle: sp("Today"),
				ContactName:  sp("Alice"),
				Tags:         []things.TagRef{{UUID: "T1", Title: sp("P1")}},
				Checklist:    []things.ChecklistItem{{UUID: "cl1", Index: si(0)}},
			},
		},
		ChecklistItems: []things.ChecklistItem{{UUID: "cl1"}},
		Contacts:       []things.Contact{{UUID: "C1"}},
		Tombstones:     []things.Tombstone{{UUID: "tomb"}},
		Links:          &things.Links{TaskTag: []things.TaskTagLink{{Task: "t1", Tag: "T1"}}},
		Hierarchy:      &things.Hierarchy{},
	}
}

func TestRegistryLookup_unknown(t *testing.T) {
	t.Parallel()
	r := preset.NewRegistry(preset.All{}, preset.Tasks{}, preset.TasksTags{}, preset.TasksProjects{})
	_, err := r.Lookup("foo")
	require.Error(t, err)
	require.Contains(t, err.Error(), "unknown include preset")
	require.Contains(t, err.Error(), `"foo"`)
}

func TestRegistry_NamesSorted(t *testing.T) {
	t.Parallel()
	r := preset.NewRegistry(preset.TasksTags{}, preset.All{}, preset.Tasks{}, preset.TasksProjects{})
	require.Equal(t, []string{"all", "tasks", "tasks+projects", "tasks+tags"}, r.Names())
}

func TestPresetAll_identity(t *testing.T) {
	t.Parallel()
	in := full()
	out := preset.All{}.Apply(in)
	require.Equal(t, in.Areas, out.Areas)
	require.Equal(t, in.Tasks, out.Tasks)
	require.NotNil(t, out.Hierarchy)
	require.NotNil(t, out.Links)
}

func TestPresetTasks_strips(t *testing.T) {
	t.Parallel()
	out := preset.Tasks{}.Apply(full())
	require.Nil(t, out.Areas)
	require.Nil(t, out.Tags)
	require.Nil(t, out.ChecklistItems)
	require.Nil(t, out.Contacts)
	require.Nil(t, out.Tombstones)
	require.Nil(t, out.Links)
	require.Nil(t, out.Hierarchy)
	require.Len(t, out.Tasks, 1)
	require.Nil(t, out.Tasks[0].Tags)
	require.Nil(t, out.Tasks[0].Checklist)
	require.Nil(t, out.Tasks[0].AreaTitle)
	require.Nil(t, out.Tasks[0].ProjectTitle)
	require.Nil(t, out.Tasks[0].HeadingTitle)
	require.Nil(t, out.Tasks[0].ContactName)
	require.NotNil(t, out.Meta.Counts.Tasks)
	require.Equal(t, 1, *out.Meta.Counts.Tasks)
	require.Nil(t, out.Meta.Counts.Areas)
}

func TestPresetTasksTags_strips(t *testing.T) {
	t.Parallel()
	out := preset.TasksTags{}.Apply(full())
	require.Nil(t, out.Areas)
	require.Nil(t, out.ChecklistItems)
	require.Nil(t, out.Contacts)
	require.Nil(t, out.Tombstones)
	require.Nil(t, out.Links)
	require.Nil(t, out.Hierarchy)
	require.Len(t, out.Tags, 1)
	require.Len(t, out.Tasks, 1)
	require.NotNil(t, out.Tasks[0].Tags) // оставлены!
	require.Nil(t, out.Tasks[0].Checklist)
	require.Nil(t, out.Tasks[0].AreaTitle)
	require.NotNil(t, out.Meta.Counts.Tasks)
	require.NotNil(t, out.Meta.Counts.Tags)
	require.Nil(t, out.Meta.Counts.Areas)
}

func TestPresetTasksProjects_strips(t *testing.T) {
	t.Parallel()
	out := preset.TasksProjects{}.Apply(full())
	require.Nil(t, out.Tags)
	require.Nil(t, out.ChecklistItems)
	require.Nil(t, out.Contacts)
	require.Nil(t, out.Tombstones)
	require.Nil(t, out.Links)
	require.Nil(t, out.Hierarchy)
	require.Len(t, out.Areas, 1)
	require.Len(t, out.Tasks, 1)
	require.Nil(t, out.Tasks[0].Tags)
	require.Nil(t, out.Tasks[0].Checklist)
	require.NotNil(t, out.Tasks[0].AreaTitle)
	require.NotNil(t, out.Tasks[0].ProjectTitle)
	require.NotNil(t, out.Meta.Counts.Tasks)
	require.NotNil(t, out.Meta.Counts.Areas)
	require.Nil(t, out.Meta.Counts.Tags)
}
