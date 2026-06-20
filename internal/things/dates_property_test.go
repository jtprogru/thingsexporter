package things_test

import (
	"encoding/hex"
	"fmt"
	"testing"
	"time"

	"github.com/jtprogru/thingsexporter/internal/things"
	"github.com/stretchr/testify/require"
	"pgregory.net/rapid"
)

// PropCoreDataRoundTrip — CP-2: for any valid Unix timestamp in the range
// [Core Data epoch, 2100-01-01], parsing the result of CoreDataToISO yields the original seconds.
func PropCoreDataRoundTrip(t *rapid.T) {
	unixSec := rapid.Int64Range(978307200, 4102444800).Draw(t, "unixSec")
	coreSec := float64(unixSec - 978307200)
	got := things.CoreDataToISO(&coreSec)
	require.NotNil(t, got)
	parsed, err := time.Parse("2006-01-02T15:04:05.000000-07:00", *got)
	require.NoError(t, err)
	require.Equal(t, unixSec, parsed.Unix(), "round-trip mismatch")
	_, offset := parsed.Zone()
	require.Equal(t, 0, offset, "must be UTC")
}

func TestPropCoreDataRoundTrip(t *testing.T) {
	rapid.Check(t, PropCoreDataRoundTrip)
}

// PropPackedDateValid — CP-3: for any valid (year, month, day),
// PackedDateToISO(pack(y,m,d)) == sprintf("%04d-%02d-%02d", y, m, d).
func PropPackedDateValid(t *rapid.T) {
	year := rapid.IntRange(1970, 2100).Draw(t, "year")
	month := rapid.IntRange(1, 12).Draw(t, "month")
	day := rapid.IntRange(1, 31).Draw(t, "day")
	n := int64(year)<<16 | int64(month)<<12 | int64(day)<<7
	got := things.PackedDateToISO(&n)
	require.NotNil(t, got)
	want := fmt.Sprintf("%04d-%02d-%02d", year, month, day)
	require.Equal(t, want, *got)
}

func TestPropPackedDateValid(t *testing.T) {
	rapid.Check(t, PropPackedDateValid)
}

// PropEnumTotality — CP-5: known code → name; unknown code → nil.
func PropEnumTotality(t *rapid.T) {
	// Random code in extended range, then check based on which set it falls into.
	code := rapid.Int64Range(-100, 100).Draw(t, "code")
	want := map[int64]string{0: "todo", 1: "project", 2: "heading"}
	got := things.TaskTypeName(&code)
	if name, ok := want[code]; ok {
		require.NotNil(t, got)
		require.Equal(t, name, *got)
	} else {
		require.Nil(t, got)
	}
}

func TestPropEnumTotality(t *testing.T) {
	rapid.Check(t, PropEnumTotality)
}

// PropBlobEncoding — CP-6: drop ⇒ nil; empty ⇒ nil; otherwise ⇒ hex.EncodeToString.
func PropBlobEncoding(t *rapid.T) {
	b := rapid.SliceOfN(rapid.Byte(), 0, 256).Draw(t, "b")
	drop := rapid.Bool().Draw(t, "drop")
	got := things.EncodeBlob(b, drop)
	switch {
	case drop:
		require.Nil(t, got)
	case len(b) == 0:
		require.Nil(t, got)
	default:
		require.NotNil(t, got)
		require.NotNil(t, got.Hex)
		require.Equal(t, hex.EncodeToString(b), *got.Hex)
	}
}

func TestPropBlobEncoding(t *testing.T) {
	rapid.Check(t, PropBlobEncoding)
}
