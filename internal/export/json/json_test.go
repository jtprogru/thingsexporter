package json_test

import (
	"bytes"
	encjson "encoding/json"
	"strings"
	"testing"

	"github.com/jtprogru/thingsexporter/internal/export"
	jsonwriter "github.com/jtprogru/thingsexporter/internal/export/json"
	"github.com/jtprogru/thingsexporter/internal/things"
	"github.com/stretchr/testify/require"
)

func sampleExport() things.Export {
	cyr := "Тест"
	return things.Export{
		Schema: things.SchemaVersion,
		Meta: things.Meta{
			Source:     "fixture.sqlite",
			ExportedAt: "2026-05-25T00:00:00.000000+00:00",
			Counts:     things.Counts{},
		},
		Tasks: []things.Task{
			{UUID: "u1", Title: &cyr},
		},
	}
}

func TestJsonWriter_format_returns_json(t *testing.T) {
	t.Parallel()
	require.Equal(t, "json", jsonwriter.Writer{}.Format())
}

func TestJsonWriter_compact(t *testing.T) {
	t.Parallel()
	var buf bytes.Buffer
	require.NoError(t, jsonwriter.Writer{}.Write(&buf, sampleExport(), export.Options{Indent: 0}))
	s := buf.String()
	// единственный '\n' — это терминатор Encoder.Encode
	require.Equal(t, 1, strings.Count(s, "\n"), "compact JSON must have only trailing newline")
	require.False(t, strings.Contains(s, "  "), "compact JSON must not have double spaces")
}

func TestJsonWriter_indent_two(t *testing.T) {
	t.Parallel()
	var buf bytes.Buffer
	require.NoError(t, jsonwriter.Writer{}.Write(&buf, sampleExport(), export.Options{Indent: 2}))
	s := buf.String()
	require.Contains(t, s, "\n  \"schema\":", "top-level keys must be indented 2 spaces")

	// Парсится обратно
	var back map[string]any
	require.NoError(t, encjson.Unmarshal(buf.Bytes(), &back))
	require.Equal(t, things.SchemaVersion, back["schema"])
}

func TestJsonWriter_noASCIIEscape(t *testing.T) {
	t.Parallel()
	var buf bytes.Buffer
	require.NoError(t, jsonwriter.Writer{}.Write(&buf, sampleExport(), export.Options{Indent: 0}))
	require.Contains(t, buf.String(), "Тест", "cyrillic must remain UTF-8, not \\u escaped")
}
