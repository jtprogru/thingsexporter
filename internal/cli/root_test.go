package cli_test

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRoot_helpExits0(t *testing.T) {
	t.Parallel()
	deps, stdout, _ := newTestDeps(t)
	require.NoError(t, runCmd(t, deps, "--help"))
	require.Contains(t, stdout.String(), "thingsexporter")
}
