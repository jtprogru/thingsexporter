package sqlite_test

import (
	"context"
	"testing"

	sqlitestore "github.com/jtprogru/thingsexporter/internal/store/sqlite"
	"github.com/stretchr/testify/require"
)

func TestOpen_readOnlyDSN(t *testing.T) {
	t.Parallel()
	path := buildFixture(t)
	db, err := sqlitestore.Open(path)
	require.NoError(t, err)
	defer func() { _ = db.Close() }()

	require.NoError(t, db.PingContext(context.Background()))

	// Попытка записи должна провалиться.
	_, err = db.ExecContext(context.Background(),
		`INSERT INTO "TMArea" ("uuid","title") VALUES ('X','Y')`)
	require.Error(t, err, "expected read-only refusal on write")
}

func TestOpen_emptyPath(t *testing.T) {
	t.Parallel()
	_, err := sqlitestore.Open("")
	require.Error(t, err)
}

func TestOpen_missingFile(t *testing.T) {
	t.Parallel()
	db, err := sqlitestore.Open("/nonexistent/path/db.sqlite")
	require.NoError(t, err) // sql.Open ленив
	defer func() { _ = db.Close() }()
	err = db.PingContext(context.Background())
	require.Error(t, err)
}
