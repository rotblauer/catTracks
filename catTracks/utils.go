package catTracks

import "encoding/binary"

// i64tob returns an 8-byte big endian representation of v.
func i64tob(v int64) []byte {
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, uint64(v))
	return b
}
