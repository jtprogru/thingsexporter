package things

import (
	"fmt"
	"math"
	"time"
)

// coreDataEpochUnix — секунды от Unix epoch до Core Data epoch (2001-01-01 UTC).
const coreDataEpochUnix = int64(978307200)

// CoreDataToISO конвертирует Core Data timestamp (вещественное число секунд от 2001-01-01 UTC)
// в строку ISO 8601 UTC с микросекундной точностью.
// Возвращает nil для nil-входа, NaN и Inf.
func CoreDataToISO(v *float64) *string {
	if v == nil {
		return nil
	}
	f := *v
	if math.IsNaN(f) || math.IsInf(f, 0) {
		return nil
	}
	sec := math.Trunc(f)
	frac := f - sec
	if frac < 0 {
		sec -= 1
		frac += 1
	}
	t := time.Unix(coreDataEpochUnix+int64(sec), int64(math.Round(frac*1e9))).UTC()
	s := t.Format("2006-01-02T15:04:05.000000-07:00")
	return &s
}

// PackedDateToISO декодирует packed-дату Things 3 (битовая раскладка
// (year<<16) | (month<<12) | (day<<7)) в строку "YYYY-MM-DD".
// Возвращает nil для nil, 0 и любых значений, у которых декодированные
// year/month/day выходят за валидные диапазоны.
func PackedDateToISO(v *int64) *string {
	if v == nil || *v == 0 {
		return nil
	}
	n := *v
	year := int((n >> 16) & 0xFFFF)
	month := int((n >> 12) & 0x0F)
	day := int((n >> 7) & 0x1F)
	if year < 1970 || year > 2100 || month < 1 || month > 12 || day < 1 || day > 31 {
		return nil
	}
	s := fmt.Sprintf("%04d-%02d-%02d", year, month, day)
	return &s
}
