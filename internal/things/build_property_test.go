package things_test

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/jtprogru/thingsexporter/internal/things"
	"github.com/stretchr/testify/require"
	"pgregory.net/rapid"
)

func genPtrInt64() *rapid.Generator[*int64] {
	return rapid.Custom(func(t *rapid.T) *int64 {
		if rapid.Bool().Draw(t, "isNil") {
			return nil
		}
		v := rapid.Int64Range(-10000, 10000).Draw(t, "i")
		return &v
	})
}

func genTrashed() *rapid.Generator[*int64] {
	return rapid.Custom(func(t *rapid.T) *int64 {
		switch rapid.IntRange(0, 2).Draw(t, "t") {
		case 0:
			return nil
		case 1:
			zero := int64(0)
			return &zero
		default:
			one := int64(1)
			return &one
		}
	})
}

func genRawData() *rapid.Generator[things.RawData] {
	return rapid.Custom(func(t *rapid.T) things.RawData {
		nAreas := rapid.IntRange(0, 4).Draw(t, "nAreas")
		nTags := rapid.IntRange(0, 4).Draw(t, "nTags")
		nTasks := rapid.IntRange(0, 8).Draw(t, "nTasks")

		areas := make([]things.RawArea, 0, nAreas)
		for i := 0; i < nAreas; i++ {
			areas = append(areas, things.RawArea{
				UUID:  fmt.Sprintf("A-%d", i),
				Title: ptrString(fmt.Sprintf("area-%d", i)),
				Index: genPtrInt64().Draw(t, fmt.Sprintf("a-idx-%d", i)),
			})
		}

		tags := make([]things.RawTag, 0, nTags)
		for i := 0; i < nTags; i++ {
			tags = append(tags, things.RawTag{
				UUID:  fmt.Sprintf("T-%d", i),
				Title: ptrString(fmt.Sprintf("tag-%d", i)),
			})
		}

		// task types: 0 todo, 1 project, 2 heading
		tasks := make([]things.RawTask, 0, nTasks)
		for i := 0; i < nTasks; i++ {
			ttype := int64(rapid.IntRange(0, 2).Draw(t, fmt.Sprintf("t-type-%d", i)))
			tasks = append(tasks, things.RawTask{
				UUID:    fmt.Sprintf("task-%d", i),
				Title:   ptrString(fmt.Sprintf("title-%d", i)),
				Type:    &ttype,
				Status:  ptrInt64(0),
				Trashed: genTrashed().Draw(t, fmt.Sprintf("trash-%d", i)),
				Index:   genPtrInt64().Draw(t, fmt.Sprintf("t-idx-%d", i)),
			})
		}

		// task-tag pairs (each task — a random subset of tags)
		var pairs []things.TaskTagLink
		for i := 0; i < nTasks; i++ {
			for j := 0; j < nTags; j++ {
				if rapid.Bool().Draw(t, fmt.Sprintf("p-%d-%d", i, j)) {
					pairs = append(pairs, things.TaskTagLink{
						Task: fmt.Sprintf("task-%d", i),
						Tag:  fmt.Sprintf("T-%d", j),
					})
				}
			}
		}

		return things.RawData{
			Areas:        areas,
			Tags:         tags,
			Tasks:        tasks,
			TaskTagPairs: pairs,
		}
	})
}

// PropCountsMatchCollections — CP-7.
func PropCountsMatchCollections(t *rapid.T) {
	raw := genRawData().Draw(t, "raw")
	e := things.Build(raw, things.BuildOptions{})
	require.NotNil(t, e.Meta.Counts.Areas)
	require.Equal(t, len(e.Areas), *e.Meta.Counts.Areas)
	require.NotNil(t, e.Meta.Counts.Tags)
	require.Equal(t, len(e.Tags), *e.Meta.Counts.Tags)
	require.NotNil(t, e.Meta.Counts.Tasks)
	require.Equal(t, len(e.Tasks), *e.Meta.Counts.Tasks)
}

func TestPropCountsMatchCollections(t *testing.T) {
	rapid.Check(t, PropCountsMatchCollections)
}

