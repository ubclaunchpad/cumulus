package blockchain

// BlockHeader contains metadata about a block
import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/gob"
	"fmt"
	"io"
)

const (
	// BlockSize is the maximum size of a block in bytes when marshaled (about 250K).
	BlockSize = 1 << 18
	// BlockHeaderLen is the length in bytes of a block header.
	BlockHeaderLen = 32/8 + HashLen + AddrLen
)

// BlockHeader contains metadata about a block
type BlockHeader struct {
	BlockNumber uint32
	LastBlock   Hash
	Miner       Address
}

// Marshal converts a BlockHeader to a byte slice
func (bh *BlockHeader) Marshal() []byte {
	buf := make([]byte, 4, BlockHeaderLen)
	binary.LittleEndian.PutUint32(buf, bh.BlockNumber)
	for _, b := range bh.LastBlock {
		buf = append(buf, b)
	}
	buf = append(buf, bh.Miner.Marshal()...)
	return buf
}

// Block represents a block in the blockchain. Contains transactions and header metadata.
type Block struct {
	BlockHeader
	Transactions []*Transaction
}

// Len returns the length in bytes of the Block.
func (b *Block) Len() int {
	l := BlockHeaderLen
	for _, t := range b.Transactions {
		l += t.Len()
	}
	return l
}

// Marshal converts a Block to a byte slice.
func (b *Block) Marshal() []byte {
	buf := make([]byte, 0, b.Len())
	buf = append(buf, b.BlockHeader.Marshal()...)
	for _, t := range b.Transactions {
		buf = append(buf, t.Marshal()...)
	}
	return buf
}

// Encode writes the marshalled block to the given io.Writer
func (b *Block) Encode(w io.Writer) {
	err := gob.NewEncoder(w).Encode(b)
	if err != nil {
		fmt.Println(err.Error())
	}
}

// Decode reads the marshalled block from the given io.Reader and populates b
func (b *Block) Decode(r io.Reader) {
	gob.NewDecoder(r).Decode(b)
}

// Hash computes and returns the SHA256 hash of the block
func (b *Block) Hash() Hash {
	return sha256.Sum256(b.Marshal())
}
