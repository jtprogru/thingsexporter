// Package version exposes build-time identification for thingsexporter.
// Values are injected via -ldflags from goreleaser; defaults are used for
// local builds without ldflag overrides.
package version

var (
	Version = "dev"
	Commit  = ""
	Date    = ""
	BuiltBy = ""
)