// PropTagsEnrichment — CP-8.
func PropTagsEnrichment(t *rapid.T) {
	raw := genRawData().Draw(t, "raw")
	e := things.Build(raw, things.BuildOptions{})
	// for each task, the number of tag-refs == the number of pairs with its UUID
	for _, task := range e.Tasks {
		expected := 0
		for _, p := range raw.TaskTagPairs {
			if p.Task == task.UUID {
				expected++
			}
		}
		require.Len(t, task.Tags, expected, "task %s tags mismatch", task.UUID)
	}
}

func TestPropTagsEnrichment(t *testing.T) {
	rapid.Check(t, PropTagsEnrichment)
}

// PropHierarchyExcludesTrashed — CP-9.
func PropHierarchyExcludesTrashed(t *rapid.T) {
	raw := genRawData().Draw(t, "raw")
	e := things.Build(raw, things.BuildOptions{})
	require.NotNil(t, e.Hierarchy)
	trashed := map[string]bool{}
	for _, rt := range raw.Tasks {
		if rt.Trashed != nil && *rt.Trashed == 1 {
			trashed[rt.UUID] = true
		}
	}
	for _, a := range e.Hierarchy.Areas {
		for _, it := range a.Items {
			require.False(t, trashed[it.UUID], "trashed task %s leaked into hierarchy area", it.UUID)
		}
	}
	for _, it := range e.Hierarchy.InboxOrOrphanTasks {
		require.False(t, trashed[it.UUID], "trashed task %s leaked into inbox", it.UUID)
	}
}

func TestPropHierarchyExcludesTrashed(t *testing.T) {
	rapid.Check(t, PropHierarchyExcludesTrashed)
}

// PropHierarchyOrdering — CP-10.
func PropHierarchyOrdering(t *rapid.T) {
	raw := genRawData().Draw(t, "raw")
	e := things.Build(raw, things.BuildOptions{})
	require.NotNil(t, e.Hierarchy)
	// Areas: non-nil indexes are sorted ASC, nil — at the end
	var lastIdx *int64
	seenNil := false
	for _, a := range e.Hierarchy.Areas {
		if a.Index == nil {
			seenNil = true
			continue
		}
		require.False(t, seenNil, "non-nil index after nil - wrong ordering")
		if lastIdx != nil {
			require.LessOrEqual(t, *lastIdx, *a.Index)
		}
		lastIdx = a.Index
	}
}

func TestPropHierarchyOrdering(t *testing.T) {
	rapid.Check(t, PropHierarchyOrdering)
}

// PropNoBlobsPropagation — CP-18.
func PropNoBlobsPropagation(t *rapid.T) {
	raw := genRawData().Draw(t, "raw")
	// Fill all areas and tasks with BLOB bytes.
	for i := range raw.Areas {
		raw.Areas[i].CachedTags = []byte{0xab, 0xcd}
		raw.Areas[i].Experimental = []byte{0x01}
	}
	for i := range raw.Tasks {
		raw.Tasks[i].CachedTags = []byte{0xff}
		raw.Tasks[i].Experimental = []byte{0x02}
		raw.Tasks[i].Repeater = []byte{0x03}
		raw.Tasks[i].Rt1RecurrenceRule = []byte{0x04}
	}
	e := things.Build(raw, things.BuildOptions{NoBlobs: true})
	for _, a := range e.Areas {
		require.Nil(t, a.CachedTags)
		require.Nil(t, a.Experimental)
	}
	for _, tk := range e.Tasks {
		require.Nil(t, tk.CachedTags)
		require.Nil(t, tk.Experimental)
		require.Nil(t, tk.Repeater)
		require.Nil(t, tk.Rt1RecurrenceRule)
	}
}

func TestPropNoBlobsPropagation(t *testing.T) {
	rapid.Check(t, PropNoBlobsPropagation)
}

// PropSchemaPresent — CP-20.
func PropSchemaPresent(t *rapid.T) {
	raw := genRawData().Draw(t, "raw")
	e := things.Build(raw, things.BuildOptions{})
	b, err := json.Marshal(e)
	require.NoError(t, err)
	require.Contains(t, string(b), `"schema":"thingsexporter/v1"`)
}

func TestPropSchemaPresent(t *testing.T) {
	rapid.Check(t, PropSchemaPresent)
}
