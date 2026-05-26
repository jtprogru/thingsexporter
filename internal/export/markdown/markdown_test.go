package markdown_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/jtprogru/thingsexporter/internal/export"
	md "github.com/jtprogru/thingsexporter/internal/export/markdown"
	"github.com/jtprogru/thingsexporter/internal/things"
	"github.com/stretchr/testify/require"
)

func sp(s string) *string { return &s }
func si(v int64) *int64   { return &v }

func sample() things.Export {
	openName := "open"
	completedName := "completed"
	canceledName := "canceled"
	todoName := "todo"
	projectName := "project"

	tasks := []things.Task{
		{UUID: "t1", Title: sp("buy milk"), TypeName: &todoName, StatusName: &openName,
			Area:        sp("A-work"),
			Tags:        []things.TagRef{{UUID: "T-p1", Title: sp("P1")}, {UUID: "T-w", Title: sp("work")}},
			DeadlineISO: sp("2024-10-28")},
		{UUID: "t2", Title: sp("garbage"), TypeName: &todoName, StatusName: &completedName,
			Area: sp("A-work")},
		{UUID: "p1", Title: sp("Build deck"), TypeName: &projectName, StatusName: &openName,
			Area: sp("A-home")},
		{UUID: "tc", Title: sp("cancelled thing"), TypeName: &todoName, StatusName: &canceledName,
			Area: sp("A-home")},
		{UUID: "ti", Title: sp("call dentist"), TypeName: &todoName, StatusName: &openName,
			Notes: sp("ring at 9am\nbring insurance card")},
		{UUID: "child", Title: sp("buy wood"), TypeName: &todoName, StatusName: &openName,
			Project: sp("p1")},
	}

	// task with checklist
	tasks[0].Checklist = []things.ChecklistItem{
		{UUID: "cl1", Title: sp("brand"), StatusName: &openName, Index: si(0)},
		{UUID: "cl2", Title: sp("expiry"), StatusName: &completedName, Index: si(1)},
	}

	return things.Export{
		Schema: things.SchemaVersion,
		Tasks:  tasks,
		Hierarchy: &things.Hierarchy{
			Areas: []things.HierarchyArea{
				{
					UUID: "A-work", Title: sp("Work"),
					Items: []things.HierarchyItem{
						{UUID: "t1", Title: sp("buy milk"), TypeName: &todoName, StatusName: &openName},
						{UUID: "t2", Title: sp("garbage"), TypeName: &todoName, StatusName: &completedName},
					},
				},
				{
					UUID: "A-home", Title: sp("Home"),
					Items: []things.HierarchyItem{
						{UUID: "p1", Title: sp("Build deck"), TypeName: &projectName, StatusName: &openName},
						{UUID: "tc", Title: sp("cancelled thing"), TypeName: &todoName, StatusName: &canceledName},
					},
				},
			},
			InboxOrOrphanTasks: []things.HierarchyItem{
				{UUID: "ti", Title: sp("call dentist"), TypeName: &todoName, StatusName: &openName},
			},
		},
	}
}

func render(t testing.TB) string {
	t.Helper()
	var buf bytes.Buffer
	require.NoError(t, md.Writer{}.Write(&buf, sample(), export.Options{}))
	return buf.String()
}

func TestMarkdownWriter_format(t *testing.T) {
	t.Parallel()
	require.Equal(t, "markdown", md.Writer{}.Format())
}

func TestMarkdownWriter_inboxAndAreas(t *testing.T) {
	t.Parallel()
	s := render(t)
	require.True(t, strings.Contains(s, "# Inbox"))
	require.True(t, strings.Contains(s, "# Areas"))
	require.True(t, strings.Contains(s, "## Work"))
	require.True(t, strings.Contains(s, "## Home"))
}

func TestMarkdownWriter_checkboxes(t *testing.T) {
	t.Parallel()
	s := render(t)
	require.Contains(t, s, "- [ ] buy milk")
	require.Contains(t, s, "- [x] garbage")
	require.Contains(t, s, "- [-] cancelled thing")
}

func TestMarkdownWriter_projectAsSubHeading(t *testing.T) {
	t.Parallel()
	s := render(t)
	require.Contains(t, s, "### Build deck")
	require.Contains(t, s, "- [ ] buy wood")
}

func TestMarkdownWriter_tagsAndDeadline(t *testing.T) {
	t.Parallel()
	s := render(t)
	require.Contains(t, s, "- [ ] buy milk  #P1 #work ⏰ 2024-10-28")
}

func TestMarkdownWriter_notesIndent(t *testing.T) {
	t.Parallel()
	s := render(t)
	require.Contains(t, s, "- [ ] call dentist")
	require.Contains(t, s, "    ring at 9am")
	require.Contains(t, s, "    bring insurance card")
}

func TestMarkdownWriter_checklistNested(t *testing.T) {
	t.Parallel()
	s := render(t)
	require.Contains(t, s, "  - [ ] brand")
	require.Contains(t, s, "  - [x] expiry")
}
