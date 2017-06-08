package blockchain

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"log"
)

const (
	// HashLen is the length in bytes of a hash.
	HashLen = 32
	// GreaterThan is the value to the CompareTo function returns if h1 is greater than h2
	GreaterThan int = 1
	// LessThan is the value to the CompareTo function returns if h1 is less than h2
	LessThan int = -1
	// EqualTo is the value to the CompareTo function returns if h1 is equal to h2
	EqualTo int = 0
	// MaxDifficultyHex is the maximum difficulty value represented as a hex string
	MaxDifficultyHex = "00000000FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF"
	// MaxHashHex is the maximum hash value represented as a hex string
	MaxHashHex = "FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF"
)

var (
	// MaxDifficulty is the maximum difficulty value
	MaxDifficulty = HexStringToHash(MaxDifficultyHex)
	// MaxHash is the maximum hash value
	MaxHash = HexStringToHash(MaxHashHex)
	// MinHash is the minimum hash value
	MinHash = HexStringToHash("0")
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
func CompareTo(h1 Hash, h2 Hash, comparator int) bool {
	for i := HashLen - 1; i >= 0; i-- {
		if h1[i] > h2[i] {
			return 1 == comparator
		} else if h1[i] < h2[i] {
			return -1 == comparator
		}
	}
	return 0 == comparator
}

// ReverseHash reverses the endianess of a hash
func ReverseHash(h Hash) Hash {
	var reverse Hash
	for i, j := 0, HashLen-1; i < j; i, j = i+1, j-1 {
		reverse[i], reverse[j] = h[j], h[i]
	}
	return reverse
}

// HexStringToHash converts a big endian hex string to a hash
func HexStringToHash(s string) Hash {
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
		log.Fatal(e)
	}

	var hash Hash
	for i := 0; i < HashLen; i++ {
		hash[i] = bytes[HashLen-1-i]
	}

	return hash
}

// HashToCompact converts a hash to a compact number
func HashToCompact(h Hash) uint32 {
	// Return 0 if hash is equal to 0
	if CompareTo(h, MinHash, EqualTo) {
		return 0
	}

	var compact []byte
	compact = make([]byte, 4)

	// Find MSB
	var msb int
	for msb = HashLen - 1; msb >= 0 && h[msb] == 0; msb-- {
	}

	// Set MSB of compact number to the size of the number
	size := msb + 1

	// Prepend 0 if msb is less than 0x7f
	if h[msb] > 0x7f {
		size++
		compact[1] = 0
		compact[2] = h[msb]
		if msb-1 < 0 {
			compact[3] = 0
		} else {
			compact[3] = h[msb-1]
		}
	} else {
		compact[1] = h[msb]

		if msb-1 < 0 {
			compact[2] = 0
		} else {
			compact[2] = h[msb-1]
		}

		if msb-2 < 0 {
			compact[3] = 0
		} else {
			compact[3] = h[msb-2]
		}
	}

	compact[0] = byte(size)

	return binary.BigEndian.Uint32(compact)
}

// CompactToHash converts a compact number to a hash
func CompactToHash(c uint32) Hash {
	if c == 0 {
		return MinHash
	}

	maxCompactNumber := HashToCompact(MaxDifficulty)
	if c > maxCompactNumber {
		// TODO: Handle Error
	}

	buf := make([]byte, 4)
	binary.BigEndian.PutUint32(buf, c)

	size := int(buf[0])
	if size > HashLen {
		// TODO: Handle Error
	}

	var h Hash
	for i, j := size-1, 1; j <= 3; i, j = i-1, j+1 {
		if j <= size {
			h[i] = buf[j]
		}
	}

	return h
}
