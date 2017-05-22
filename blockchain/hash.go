package blockchain

import "crypto/sha256"

// HashLen is the length in bytes of a hash.
const HashLen = 32

// Hash represents a 256-bit hash of a block or transaction
type Hash [HashLen]byte

// Marshal converts a Hash to a slice.
func (h Hash) Marshal() []byte {
	buf := make([]byte, HashLen)
	for i, b := range h {
		buf[i] = b
	}
	return buf
}

// Marshaller is any type that can convert itself to a byte slice
type Marshaller interface {
	Marshal() []byte
}

// HashSum computes the SHA256 hash  of a Marshaller.
func HashSum(m Marshaller) Hash {
	return sha256.Sum256(m.Marshal())
}
