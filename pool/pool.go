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
	return p.Transactions.Len()
}

// Get returns the transctions with input transaction Hash h.
func (p *Pool) Get(h blockchain.Hash) (interface{}, bool) {
	return p.Transactions.Get(h)
}

// Set inserts a transaction into the pool, returning
// true if the Transaction was inserted (was valid).
func (p *Pool) Set(t *blockchain.Transaction, bc *blockchain.BlockChain) {
	if ok, _ := bc.ValidTransaction(t); ok {
		p.Transactions.Set(t.Input.Hash, t)
	}
}

// Delete removes a transaction from the Pool.
func (p *Pool) Delete(t *blockchain.Transaction) {
	p.Transactions.Delete(t.Input.Hash)
}

// Update updates the Pool by removing the Transactions found in the
// Block. If the Block is found invalid, then false is returned and no
// Transactions are removed from the Pool.
func (p *Pool) Update(b *blockchain.Block) bool {
	return false
}

// GetBlock returns a new Block from the highest priority Transactions in
// the Pool, as well as a error indicating whether there were any
// Transactions to create a Block.
func (p *Pool) GetBlock() (*blockchain.Block, error) {
	return newBlock(), errors.New("no transactions in pool")
}
