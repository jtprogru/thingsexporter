package cli_test

import (
	"errors"
	"fmt"
	"io"
	"testing"

	"github.com/jtprogru/thingsexporter/internal/cli"
	"github.com/stretchr/testify/require"
)

func TestAsExitCode_table(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name string
		in   error
		want int
	}{
		{"nil → 0", nil, 0},
		{"plain error → 2", errors.New("x"), 2},
		{"ExitCodeError 2 → 2", &cli.ExitCodeError{Code: 2, Err: io.EOF}, 2},
		{"ExitCodeError 0 → 0", &cli.ExitCodeError{Code: 0, Err: io.EOF}, 0},
		{"wrapped ExitCodeError → 2", fmt.Errorf("wrapped: %w", &cli.ExitCodeError{Code: 2, Err: io.EOF}), 2},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			require.Equal(t, tc.want, cli.AsExitCode(tc.in))
		})
	}
}
