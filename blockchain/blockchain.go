package blockchain

import (
	"encoding/gob"
	"errors"
	"io"
)

// BlockChain represents a linked list of blocks
type BlockChain struct {
	Blocks []*Block
	Head   Hash
}

// Len returns the length of the BlockChain when marshalled
func (bc *BlockChain) Len() int {
	return len(bc.Marshal())
}

// Marshal converts the BlockChain to a byte slice.
func (bc *BlockChain) Marshal() []byte {
	var buf []byte
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
	bc.Head = HashSum(b)
}

// LastBlock returns a pointer to the last block in the given blockchain, or nil
// if the blockchain is empty.
func (bc *BlockChain) LastBlock() *Block {
	if len(bc.Blocks) == 0 {
		return nil
	}
	return bc.Blocks[len(bc.Blocks)-1]
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

// CopyBlockByLastBlockHash returns a copy of the block in the local chain that
// comes directly after the block with the given hash. Returns error if no such
// block is found.
func (bc *BlockChain) CopyBlockByLastBlockHash(hash Hash) (*Block, error) {
	// Find the block with the given hash
	for _, block := range bc.Blocks {
		if block.LastBlock == hash {
			newBlock := *block
			newBlock.Transactions = make([]*Transaction, len(block.Transactions))
			copy(newBlock.Transactions, block.Transactions)
			return &newBlock, nil
		}
	}
	return nil, errors.New("No such block")
}

// RollBack removes the last block from the blockchain. Returns the block that
// was removed from the end of the chain, or nil if the blockchain is empty.
func (bc *BlockChain) RollBack() *Block {
	if len(bc.Blocks) == 0 {
		return nil
	}
	prevHead := bc.LastBlock()
	bc.Blocks = bc.Blocks[:len(bc.Blocks)-1]
	if len(bc.Blocks) == 0 {
		bc.Head = NilHash
	} else {
		bc.Head = HashSum(bc.LastBlock())
	}
	return prevHead
}
