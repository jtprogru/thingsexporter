package sqlite_test

import (
	"context"
	"testing"

	sqlitestore "github.com/jtprogru/thingsexporter/internal/store/sqlite"
	"github.com/stretchr/testify/require"
)

func mustOpenFixture(t testing.TB) *sqlitestore.Repository {
	t.Helper()
	path := buildFixture(t)
	db, err := sqlitestore.Open(path)
	require.NoError(t, err)
	repo := sqlitestore.NewRepository(db)
	t.Cleanup(func() { _ = repo.Close() })
	return repo
}

func TestRepositoryReadAll_fixture(t *testing.T) {
	t.Parallel()
	repo := mustOpenFixture(t)
	raw, err := repo.ReadAll(context.Background())
	require.NoError(t, err)
	require.Len(t, raw.Areas, 2)
	require.Len(t, raw.Tags, 3)
	require.Len(t, raw.Tasks, 5)
	require.Len(t, raw.Checklist, 2)
	require.Len(t, raw.Contacts, 1)
	require.Len(t, raw.Tombstones, 1)
	require.Len(t, raw.TaskTagPairs, 3)
	require.Len(t, raw.AreaTagPairs, 1)
	require.NotEmpty(t, raw.MetaRows)

	// Спот-проверки полей
	var task1 *struct{ trashed *int64 }
	for _, t := range raw.Tasks {
		if t.UUID == "task-trashed" {
			v := *t.Trashed
			task1 = &struct{ trashed *int64 }{trashed: &v}
		}
	}
	require.NotNil(t, task1)
	require.NotNil(t, task1.trashed)
	require.Equal(t, int64(1), *task1.trashed)
}

func TestRepositoryReadCounts_fixture(t *testing.T) {
	t.Parallel()
	repo := mustOpenFixture(t)
	c, err := repo.ReadCounts(context.Background())
	require.NoError(t, err)
	require.NotNil(t, c.Areas)
	require.Equal(t, 2, *c.Areas)
	require.NotNil(t, c.Tasks)
	require.Equal(t, 5, *c.Tasks)
	require.NotNil(t, c.Tags)
	require.Equal(t, 3, *c.Tags)
	require.NotNil(t, c.ChecklistItems)
	require.Equal(t, 2, *c.ChecklistItems)
	require.NotNil(t, c.TaskTagLinks)
	require.Equal(t, 3, *c.TaskTagLinks)
	require.NotNil(t, c.AreaTagLinks)
	require.Equal(t, 1, *c.AreaTagLinks)
}

func TestRepositoryDatabaseVersion_meta(t *testing.T) {
	t.Parallel()
	repo := mustOpenFixture(t)
	v, err := repo.DatabaseVersion(context.Background())
	require.NoError(t, err)
	require.NotNil(t, v)
	require.Equal(t, 26, *v)
}
