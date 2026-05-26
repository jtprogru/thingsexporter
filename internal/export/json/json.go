// Package json реализует JSON-форматтер для Export.
package json

import (
	encjson "encoding/json"
	"io"
	"strings"

	"github.com/jtprogru/thingsexporter/internal/export"
	"github.com/jtprogru/thingsexporter/internal/things"
)

// Writer — реализация export.Writer для JSON.
type Writer struct{}

// Format возвращает имя формата.
func (Writer) Format() string { return "json" }

// Write сериализует Export в out как UTF-8 JSON.
// При opts.Indent > 0 — pretty-print с указанным количеством пробелов.
// При opts.Indent == 0 — компактный (через encjson.Encoder без SetIndent).
// HTML-escape отключён, чтобы кириллица/эмодзи не превращались в \uXXXX.
func (Writer) Write(out io.Writer, data things.Export, opts export.Options) error {
	enc := encjson.NewEncoder(out)
	enc.SetEscapeHTML(false)
	if opts.Indent > 0 {
		enc.SetIndent("", strings.Repeat(" ", opts.Indent))
	}
	return enc.Encode(data)
}
