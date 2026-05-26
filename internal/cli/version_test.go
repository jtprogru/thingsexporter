package cli_test

import (
	"regexp"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestVersionCmd_outputFormat(t *testing.T) {
	t.Parallel()
	deps, stdout, _ := newTestDeps(t)
	require.NoError(t, runCmd(t, deps, "version"))
	require.Regexp(t, regexp.MustCompile(`^thingsexporter \S+\n  commit:\s+\S+`), stdout.String())
}
