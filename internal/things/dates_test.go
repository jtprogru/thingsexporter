package things_test

import (
	"math"
	"testing"

	"github.com/jtprogru/thingsexporter/internal/things"
	"github.com/stretchr/testify/require"
)

func ptrFloat64(v float64) *float64 { return &v }
func ptrInt64(v int64) *int64       { return &v }
func ptrString(v string) *string    { return &v }

func TestCoreDataToISO_table(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name string
		in   *float64
		want *string
	}{
		{"nil input", nil, nil},
		{"epoch zero", ptrFloat64(0), ptrString("2001-01-01T00:00:00.000000+00:00")},
		{"NaN", ptrFloat64(math.NaN()), nil},
		{"+Inf", ptrFloat64(math.Inf(1)), nil},
		{"-Inf", ptrFloat64(math.Inf(-1)), nil},
		// CORE_DATA_EPOCH (978307200) + 746541716 = 1724848916 = 2024-08-28T12:41:56 UTC
		{"known value", ptrFloat64(746541716.0), ptrString("2024-08-28T12:41:56.000000+00:00")},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := things.CoreDataToISO(tc.in)
			if tc.want == nil {
				require.Nil(t, got, "expected nil")
				return
			}
			require.NotNil(t, got, "expected non-nil")
			require.Equal(t, *tc.want, *got)
		})
	}
}

func packDate(year, month, day int) int64 {
	return int64(year)<<16 | int64(month)<<12 | int64(day)<<7
}

func TestPackedDateToISO_known(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name string
		in   *int64
		want *string
	}{
		{"nil", nil, nil},
		{"zero", ptrInt64(0), nil},
		{"2024-10-28", ptrInt64(packDate(2024, 10, 28)), ptrString("2024-10-28")},
		{"1970-01-01 boundary low", ptrInt64(packDate(1970, 1, 1)), ptrString("1970-01-01")},
		{"2100-12-31 boundary high", ptrInt64(packDate(2100, 12, 31)), ptrString("2100-12-31")},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := things.PackedDateToISO(tc.in)
			if tc.want == nil {
				require.Nil(t, got)
				return
			}
			require.NotNil(t, got)
			require.Equal(t, *tc.want, *got)
		})
	}
}

func TestPackedDateToISO_invalid(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name string
		in   *int64
	}{
		{"year 1969", ptrInt64(packDate(1969, 1, 1))},
		{"year 2101", ptrInt64(packDate(2101, 1, 1))},
		{"month 0", ptrInt64(packDate(2024, 0, 1))},
		{"month 13", ptrInt64(packDate(2024, 13, 1))},
		{"day 0", ptrInt64(packDate(2024, 1, 0))},
		{"day 32", ptrInt64(packDate(2024, 1, 32))},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := things.PackedDateToISO(tc.in)
			require.Nil(t, got, "invalid packed date should yield nil")
		})
	}
}
