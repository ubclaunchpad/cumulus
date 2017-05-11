package blockchain

// BlockHeader contains metadata about a block
import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/gob"
	"io"
)

// BlockHeader contains metadata about a block
type BlockHeader struct {
	blockNumber uint32
	lastBlock   Hash
	miner       Wallet
}

// Marshal converts a BlockHeader to a byte slice
func (bh *BlockHeader) Marshal() []byte {
	buf := []byte{}
	binary.LittleEndian.PutUint32(buf, bh.blockNumber)
	for _, b := range bh.lastBlock {
		buf = append(buf, b)
	}
	buf = append(buf, bh.miner.Marshal()...)
	return buf
}

// Block represents a block in the blockchain. Contains transactions and header metadata.
type Block struct {
	BlockHeader
	transactions []*Transaction
}

// Marshal converts a Block to a byte slice
func (b *Block) Marshal() []byte {
	buf := b.BlockHeader.Marshal()
	for _, t := range b.transactions {
		buf = append(buf, t.Marshal()...)
	}
	return buf
}

// Encode writes the marshalled block to the given io.Writer
func (b *Block) Encode(w io.Writer) {
	gob.NewEncoder(w).Encode(b)
}

// Decode reads the marshalled block from the given io.Reader
func (b *Block) Decode(r io.Reader) {
	gob.NewDecoder(r).Decode(b)
}

// Hash computes and returns the SHA256 hash of the block
func (b *Block) Hash() Hash {
	return sha256.Sum256(b.Marshal())
}
