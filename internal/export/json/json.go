// Package json implements a JSON formatter for Export.
package json

import (
	encjson "encoding/json"
	"io"
	"strings"

	"github.com/jtprogru/thingsexporter/internal/export"
	"github.com/jtprogru/thingsexporter/internal/things"
)

// Writer is the export.Writer implementation for JSON.
type Writer struct{}

// Format returns the format name.
func (Writer) Format() string { return "json" }

// Write serializes Export to out as UTF-8 JSON.
// When opts.Indent > 0 it pretty-prints with the given number of spaces.
// When opts.Indent == 0 it produces compact output (via encjson.Encoder without SetIndent).
// HTML escaping is disabled so that Cyrillic/emoji are not turned into \uXXXX.
func (Writer) Write(out io.Writer, data things.Export, opts export.Options) error {
	enc := encjson.NewEncoder(out)
	enc.SetEscapeHTML(false)
	if opts.Indent > 0 {
		enc.SetIndent("", strings.Repeat(" ", opts.Indent))
	}
	return enc.Encode(data)
}
