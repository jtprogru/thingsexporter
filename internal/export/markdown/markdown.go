// Package markdown implements a Markdown formatter for Export.
package markdown

import (
	"bufio"
	"io"
	"sort"
	"strings"

	"github.com/jtprogru/thingsexporter/internal/export"
	"github.com/jtprogru/thingsexporter/internal/things"
)

// Writer is the export.Writer implementation for Markdown.
type Writer struct{}

// Format returns the format name.
func (Writer) Format() string { return "markdown" }

// Write renders Export as a hierarchy `# Inbox` + `# Areas → ## Area → ### Project → tasks`.
// When there is no Hierarchy (presets without a hierarchy) it produces a flat list of tasks.
func (Writer) Write(out io.Writer, data things.Export, _ export.Options) error {
	w := bufio.NewWriter(out)
	defer func() { _ = w.Flush() }()

	tasksByUUID := make(map[string]things.Task, len(data.Tasks))
	tasksByProject := make(map[string][]things.Task)
	for _, t := range data.Tasks {
		tasksByUUID[t.UUID] = t
		if t.Project != nil {
			tasksByProject[*t.Project] = append(tasksByProject[*t.Project], t)
		}
	}
	for k := range tasksByProject {
		items := tasksByProject[k]
		sort.SliceStable(items, func(i, j int) bool {
			return indexLessNilLast(items[i].Index, items[j].Index)
		})
		tasksByProject[k] = items
	}

	if data.Hierarchy != nil {
		writeInbox(w, data.Hierarchy.InboxOrOrphanTasks, tasksByUUID)
		writeAreas(w, data.Hierarchy.Areas, tasksByUUID, tasksByProject)
		return w.Flush()
	}

	// Flat mode: just a list of tasks.
	for _, t := range data.Tasks {
		writeTaskLine(w, t)
	}
	return w.Flush()
}

func writeInbox(w *bufio.Writer, items []things.HierarchyItem, tasksByUUID map[string]things.Task) {
	_, _ = w.WriteString("# Inbox\n\n")
	if len(items) == 0 {
		_, _ = w.WriteString("_empty_\n\n")
		return
	}
	for _, it := range items {
		if t, ok := tasksByUUID[it.UUID]; ok {
			writeTaskLine(w, t)
		} else {
			writeHierarchyItem(w, it)
		}
	}
	_, _ = w.WriteString("\n")
}

func writeAreas(w *bufio.Writer, areas []things.HierarchyArea, tasksByUUID map[string]things.Task, tasksByProject map[string][]things.Task) {
	_, _ = w.WriteString("# Areas\n\n")
	for _, a := range areas {
		title := stringOr(a.Title, "(untitled)")
		_, _ = w.WriteString("## " + title + "\n\n")
		for _, it := range a.Items {
			t, ok := tasksByUUID[it.UUID]
			if !ok {
				writeHierarchyItem(w, it)
				continue
			}
			// Project — render as ### with its tasks
			if t.TypeName != nil && *t.TypeName == "project" {
				_, _ = w.WriteString("### " + stringOr(t.Title, "(untitled)") + "\n\n")
				for _, child := range tasksByProject[t.UUID] {
					writeTaskLine(w, child)
				}
				_, _ = w.WriteString("\n")
				continue
			}
			writeTaskLine(w, t)
		}
		_, _ = w.WriteString("\n")
	}
}

func writeHierarchyItem(w *bufio.Writer, it things.HierarchyItem) {
	mark := checkbox(it.StatusName)
	_, _ = w.WriteString("- " + mark + " " + stringOr(it.Title, "(untitled)") + "\n")
}

func writeTaskLine(w *bufio.Writer, t things.Task) {
	mark := checkbox(t.StatusName)
	title := stringOr(t.Title, "(untitled)")
	suffix := buildSuffix(t)
	_, _ = w.WriteString("- " + mark + " " + title + suffix + "\n")
	if t.Notes != nil && *t.Notes != "" {
		for _, line := range strings.Split(*t.Notes, "\n") {
			_, _ = w.WriteString("    " + line + "\n")
		}
	}
	if len(t.Checklist) > 0 {
		for _, c := range t.Checklist {
			cmark := checkbox(c.StatusName)
			ctitle := stringOr(c.Title, "(untitled)")
			_, _ = w.WriteString("  - " + cmark + " " + ctitle + "\n")
		}
	}
}

func buildSuffix(t things.Task) string {
	parts := []string{}
	if len(t.Tags) > 0 {
		tags := make([]string, 0, len(t.Tags))
		for _, tg := range t.Tags {
			if tg.Title == nil {
				continue
			}
			tags = append(tags, "#"+*tg.Title)
		}
		if len(tags) > 0 {
			parts = append(parts, strings.Join(tags, " "))
		}
	}
	if t.DeadlineISO != nil {
		parts = append(parts, "⏰ "+*t.DeadlineISO)
	}
	if len(parts) == 0 {
		return ""
	}
	return "  " + strings.Join(parts, " ")
}

func checkbox(status *string) string {
	if status == nil {
		return "[ ]"
	}
	switch *status {
	case "completed":
		return "[x]"
	case "canceled":
		return "[-]"
	default:
		return "[ ]"
	}
}

func stringOr(p *string, def string) string {
	if p == nil {
		return def
	}
	return *p
}

func indexLessNilLast(a, b *int64) bool {
	switch {
	case a == nil && b == nil:
		return false
	case a == nil:
		return false
	case b == nil:
		return true
	default:
		return *a < *b
	}
}
