package things_test

import (
	"testing"

	"github.com/jtprogru/thingsexporter/internal/things"
	"github.com/stretchr/testify/require"
)

func TestTaskTypeName_known(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name string
		in   *int64
		want *string
	}{
		{"nil", nil, nil},
		{"0 todo", ptrInt64(0), ptrString("todo")},
		{"1 project", ptrInt64(1), ptrString("project")},
		{"2 heading", ptrInt64(2), ptrString("heading")},
		{"unknown 99", ptrInt64(99), nil},
		{"negative", ptrInt64(-1), nil},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := things.TaskTypeName(tc.in)
			if tc.want == nil {
				require.Nil(t, got)
				return
			}
			require.NotNil(t, got)
			require.Equal(t, *tc.want, *got)
		})
	}
}

func TestTaskStatusName_known(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name string
		in   *int64
		want *string
	}{
		{"nil", nil, nil},
		{"0 open", ptrInt64(0), ptrString("open")},
		{"2 canceled", ptrInt64(2), ptrString("canceled")},
		{"3 completed", ptrInt64(3), ptrString("completed")},
		{"unknown 1", ptrInt64(1), nil},
		{"unknown 4", ptrInt64(4), nil},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := things.TaskStatusName(tc.in)
			if tc.want == nil {
				require.Nil(t, got)
				return
			}
			require.NotNil(t, got)
			require.Equal(t, *tc.want, *got)
		})
	}
}

func TestChecklistStatusName_known(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name string
		in   *int64
		want *string
	}{
		{"nil", nil, nil},
		{"0 open", ptrInt64(0), ptrString("open")},
		{"3 completed", ptrInt64(3), ptrString("completed")},
		{"unknown 2", ptrInt64(2), nil},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := things.ChecklistStatusName(tc.in)
			if tc.want == nil {
				require.Nil(t, got)
				return
			}
			require.NotNil(t, got)
			require.Equal(t, *tc.want, *got)
		})
	}
}
