package cli_test

import (
	encjson "encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jtprogru/thingsexporter/internal/cli"
	"github.com/stretchr/testify/require"
)

func TestExportCmd_rootDefaults(t *testing.T) {
	t.Parallel()
	deps, stdout, _ := newTestDeps(t)
	require.NoError(t, runCmd(t, deps))
	var payload map[string]any
	require.NoError(t, encjson.Unmarshal(stdout.Bytes(), &payload))
	require.Equal(t, "thingsexporter/v1", payload["schema"])
	require.NotNil(t, payload["meta"])
	require.NotNil(t, payload["areas"])
}

func TestExportCmd_format_markdown(t *testing.T) {
	t.Parallel()
	deps, stdout, _ := newTestDeps(t)
	require.NoError(t, runCmd(t, deps, "--format", "markdown"))
	s := stdout.String()
	require.Contains(t, s, "# Inbox")
	require.Contains(t, s, "# Areas")
}

func TestExportCmd_includeTasks_strips(t *testing.T) {
	t.Parallel()
	deps, stdout, _ := newTestDeps(t)
	require.NoError(t, runCmd(t, deps, "--include", "tasks"))
	var payload map[string]any
	require.NoError(t, encjson.Unmarshal(stdout.Bytes(), &payload))
	_, hasAreas := payload["areas"]
	require.False(t, hasAreas, "tasks preset must not have areas")
	_, hasTasks := payload["tasks"]
	require.True(t, hasTasks)
}

func TestExportCmd_unknownFormat_exit2(t *testing.T) {
	t.Parallel()
	deps, _, _ := newTestDeps(t)
	err := runCmd(t, deps, "--format", "yaml")
	require.Error(t, err)
	require.Equal(t, 2, cli.AsExitCode(err))
}

func TestExportCmd_unknownInclude_exit2(t *testing.T) {
	t.Parallel()
	deps, _, _ := newTestDeps(t)
	err := runCmd(t, deps, "--include", "foo")
	require.Error(t, err)
	require.Equal(t, 2, cli.AsExitCode(err))
}

func TestExportCmd_missingDBNonMac(t *testing.T) {
	t.Parallel()
	deps, _, _ := newTestDeps(t)
	deps.DiscoverDB = func() (string, bool) { return "", false }
	err := runCmd(t, deps, "--include", "all")
	require.Error(t, err)
	require.Equal(t, 2, cli.AsExitCode(err))
	require.Contains(t, err.Error(), "--db is required")
}

func TestExportCmd_quietSuppresses(t *testing.T) {
	t.Parallel()
	deps, _, stderr := newTestDeps(t)
	require.NoError(t, runCmd(t, deps, "--quiet"))
	require.Empty(t, stderr.String(), "quiet must suppress stderr summary")
}

func TestExportCmd_summary(t *testing.T) {
	t.Parallel()
	deps, _, stderr := newTestDeps(t)
	require.NoError(t, runCmd(t, deps))
	out := stderr.String()
	require.Contains(t, out, "OK -> stdout")
	require.Contains(t, out, "tasks: 3")
}

func TestExportCmd_outToFile(t *testing.T) {
	t.Parallel()
	deps, _, _ := newTestDeps(t)
	outFile := filepath.Join(t.TempDir(), "export.json")
	require.NoError(t, runCmd(t, deps, "--out", outFile))
	body, err := os.ReadFile(outFile)
	require.NoError(t, err)
	// file output uses the default indent=2, hence "schema": "...".
	require.Contains(t, string(body), `"schema": "thingsexporter/v1"`)
}

func TestExportCmd_indentZero_compact(t *testing.T) {
	t.Parallel()
	deps, stdout, _ := newTestDeps(t)
	require.NoError(t, runCmd(t, deps, "--indent", "0"))
	s := stdout.String()
	// With indent=0 there are no internal newlines; the only \n is the final one.
	require.Equal(t, 1, strings.Count(s, "\n"))
}

func TestExportCmd_noBlobs(t *testing.T) {
	t.Parallel()
	deps, stdout, _ := newTestDeps(t)
	require.NoError(t, runCmd(t, deps, "--no-blobs"))
	// our fixture has no non-empty BLOBs, so we check that the output has no __blob_hex__
	require.NotContains(t, stdout.String(), "__blob_hex__")
}

func TestExportCmd_negativeIndent(t *testing.T) {
	t.Parallel()
	deps, _, _ := newTestDeps(t)
	err := runCmd(t, deps, "--indent", "-1")
	require.Error(t, err)
	require.Equal(t, 2, cli.AsExitCode(err))
}
