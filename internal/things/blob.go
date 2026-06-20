package things

import "encoding/hex"

// BlobValue represents a BLOB field in the JSON output: {"__blob_hex__":"<lowercase-hex>"}.
// A nil pointer in code is serialized as JSON null.
type BlobValue struct {
	Hex *string `json:"__blob_hex__,omitempty"`
}

// EncodeBlob encodes the raw bytes of a BLOB column into a *BlobValue.
// Returns nil if b is empty, or if drop == true.
func EncodeBlob(b []byte, drop bool) *BlobValue {
	if drop || len(b) == 0 {
		return nil
	}
	h := hex.EncodeToString(b)
	return &BlobValue{Hex: &h}
}
