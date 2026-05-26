// Package sqlite реализует чтение Things 3 SQLite БД в режиме read-only
// и авто-определение её стандартного пути на macOS.
package sqlite

import "path/filepath"

// DefaultMacOSDBPath — относительный (от $HOME) путь к Things 3 БД на macOS.
const DefaultMacOSDBPath = "Library/Group Containers/JLMPQHK86H.com.culturedcode.ThingsMac/Things Database.thingsdatabase/main.sqlite"

// Discover пытается найти Things 3 БД в стандартном месте.
// Возвращает абсолютный путь и true, если ОС — darwin, home непуст и файл существует
// (проверка статусом через statFn).
func Discover(home, goos string, statFn func(string) error) (string, bool) {
	if home == "" || goos != "darwin" {
		return "", false
	}
	p := filepath.Join(home, DefaultMacOSDBPath)
	if err := statFn(p); err != nil {
		return "", false
	}
	return p, true
}
