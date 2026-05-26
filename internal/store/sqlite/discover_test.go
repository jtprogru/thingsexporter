package sqlite_test

import (
	"errors"
	"path/filepath"
	"testing"

	sqlitestore "github.com/jtprogru/thingsexporter/internal/store/sqlite"
	"github.com/stretchr/testify/require"
)

func TestDiscover_matrix(t *testing.T) {
	t.Parallel()
	okFn := func(_ string) error { return nil }
	errFn := func(_ string) error { return errors.New("not found") }

	cases := []struct {
		name    string
		home    string
		goos    string
		statFn  func(string) error
		wantOk  bool
		wantSfx string
	}{
		{"macOS happy path", "/u", "darwin", okFn, true, sqlitestore.DefaultMacOSDBPath},
		{"macOS stat err", "/u", "darwin", errFn, false, ""},
		{"linux always nope", "/u", "linux", okFn, false, ""},
		{"windows always nope", "/u", "windows", okFn, false, ""},
		{"empty home", "", "darwin", okFn, false, ""},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			path, ok := sqlitestore.Discover(tc.home, tc.goos, tc.statFn)
			require.Equal(t, tc.wantOk, ok)
			if tc.wantOk {
				require.Equal(t, filepath.Join(tc.home, tc.wantSfx), path)
			} else {
				require.Empty(t, path)
			}
		})
	}
}
