package blockchain

import (
	"encoding/gob"
	"io"
)

// BlockChain represents a linked list of blocks
type BlockChain struct {
	Blocks []*Block
	Head   Hash
}

// Len returns the length of the BlockChain when marshalled
func (bc *BlockChain) Len() int {
	l := 0
	for _, b := range bc.Blocks {
		l += b.Len()
	}
	return l + HashLen
}

// Marshal converts the BlockChain to a byte slice.
func (bc *BlockChain) Marshal() []byte {
	buf := make([]byte, 0, bc.Len())
	for _, b := range bc.Blocks {
		buf = append(buf, b.Marshal()...)
	}
	return append(buf, bc.Head.Marshal()...)
}

// Encode writes the marshalled blockchain to the given io.Writer
func (bc *BlockChain) Encode(w io.Writer) {
	gob.NewEncoder(w).Encode(bc)
}

// DecodeBlockChain reads the marshalled blockchain from the given io.Reader
func DecodeBlockChain(r io.Reader) *BlockChain {
	var bc BlockChain
	gob.NewDecoder(r).Decode(&bc)
	return &bc
}

// AppendBlock adds a block to the end of the block chain.
func (bc *BlockChain) AppendBlock(b *Block) {
	b.BlockNumber = uint32(len(bc.Blocks))
	b.LastBlock = HashSum(bc.Blocks[b.BlockNumber-1])
	bc.Blocks = append(bc.Blocks, b)
}

// GetInputTransaction returns the input Transaction to t. If the input does
// not exist, then GetInputTransaction returns nil.
func (bc *BlockChain) GetInputTransaction(t *Transaction) *Transaction {
	if t.Input.BlockNumber > uint32(len(bc.Blocks)) {
		return nil
	}
	b := bc.Blocks[t.Input.BlockNumber]
	if t.Input.Index > uint32(len(b.Transactions)) {
		return nil
	}
	return b.Transactions[t.Input.Index]
}

// ContainsTransaction returns true if the BlockChain contains the transaction
// in a block between start and stop as indexes.
func (bc *BlockChain) ContainsTransaction(t *Transaction, start, stop uint32) (bool, uint32, uint32) {
	for i := start; i < stop; i++ {
		if exists, j := bc.Blocks[i].ContainsTransaction(t); exists {
			return true, i, j
		}
	}
	return false, 0, 0
}
