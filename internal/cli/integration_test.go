package cli_test

import (
	encjson "encoding/json"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

// TestIntegration_jsonExport_fixture — REQ-8.2: полный JSON-экспорт фикстуры
// возвращает осмысленные счётчики и поля.
func TestIntegration_jsonExport_fixture(t *testing.T) {
	t.Parallel()
	deps, stdout, _ := newTestDeps(t)
	require.NoError(t, runCmd(t, deps, "--include", "all"))

	var payload struct {
		Schema string `json:"schema"`
		Meta   struct {
			Counts map[string]int `json:"counts"`
		} `json:"meta"`
		Hierarchy struct {
			Areas              []map[string]any `json:"areas"`
			InboxOrOrphanTasks []map[string]any `json:"inbox_or_orphan_tasks"`
		} `json:"hierarchy"`
		Links struct {
			TaskTag []map[string]string `json:"taskTag"`
			AreaTag []map[string]string `json:"areaTag"`
		} `json:"links"`
	}
	require.NoError(t, encjson.Unmarshal(stdout.Bytes(), &payload))

	require.Equal(t, "thingsexporter/v1", payload.Schema)
	require.Equal(t, 2, payload.Meta.Counts["areas"])
	require.Equal(t, 3, payload.Meta.Counts["tasks"])
	require.Equal(t, 2, payload.Meta.Counts["tags"])
	require.Equal(t, 2, payload.Meta.Counts["taskTagLinks"])
	require.Equal(t, 1, payload.Meta.Counts["areaTagLinks"])
	require.Len(t, payload.Hierarchy.Areas, 2)
	require.Len(t, payload.Hierarchy.InboxOrOrphanTasks, 1)
}

// TestIntegration_markdownExport_fixture — REQ-8.3: Markdown содержит
// иерархические заголовки.
func TestIntegration_markdownExport_fixture(t *testing.T) {
	t.Parallel()
	deps, stdout, _ := newTestDeps(t)
	require.NoError(t, runCmd(t, deps, "--format", "markdown"))
	s := stdout.String()
	require.Contains(t, s, "# Inbox")
	require.Contains(t, s, "# Areas")
	require.Contains(t, s, "## Work")
	require.Contains(t, s, "## Home")
	require.Contains(t, s, "### Build deck")
	// hierarchy уже сортирует Areas по Index — Work (-100) идёт первым
	require.Less(t, strings.Index(s, "## Work"), strings.Index(s, "## Home"))
}
