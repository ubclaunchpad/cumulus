package blockchain

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/gob"
	"io"
)

// Hash represents a hash of a block or transaction
type Hash [32]byte

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

// Encode reads the marshalled block from the given io.Reader
func (b *Block) Decode(r io.Reader) {
	gob.NewDecoder(r).Decode(b)
}

// Hash computes and returns the SHA256 hash of the block
func (b *Block) Hash() Hash {
	return sha256.Sum256(b.Marshal())
}

// BlockChain represents a linked list of blocks
type BlockChain struct {
	blocks []*Block
	head   Hash
}

// Encode writes the marshalled blockchain to the given io.Writer
func (b *BlockChain) Encode(w io.Writer) {
	gob.NewEncoder(w).Encode(b)
}

// Encode reads the marshalled blockchain from the given io.Reader
func (b *BlockChain) Decode(r io.Reader) {
	gob.NewDecoder(r).Decode(b)
}
