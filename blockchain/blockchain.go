package blockchain

import (
	"encoding/gob"
	"io"
)

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

// BlockChain represents a linked list of blocks
type BlockChain struct {
	Blocks []*Block
	Head   Hash
}

// Encode writes the marshalled blockchain to the given io.Writer
func (bc *BlockChain) Encode(w io.Writer) {
	gob.NewEncoder(w).Encode(bc)
}

// Decode reads the marshalled blockchain from the given io.Reader
func (bc *BlockChain) Decode(r io.Reader) {
	gob.NewDecoder(r).Decode(bc)
}

// ValidTransaction checks whether a transaction is valid, assuming the blockchain is valid.
func (bc *BlockChain) ValidTransaction(t *Transaction) bool {
	// Find the transaction input (I) in the chain (by hash)
	// Check that output to sender in I is equal to outputs in T
	// Verify signature of T
	return false
}

// ValidBlock checks whether a block is valid
func (bc *BlockChain) ValidBlock(b *Block) bool {
	for _, t := range b.Transactions {
		if !bc.ValidTransaction(t) {
			return false
		}
	}
	// Check that block number is one greater than last block
	// Check that hash of last block is correct
	return false
}
