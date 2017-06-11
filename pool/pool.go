package pool

import (
	"errors"

	"github.com/cevaris/ordered_map"
	"github.com/ubclaunchpad/cumulus/blockchain"
)

// Pool is a set of valid Transactions.
type Pool struct {
	Transactions ordered_map.OrderedMap
}

// New initializes a new pool.
func New() *Pool {
	return &Pool{
		Transactions: *ordered_map.NewOrderedMap(),
	}
}

// Len returns the number of transactions in the Pool.
func (p *Pool) Len() int {
	return 0
}

// GetTransaction returns the transctions with input transaction Hash h.
func (p *Pool) GetTransaction(h blockchain.Hash) []*blockchain.Transaction {
	return nil
}

// PutTransaction inserts a transaction into the pool, returning
// true if the Transaction was inserted (was valid).
func (p *Pool) PutTransaction(t *blockchain.Transaction) bool {
	return false
}

// RemoveTransaction removes a transaction from the Pool, returning
// true if the Transaction existed in the pool.
func (p *Pool) RemoveTransaction(t *blockchain.Transaction) bool {
	return false
}

// UpdatePool updates the Pool by removing the Transactions found in the
// Block. If the Block is found invalid, then false is returned and no
// Transactions are removed from the Pool.
func (p *Pool) UpdatePool(b *blockchain.Block) bool {
	return false
}

// GetNewBlock returns a new Block from the highest priority Transactions in
// the Pool, as well as a error indicating whether there were any
// Transactions to create a Block.
func (p *Pool) GetNewBlock() (*blockchain.Block, error) {
	return newBlock(), errors.New("no transactions in pool")
}
