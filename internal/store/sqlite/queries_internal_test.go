package sqlite

import (
	"regexp"
	"testing"
)

// TestQueriesQuoteReservedWords (CP-17): every occurrence of a SQL reserved word
// used as a column identifier in our SELECTs must be wrapped in `"`. Behavioral
// coverage is provided by TestRepositoryReadAll_fixture; this test is a static
// guard against regressions when editing the SQL.
func TestQueriesQuoteReservedWords(t *testing.T) {
	t.Parallel()
	queries := map[string]string{
		"selectAreasSQL":               selectAreasSQL,
		"selectTagsSQL":                selectTagsSQL,
		"selectTasksSQL":               selectTasksSQL,
		"selectChecklistSQL":           selectChecklistSQL,
		"selectContactsSQL":            selectContactsSQL,
		"selectTombstonesSQL":          selectTombstonesSQL,
		"selectTaskTagsSQL":            selectTaskTagsSQL,
		"selectAreaTagsSQL":            selectAreaTagsSQL,
		"selectMetaRowsSQL":            selectMetaRowsSQL,
		"selectMetaDatabaseVersionSQL": selectMetaDatabaseVersionSQL,
	}
	reserved := []string{"index", "type", "status", "start"}
	for _, w := range reserved {
		re := regexp.MustCompile(`(?i)\b` + w + `\b`)
		for name, sql := range queries {
			for _, m := range re.FindAllStringIndex(sql, -1) {
				if !quoted(sql, m[0], m[1]) {
					left := m[0] - 8
					if left < 0 {
						left = 0
					}
					right := m[1] + 8
					if right > len(sql) {
						right = len(sql)
					}
					t.Errorf("%s: reserved word %q not double-quoted near %q",
						name, w, sql[left:right])
				}
			}
		}
	}
}

func quoted(s string, left, right int) bool {
	if left == 0 || right >= len(s) {
		return false
	}
	return s[left-1] == '"' && s[right] == '"'
}
