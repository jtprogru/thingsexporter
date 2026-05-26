package sqlite

import (
	"database/sql"
	"errors"
	"fmt"

	_ "modernc.org/sqlite"
)

// Open открывает SQLite-файл строго в режиме read-only.
// Реальная проверка существования происходит на db.PingContext.
func Open(path string) (*sql.DB, error) {
	if path == "" {
		return nil, errors.New("empty db path")
	}
	dsn := fmt.Sprintf("file:%s?mode=ro", path)
	return sql.Open("sqlite", dsn)
}
