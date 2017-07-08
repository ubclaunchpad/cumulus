package blockchain

import (
	"crypto/sha256"
	"math/big"
)

const (
	// HashLen is the length in bytes of a hash.
	HashLen = 32
)

var (
	// NilHash represents a nil hash
	NilHash = BigIntToHash(big.NewInt(0))
)

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

// HashSum computes the SHA256 squared hash of a Marshaller.
func HashSum(m Marshaller) Hash {
	hash := sha256.Sum256(m.Marshal())
	return sha256.Sum256(hash[:])
}

// LessThan returns true if the receiver hash is less than the hash provided, and false otherwise
func (h Hash) LessThan(h2 Hash) bool {
	for i := HashLen - 1; i >= 0; i-- {
		if h[i] > h2[i] {
			return false
		} else if h[i] < h2[i] {
			return true
		}
	}
	return false
}

// HashToBigInt converts a hash to a big int pointer
func HashToBigInt(h Hash) *big.Int {
	for i, j := 0, HashLen-1; i < j; i, j = i+1, j-1 {
		h[i], h[j] = h[j], h[i]
	}
	return new(big.Int).SetBytes(h[:])
}

// BigIntToHash converts a big integer to a hash
func BigIntToHash(x *big.Int) Hash {
	bytes := x.Bytes()

	var result Hash
	for i := 0; i < HashLen; i++ {
		result[i] = 0
	}

	if len(bytes) > HashLen {
		return result
	}

	for i := 0; i < len(bytes); i++ {
		result[len(bytes)-1-i] = bytes[i]
	}
	return result
}
