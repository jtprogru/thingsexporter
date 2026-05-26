package things

import "encoding/hex"

// BlobValue представляет BLOB-поле в JSON-выводе: {"__blob_hex__":"<lowercase-hex>"}.
// Указатель nil в коде сериализуется как JSON null.
type BlobValue struct {
	Hex *string `json:"__blob_hex__,omitempty"`
}

// EncodeBlob кодирует сырые байты BLOB-колонки в *BlobValue.
// Возвращает nil если b пуст, либо если drop == true.
func EncodeBlob(b []byte, drop bool) *BlobValue {
	if drop || len(b) == 0 {
		return nil
	}
	h := hex.EncodeToString(b)
	return &BlobValue{Hex: &h}
}
