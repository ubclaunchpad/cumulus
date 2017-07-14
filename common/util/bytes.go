package util

import "encoding/binary"

// AppendUint32 appends a uint32 to a slice, and returns the appended slice
func AppendUint32(s []byte, n uint32) []byte {
	temp := make([]byte, 4)
	binary.LittleEndian.PutUint32(temp, n)
	return append(s, temp...)
}

// AppendUint64 appends a uint64 to a slice, and returns the appended slice
func AppendUint64(s []byte, n uint64) []byte {
	temp := make([]byte, 8)
	binary.LittleEndian.PutUint64(temp, n)
	return append(s, temp...)
}
