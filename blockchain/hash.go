package blockchain

import (
	"crypto/sha256"
	"encoding/hex"

	log "github.com/Sirupsen/logrus"
)

const (
	// HashLen is the length in bytes of a hash.
	HashLen = 32
	// MaxDifficultyHex is the maximum difficulty value represented as a hex string
	MaxDifficultyHex = "0x000000FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF"
	// MaxHashHex is the maximum hash value represented as a hex string
	MaxHashHex = "0xFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF"
)

// CompareTo comparator constants
const (
	// GreaterThan is the value to the CompareTo function returns if h1 is greater than h2
	GreaterThan int = 1
	// LessThan is the value to the CompareTo function returns if h1 is less than h2
	LessThan int = -1
	// EqualTo is the value to the CompareTo function returns if h1 is equal to h2
	EqualTo int = 0
)

var (
	// MaxDifficulty is the maximum difficulty value
	MaxDifficulty = new(Hash).ParseString(MaxDifficultyHex)
	// MaxHash is the maximum hash value
	MaxHash = new(Hash).ParseString(MaxHashHex)
	// MinHash is the minimum hash value
	MinHash = new(Hash).ParseString("0x00")
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

// HashSum computes the SHA256 hash  of a Marshaller.
func HashSum(m Marshaller) Hash {
	return sha256.Sum256(m.Marshal())
}

// CompareTo compares two hashes, it returns true if the operation of the first hash on the second hash specified by the comparator is true, and false otherwise
func (h Hash) CompareTo(h2 Hash, comparator int) bool {
	for i := HashLen - 1; i >= 0; i-- {
		if h[i] > h2[i] {
			return GreaterThan == comparator
		} else if h[i] < h2[i] {
			return LessThan == comparator
		}
	}
	return EqualTo == comparator
}

// ParseString sets the hash to the the value represented by a hexadecimal string, and returns the hash
func (h *Hash) ParseString(s string) *Hash {
	if len(s) == 0 || len(s) == 1 || s[:2] != "0x" {
		log.Error("Invalid hexadecimal string, must begin with 0x")
		return nil
	}

	s = s[2:]
	if len(s) > HashLen*2 {
		log.Error("Hexadeximal string is too large")
		return nil
	}

	if len(s)%2 != 0 {
		s = "0" + s
	}

	if len(s)/2 < HashLen {
		pre := ""
		for i := 0; i < HashLen-len(s)/2; i++ {
			pre += "00"
		}
		s = pre + s
	}

	var bytes, e = hex.DecodeString(s)

	if e != nil {
		log.Error(e)
		return nil
	}

	for i := 0; i < HashLen; i++ {
		h[i] = bytes[HashLen-1-i]
	}

	return h
}

// HexString returns the hexadecimal representation of the hash
func (h Hash) HexString() string {
	var hashBigEndian Hash
	for i := 0; i < HashLen; i++ {
		hashBigEndian[i] = h[HashLen-1-i]
	}
	return "0x" + hex.EncodeToString(hashBigEndian[:])
}
