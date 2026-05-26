package export_test

import (
	"io"
	"testing"

	"github.com/jtprogru/thingsexporter/internal/export"
	"github.com/jtprogru/thingsexporter/internal/things"
	"github.com/stretchr/testify/require"
)

type fakeWriter struct{ name string }

func (f fakeWriter) Format() string { return f.name }
func (f fakeWriter) Write(_ io.Writer, _ things.Export, _ export.Options) error {
	return nil
}

func TestRegistryLookup_unknownFormat(t *testing.T) {
	t.Parallel()
	r := export.NewRegistry(fakeWriter{name: "json"}, fakeWriter{name: "markdown"})
	_, err := r.Lookup("yaml")
	require.Error(t, err)
	require.Contains(t, err.Error(), "unknown format")
	require.Contains(t, err.Error(), `"yaml"`)
	require.Contains(t, err.Error(), "json")
	require.Contains(t, err.Error(), "markdown")
}

func TestRegistry_FormatsSorted(t *testing.T) {
	t.Parallel()
	r := export.NewRegistry(fakeWriter{name: "markdown"}, fakeWriter{name: "json"})
	formats := r.Formats()
	require.Equal(t, []string{"json", "markdown"}, formats)
}

func TestRegistryLookup_found(t *testing.T) {
	t.Parallel()
	r := export.NewRegistry(fakeWriter{name: "json"})
	w, err := r.Lookup("json")
	require.NoError(t, err)
	require.Equal(t, "json", w.Format())
	require.NoError(t, w.Write(io.Discard, things.Export{}, export.Options{}))
}
