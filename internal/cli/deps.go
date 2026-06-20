// Package cli is a thin wrapper over cobra that builds the thingsexporter commands.
package cli

import (
	"io"
	"os"
	"runtime"
	"time"

	"github.com/jtprogru/thingsexporter/internal/export"
	jsonwriter "github.com/jtprogru/thingsexporter/internal/export/json"
	mdwriter "github.com/jtprogru/thingsexporter/internal/export/markdown"
	"github.com/jtprogru/thingsexporter/internal/export/preset"
	sqlitestore "github.com/jtprogru/thingsexporter/internal/store/sqlite"
)

// Deps is an explicit test seam. Production runs DefaultDeps(); in tests
// the fields are swapped for buffers and fakes.
type Deps struct {
	Stdout, Stderr io.Writer
	Clock          func() time.Time

	OpenRepo   func(path string) (*sqlitestore.Repository, error)
	DiscoverDB func() (string, bool)

	Writers *export.Registry
	Presets *preset.Registry

	// SupportedDBVersions lists the accepted databaseVersion values;
	// if the actual value is not in the list, a warning is printed.
	SupportedDBVersions []int
}

// DefaultDeps assembles the working dependencies for the main process.
func DefaultDeps() Deps {
	writers := export.NewRegistry(jsonwriter.Writer{}, mdwriter.Writer{})
	presets := preset.NewRegistry(preset.All{}, preset.Structure{}, preset.Tasks{}, preset.TasksTags{}, preset.TasksProjects{})

	openRepo := func(path string) (*sqlitestore.Repository, error) {
		db, err := sqlitestore.Open(path)
		if err != nil {
			return nil, err
		}
		return sqlitestore.NewRepository(db), nil
	}

	discoverDB := func() (string, bool) {
		home, err := os.UserHomeDir()
		if err != nil || home == "" {
			return "", false
		}
		return sqlitestore.Discover(home, runtime.GOOS, func(p string) error {
			_, err := os.Stat(p)
			return err
		})
	}

	return Deps{
		Stdout:              os.Stdout,
		Stderr:              os.Stderr,
		Clock:               time.Now,
		OpenRepo:            openRepo,
		DiscoverDB:          discoverDB,
		Writers:             writers,
		Presets:             presets,
		SupportedDBVersions: []int{26},
	}
}
