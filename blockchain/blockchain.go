package blockchain

import (
	"crypto/ecdsa"
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

// ValidTransaction checks whether a transaction is valid, assuming the
// blockchain is valid.
func (bc *BlockChain) ValidTransaction(t *Transaction) bool {

	// Find the transaction input in the chain (by hash)
	var input *Transaction
	inputBlock := bc.Blocks[t.Input.BlockNumber]
	for _, transaction := range inputBlock.Transactions {
		if t.Input.Hash == HashSum(transaction) {
			input = transaction
		}
	}
	if input == nil {
		return false
	}

	// Check that output to sender in input is equal to outputs in t
	var inAmount uint64
	for _, output := range input.Outputs {
		if output.Recipient == t.Sender {
			inAmount += output.Amount
		}
	}
	var outAmount uint64
	for _, output := range t.Outputs {
		outAmount += output.Amount
	}
	if inAmount != outAmount {
		return false
	}

	// Verify signature of t
	hash := HashSum(t.TxBody)
	if !ecdsa.Verify(t.Sender.Key(), hash.Marshal(), t.Sig.R, t.Sig.S) {
		return false
	}

	// Test if identical transaction already exists in chain.
	endChain := uint32(len(bc.Blocks))
	for i := t.Input.BlockNumber; i < endChain; i++ {
		if bc.Blocks[i].ContainsTransaction(t) {
			return false
		}
	}

	return true
}

// ValidBlock checks whether a block is valid
func (bc *BlockChain) ValidBlock(b *Block) bool {
	// Check that block number is one greater than last block
	lastBlock := bc.Blocks[b.BlockNumber-1]
	if lastBlock.BlockNumber != b.BlockNumber-1 {
		return false
	}

	// Verify every Transaction in the block.
	for _, t := range b.Transactions {
		if !bc.ValidTransaction(t) {
			return false
		}
	}

	// Check that hash of last block is correct
	if HashSum(lastBlock) != b.LastBlock {
		return false
	}

	return true
}

// AppendBlock adds a block to the end of the block chain.
func (bc *BlockChain) AppendBlock(b *Block, miner Address) {
	b.BlockNumber = uint32(len(bc.Blocks))
	b.LastBlock = HashSum(bc.Blocks[b.BlockNumber-1])
	b.Miner = miner
	bc.Blocks = append(bc.Blocks, b)
}
