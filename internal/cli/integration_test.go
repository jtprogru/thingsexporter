package cli_test

import (
	encjson "encoding/json"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

// TestIntegration_jsonExport_fixture — REQ-8.2: a full JSON export of the fixture
// returns meaningful counts and fields.
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

// TestIntegration_structureExport_fixture — structure preset: areas + tags +
// hierarchy, without the tasks collection and without the tasks count in meta.counts.
func TestIntegration_structureExport_fixture(t *testing.T) {
	t.Parallel()
	deps, stdout, _ := newTestDeps(t)
	require.NoError(t, runCmd(t, deps, "--include", "structure"))

	var payload struct {
		Schema string `json:"schema"`
		Meta   struct {
			Counts map[string]int `json:"counts"`
		} `json:"meta"`
		Areas     []map[string]any `json:"areas"`
		Tags      []map[string]any `json:"tags"`
		Tasks     []map[string]any `json:"tasks"`
		Hierarchy struct {
			Areas              []map[string]any `json:"areas"`
			InboxOrOrphanTasks []map[string]any `json:"inbox_or_orphan_tasks"`
		} `json:"hierarchy"`
		Links any `json:"links"`
	}
	require.NoError(t, encjson.Unmarshal(stdout.Bytes(), &payload))

	require.Equal(t, "thingsexporter/v1", payload.Schema)
	require.Len(t, payload.Areas, 2)
	require.Len(t, payload.Tags, 2)
	require.Nil(t, payload.Tasks, "structure preset must drop tasks collection")
	require.Nil(t, payload.Links, "structure preset must drop links")
	require.Len(t, payload.Hierarchy.Areas, 2)
	require.Equal(t, 2, payload.Meta.Counts["areas"])
	require.Equal(t, 2, payload.Meta.Counts["tags"])
	_, hasTasksCount := payload.Meta.Counts["tasks"]
	require.False(t, hasTasksCount, "structure preset must omit Counts.Tasks")
}

// TestIntegration_markdownExport_fixture — REQ-8.3: Markdown contains
// hierarchical headings.
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
	// hierarchy already sorts Areas by Index — Work (-100) comes first
	require.Less(t, strings.Index(s, "## Work"), strings.Index(s, "## Home"))
}
