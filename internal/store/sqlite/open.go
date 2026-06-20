package sqlite

import (
	"database/sql"
	"errors"
	"strconv"
	"strings"

	_ "modernc.org/sqlite"
)

// busyTimeoutMs is the SQLITE_BUSY wait timeout, in milliseconds. SQLite applies
// exponential backoff internally during the wait (see sqlite3BtreeBusyHandler).
const busyTimeoutMs = 5000

// Open opens the SQLite file strictly in read-only mode.
// The actual existence check happens on db.PingContext.
func Open(path string) (*sql.DB, error) {
	if path == "" {
		return nil, errors.New("empty db path")
	}
	dsn := "file:" + escapeURIPath(path) +
		"?mode=ro&_pragma=busy_timeout(" + strconv.Itoa(busyTimeoutMs) + ")"
	return sql.Open("sqlite", dsn)
}

// escapeURIPath percent-encodes characters that break a SQLite URI:
// `%` (first, to avoid double-escaping), `?` (start of query), and `#` (fragment).
// `/` is intentionally left as-is so that absolute paths work.
var uriPathEscaper = strings.NewReplacer("%", "%25", "?", "%3F", "#", "%23")

func escapeURIPath(p string) string { return uriPathEscaper.Replace(p) }
