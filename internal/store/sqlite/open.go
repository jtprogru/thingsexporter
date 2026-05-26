package sqlite

import (
	"database/sql"
	"errors"
	"strconv"
	"strings"

	_ "modernc.org/sqlite"
)

// busyTimeoutMs — таймаут ожидания SQLITE_BUSY, в миллисекундах. SQLite сам
// делает экспоненциальный backoff внутри ожидания (см. sqlite3BtreeBusyHandler).
const busyTimeoutMs = 5000

// Open открывает SQLite-файл строго в режиме read-only.
// Реальная проверка существования происходит на db.PingContext.
func Open(path string) (*sql.DB, error) {
	if path == "" {
		return nil, errors.New("empty db path")
	}
	dsn := "file:" + escapeURIPath(path) +
		"?mode=ro&_pragma=busy_timeout(" + strconv.Itoa(busyTimeoutMs) + ")"
	return sql.Open("sqlite", dsn)
}

// escapeURIPath процентно-кодирует символы, ломающие SQLite URI:
// `%` (первым, чтобы не двойной escape), `?` (начало query) и `#` (fragment).
// `/` намеренно оставлен, чтобы абсолютные пути работали.
var uriPathEscaper = strings.NewReplacer("%", "%25", "?", "%3F", "#", "%23")

func escapeURIPath(p string) string { return uriPathEscaper.Replace(p) }
