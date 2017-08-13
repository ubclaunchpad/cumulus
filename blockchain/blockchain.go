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
}

// GetInputTransaction returns the input Transaction referenced by TxHashPointer.
// If the Transaction does not exist, then GetInputTransaction returns nil.
func (bc *BlockChain) GetInputTransaction(t *TxHashPointer) *Transaction {
	if t.BlockNumber > uint32(len(bc.Blocks)) {
		return nil
	}
	b := bc.Blocks[t.BlockNumber]
	if t.Index > uint32(len(b.Transactions)) {
		return nil
	}
	return b.Transactions[t.Index]
}

// GetAllInputs returns all the transactions referenced by a transaction
// as inputs.
func (bc *BlockChain) GetAllInputs(t *Transaction) []*Transaction {
	txns := make([]*Transaction, len(t.Inputs))
	for _, tx := range t.Inputs {
		txns = append(txns, bc.GetInputTransaction(&tx))
	}
	return txns
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

// CopyBlockByIndex returns a copy of a block in the local chain by index.
func (bc *BlockChain) CopyBlockByIndex(i uint32) (*Block, error) {
	if i >= 0 && i < uint32(len(bc.Blocks)) {
		blk := bc.Blocks[i]
		b := *blk
		b.Transactions = make([]*Transaction, len(blk.Transactions))
		copy(b.Transactions, blk.Transactions)
		return &b, nil
	}
	return nil, errors.New("block request out of bounds")
}

// InputsSpentElsewhere returns true if inputs perported to be only spent
// on transaction t have been spent elsewhere after block index `start`.
func (bc *BlockChain) InputsSpentElsewhere(t *Transaction, start uint32) bool {
	// This implementation runs in O(n * m * l * x)
	// where n = number of blocks in range.
	// 		 m = number of transactions per block.
	// 		 l = number of inputs in a transaction.
	// 		 x = a factor of the hash efficiency function on (1, 2).
	for _, b := range bc.Blocks[start:] {
		for _, txn := range b.Transactions {
			if HashSum(txn) == HashSum(t) {
				continue
			} else {
				if t.InputsIntersect(txn) {
					return true
				}
			}
		}
	}
	return false
}
