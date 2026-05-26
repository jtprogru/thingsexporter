package cli_test

import (
	encjson "encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestInspectCmd_outputsCounts(t *testing.T) {
	t.Parallel()
	deps, stdout, _ := newTestDeps(t)
	require.NoError(t, runCmd(t, deps, "inspect"))
	var payload struct {
		Path            string         `json:"path"`
		DatabaseVersion *int           `json:"databaseVersion"`
		Counts          map[string]int `json:"counts"`
	}
	require.NoError(t, encjson.Unmarshal(stdout.Bytes(), &payload))
	require.NotEmpty(t, payload.Path)
	require.NotNil(t, payload.DatabaseVersion)
	require.Equal(t, 26, *payload.DatabaseVersion)
	require.Equal(t, 2, payload.Counts["areas"])
	require.Equal(t, 3, payload.Counts["tasks"])
	require.Equal(t, 2, payload.Counts["tags"])
}
