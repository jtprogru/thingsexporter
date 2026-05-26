package things_test

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/jtprogru/thingsexporter/internal/things"
	"github.com/stretchr/testify/require"
)

func sampleRaw() things.RawData {
	return things.RawData{
		Areas: []things.RawArea{
			{UUID: "A-work", Title: ptrString("Work"), Index: ptrInt64(-100)},
			{UUID: "A-home", Title: ptrString("Home"), Index: ptrInt64(50)},
			{UUID: "A-nil", Title: ptrString("Misc")}, // Index=nil — должен пойти в конец
		},
		Tags: []things.RawTag{
			{UUID: "T-p1", Title: ptrString("P1")},
			{UUID: "T-p2", Title: ptrString("P2"), Parent: ptrString("T-p1")},
			{UUID: "T-work", Title: ptrString("work")},
		},
		Tasks: []things.RawTask{
			// task в области Work, не trashed, todo, open
			{UUID: "task-1", Title: ptrString("buy milk"), Area: ptrString("A-work"),
				Type: ptrInt64(0), Status: ptrInt64(0), Index: ptrInt64(10)},
			// trashed — не попадёт в hierarchy
			{UUID: "task-trashed", Title: ptrString("garbage"), Area: ptrString("A-work"),
				Type: ptrInt64(0), Status: ptrInt64(3), Trashed: ptrInt64(1)},
			// project (тип 1) в области Home
			{UUID: "proj-1", Title: ptrString("Build deck"), Area: ptrString("A-home"),
				Type: ptrInt64(1), Status: ptrInt64(0), Index: ptrInt64(5)},
			// inbox/orphan
			{UUID: "task-inbox", Title: ptrString("call dentist"),
				Type: ptrInt64(0), Status: ptrInt64(0), Index: ptrInt64(1)},
			// task внутри проекта — НЕ должен попасть в hierarchy.items
			{UUID: "task-in-proj", Title: ptrString("buy wood"), Project: ptrString("proj-1"),
				Type: ptrInt64(0), Status: ptrInt64(0)},
		},
		Checklist: []things.RawChecklist{
			{UUID: "cl-1", Task: ptrString("task-1"), Title: ptrString("check brand"),
				Status: ptrInt64(0), Index: ptrInt64(0)},
			{UUID: "cl-2", Task: ptrString("task-1"), Title: ptrString("verify expiry"),
				Status: ptrInt64(3), Index: ptrInt64(1)},
		},
		Contacts: []things.RawContact{
			{UUID: "C-1", DisplayName: ptrString("Alice")},
		},
		Tombstones: []things.RawTombstone{
			{UUID: "tomb-1", DeletedObjectUUID: ptrString("gone-uuid"), DeletionDate: ptrFloat64(0)},
		},
		TaskTagPairs: []things.TaskTagLink{
			{Task: "task-1", Tag: "T-p1"},
			{Task: "task-1", Tag: "T-work"},
			{Task: "proj-1", Tag: "T-p2"},
		},
		AreaTagPairs: []things.AreaTagLink{
			{Area: "A-work", Tag: "T-p1"},
		},
		MetaRows: []things.MetaRow{
			{Key: "databaseVersion", Value: "26"},
		},
	}
}

func mustBuild(t testing.TB, opts things.BuildOptions) things.Export {
	t.Helper()
	if opts.ExportedAt.IsZero() {
		opts.ExportedAt = time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	}
	if opts.Source == "" {
		opts.Source = "fixture.sqlite"
	}
	return things.Build(sampleRaw(), opts)
}

func TestBuild_schemaField(t *testing.T) {
	t.Parallel()
	e := mustBuild(t, things.BuildOptions{})
	require.Equal(t, things.SchemaVersion, e.Schema)
}

func TestBuild_enrichTaskTags(t *testing.T) {
	t.Parallel()
	e := mustBuild(t, things.BuildOptions{})
	var task1 *things.Task
	for i := range e.Tasks {
		if e.Tasks[i].UUID == "task-1" {
			task1 = &e.Tasks[i]
		}
	}
	require.NotNil(t, task1, "task-1 not found")
	require.Len(t, task1.Tags, 2)
	titles := []string{*task1.Tags[0].Title, *task1.Tags[1].Title}
	require.Contains(t, titles, "P1")
	require.Contains(t, titles, "work")
}

func TestBuild_areaProjectTitles(t *testing.T) {
	t.Parallel()
	e := mustBuild(t, things.BuildOptions{})
	var task1, taskInProj *things.Task
	for i := range e.Tasks {
		switch e.Tasks[i].UUID {
		case "task-1":
			task1 = &e.Tasks[i]
		case "task-in-proj":
			taskInProj = &e.Tasks[i]
		}
	}
	require.NotNil(t, task1)
	require.NotNil(t, task1.AreaTitle)
	require.Equal(t, "Work", *task1.AreaTitle)

	require.NotNil(t, taskInProj)
	require.NotNil(t, taskInProj.ProjectTitle)
	require.Equal(t, "Build deck", *taskInProj.ProjectTitle)
}

