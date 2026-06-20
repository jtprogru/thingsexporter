// Package sqlite implements read-only access to the Things 3 SQLite database
// and auto-detection of its standard path on macOS.
package sqlite

import "path/filepath"

// DefaultMacOSDBPath is the path to the Things 3 database on macOS, relative to $HOME.
const DefaultMacOSDBPath = "Library/Group Containers/JLMPQHK86H.com.culturedcode.ThingsMac/Things Database.thingsdatabase/main.sqlite"

// Discover attempts to find the Things 3 database in its standard location.
// It returns the absolute path and true if the OS is darwin, home is non-empty,
// and the file exists (checked via statFn).
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
