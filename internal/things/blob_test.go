package things_test

import (
	"testing"

	"github.com/jtprogru/thingsexporter/internal/things"
	"github.com/stretchr/testify/require"
)

func TestEncodeBlob_table(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name    string
		in      []byte
		drop    bool
		wantNil bool
		wantHex string
	}{
		{"nil bytes, drop=false", nil, false, true, ""},
		{"empty bytes, drop=false", []byte{}, false, true, ""},
		{"two bytes, drop=false", []byte{0xde, 0xad}, false, false, "dead"},
		{"one byte, drop=true", []byte{0xde}, true, true, ""},
		{"nil bytes, drop=true", nil, true, true, ""},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := things.EncodeBlob(tc.in, tc.drop)
			if tc.wantNil {
				require.Nil(t, got)
				return
			}
			require.NotNil(t, got)
			require.NotNil(t, got.Hex)
			require.Equal(t, tc.wantHex, *got.Hex)
		})
	}
}