func TestBuild_hierarchy_excludesTrashed(t *testing.T) {
	t.Parallel()
	e := mustBuild(t, things.BuildOptions{})
	require.NotNil(t, e.Hierarchy)
	for _, a := range e.Hierarchy.Areas {
		for _, it := range a.Items {
			require.NotEqual(t, "task-trashed", it.UUID, "trashed task leaked into hierarchy")
		}
	}
	for _, it := range e.Hierarchy.InboxOrOrphanTasks {
		require.NotEqual(t, "task-trashed", it.UUID)
	}
}

func TestBuild_hierarchy_ordering(t *testing.T) {
	t.Parallel()
	e := mustBuild(t, things.BuildOptions{})
	require.NotNil(t, e.Hierarchy)
	require.Len(t, e.Hierarchy.Areas, 3)
	// Index -100 first, then 50, then nil (A-nil)
	require.Equal(t, "A-work", e.Hierarchy.Areas[0].UUID)
	require.Equal(t, "A-home", e.Hierarchy.Areas[1].UUID)
	require.Equal(t, "A-nil", e.Hierarchy.Areas[2].UUID)
}

func TestBuild_inboxContainsOrphans(t *testing.T) {
	t.Parallel()
	e := mustBuild(t, things.BuildOptions{})
	require.NotNil(t, e.Hierarchy)
	var found bool
	for _, it := range e.Hierarchy.InboxOrOrphanTasks {
		if it.UUID == "task-inbox" {
			found = true
		}
	}
	require.True(t, found, "inbox task missing")
}

func TestBuild_counts_match(t *testing.T) {
	t.Parallel()
	e := mustBuild(t, things.BuildOptions{})
	require.NotNil(t, e.Meta.Counts.Areas)
	require.Equal(t, len(e.Areas), *e.Meta.Counts.Areas)
	require.NotNil(t, e.Meta.Counts.Tasks)
	require.Equal(t, len(e.Tasks), *e.Meta.Counts.Tasks)
	require.NotNil(t, e.Meta.Counts.Tags)
	require.Equal(t, len(e.Tags), *e.Meta.Counts.Tags)
	require.NotNil(t, e.Meta.Counts.ChecklistItems)
	require.Equal(t, len(e.ChecklistItems), *e.Meta.Counts.ChecklistItems)
	require.NotNil(t, e.Meta.Counts.TaskTagLinks)
	require.Equal(t, len(e.Links.TaskTag), *e.Meta.Counts.TaskTagLinks)
	require.NotNil(t, e.Meta.Counts.AreaTagLinks)
	require.Equal(t, len(e.Links.AreaTag), *e.Meta.Counts.AreaTagLinks)
}

func TestBuild_tagsParentTitle(t *testing.T) {
	t.Parallel()
	e := mustBuild(t, things.BuildOptions{})
	var p2 *things.Tag
	for i := range e.Tags {
		if e.Tags[i].UUID == "T-p2" {
			p2 = &e.Tags[i]
		}
	}
	require.NotNil(t, p2)
	require.NotNil(t, p2.ParentTitle)
	require.Equal(t, "P1", *p2.ParentTitle)
}

func TestBuild_areaTags(t *testing.T) {
	t.Parallel()
	e := mustBuild(t, things.BuildOptions{})
	var work *things.Area
	for i := range e.Areas {
		if e.Areas[i].UUID == "A-work" {
			work = &e.Areas[i]
		}
	}
	require.NotNil(t, work)
	require.Len(t, work.Tags, 1)
	require.Equal(t, "T-p1", work.Tags[0].UUID)
}

func TestBuild_noBlobs_strips(t *testing.T) {
	t.Parallel()
	raw := sampleRaw()
	raw.Areas[0].CachedTags = []byte{0xff, 0xee}
	e := things.Build(raw, things.BuildOptions{NoBlobs: true})
	var work *things.Area
	for i := range e.Areas {
		if e.Areas[i].UUID == "A-work" {
			work = &e.Areas[i]
		}
	}
	require.NotNil(t, work)
	require.Nil(t, work.CachedTags, "--no-blobs should null out BLOB")
}

func TestBuild_jsonMarshalCompiles(t *testing.T) {
	t.Parallel()
	e := mustBuild(t, things.BuildOptions{})
	b, err := json.Marshal(e)
	require.NoError(t, err)
	require.Contains(t, string(b), `"schema":"`+things.SchemaVersion+`"`)
}
