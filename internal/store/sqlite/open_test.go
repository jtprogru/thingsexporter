package sqlite_test

import (
	"context"
	"database/sql"
	"os"
	"path/filepath"
	"strings"
	"testing"

	_ "modernc.org/sqlite"

	sqlitestore "github.com/jtprogru/thingsexporter/internal/store/sqlite"
	"github.com/stretchr/testify/require"
)

// testEscapeURIPath дублирует логику escapeURIPath (внутренней для пакета),
// чтобы seed-DSN тоже работал на путях с `?`/`#`.
func testEscapeURIPath(p string) string {
	return strings.NewReplacer("%", "%25", "?", "%3F", "#", "%23").Replace(p)
}

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

// TestOpen_pathWithSpecialChars проверяет, что путь, содержащий `?` или `#`,
// корректно percent-кодируется в DSN и открывается тот же файл (а не разбитое
// на host?query или path#fragment).
func TestOpen_pathWithSpecialChars(t *testing.T) {
	t.Parallel()
	dir := filepath.Join(t.TempDir(), "weird?#dir")
	require.NoError(t, os.MkdirAll(dir, 0o755))
	path := filepath.Join(dir, "things.sqlite")

	// Создаём фикстуру с минимальной схемой напрямую (rwc),
	// чтобы read-only Open мог её прочитать.
	seed, err := sql.Open("sqlite", "file:"+testEscapeURIPath(path)+"?mode=rwc")
	require.NoError(t, err)
	_, err = seed.ExecContext(context.Background(),
		`CREATE TABLE marker ("k" TEXT PRIMARY KEY); INSERT INTO marker VALUES ('ok');`)
	require.NoError(t, err)
	require.NoError(t, seed.Close())

	db, err := sqlitestore.Open(path)
	require.NoError(t, err)
	defer func() { _ = db.Close() }()
	require.NoError(t, db.PingContext(context.Background()))

	var got string
	require.NoError(t, db.QueryRowContext(context.Background(),
		`SELECT "k" FROM marker`).Scan(&got))
	require.Equal(t, "ok", got)
}
