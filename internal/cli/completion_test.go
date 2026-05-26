package cli_test

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCompletionCmd_bash(t *testing.T) {
	t.Parallel()
	deps, stdout, _ := newTestDeps(t)
	require.NoError(t, runCmd(t, deps, "completion", "bash"))
	require.NotEmpty(t, stdout.String())
	require.Contains(t, stdout.String(), "thingsexporter")
}
