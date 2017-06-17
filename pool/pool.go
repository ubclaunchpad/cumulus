package pool

import (
	"time"

	"github.com/ubclaunchpad/cumulus/blockchain"
)

// PooledTransaction is a Transaction with a timetamp.
type PooledTransaction struct {
	Transaction *blockchain.Transaction
	Time        time.Time
}

// Pool is a set of valid Transactions.
type Pool struct {
	Order             []*PooledTransaction
	ValidTransactions map[blockchain.Hash]*PooledTransaction
}

// New initializes a new pool.
func New() *Pool {
	return &Pool{
		Order:             []*PooledTransaction{},
		ValidTransactions: map[blockchain.Hash]*PooledTransaction{},
	}
}

// Len returns the number of transactions in the Pool.
func (p *Pool) Len() int {
	return len(p.ValidTransactions)
}

// Get returns the transction with input transaction Hash h.
func (p *Pool) Get(h blockchain.Hash) *blockchain.Transaction {
	return p.ValidTransactions[h].Transaction
}

// GetN returns the Nth transaction in the pool.
func (p *Pool) GetN(N int) *blockchain.Transaction {
	return p.Order[N].Transaction
}

// GetIndex returns the index of the transaction in the ordering.
func (p *Pool) GetIndex(t *blockchain.Transaction) int {
	return getIndex(p.Order, p.ValidTransactions[t.Input.Hash].Time,
		0, p.Len()-1)
}

// getIndex does a binary search for a PooledTransaction by timestamp.
func getIndex(a []*PooledTransaction, target time.Time, low, high int) int {
	mid := (low + high) / 2
	if a[mid].Time == target {
		return mid
	} else if target.Before(a[mid].Time) {
		return getIndex(a, target, low, mid-1)
	} else {
		return getIndex(a, target, mid+1, high)
	}
}

// Set inserts a transaction into the pool, returning
// true if the Transaction was inserted (was valid).
func (p *Pool) Set(t *blockchain.Transaction, bc *blockchain.BlockChain) bool {
	if ok, _ := bc.ValidTransaction(t); ok {
		p.set(t)
		return true
	}
	return false
}

// SetUnsafe adds a transaction to the pool without validation.
func (p *Pool) SetUnsafe(t *blockchain.Transaction) {
	p.set(t)
}

func (p *Pool) set(t *blockchain.Transaction) {
	vt := &PooledTransaction{
		Transaction: t,
		Time:        time.Now(),
	}
	p.Order = append(p.Order, vt)
	p.ValidTransactions[t.Input.Hash] = vt
}

// Delete removes a transaction from the Pool.
func (p *Pool) Delete(t *blockchain.Transaction) {
	vt, ok := p.ValidTransactions[t.Input.Hash]
	if ok {
		i := p.GetIndex(vt.Transaction)
		p.Order = append(p.Order[0:i], p.Order[i:p.Len()-1]...)
		delete(p.ValidTransactions, t.Input.Hash)
	}
}

// Update updates the Pool by removing the Transactions found in the
// Block. If the Block is found invalid wrt bc, then false is returned and no
// Transactions are removed from the Pool.
func (p *Pool) Update(b *blockchain.Block, bc *blockchain.BlockChain) bool {
	if ok, _ := bc.ValidBlock(b); !ok {
		return false
	}
	for _, t := range b.Transactions {
		p.Delete(t)
	}
	return true
}

// PopTxns returns the the largest of l or size of pool transactions.
// It selects the highest priority transactions, and removes them from the pool.
func (p *Pool) PopTxns(l int) []*blockchain.Transaction {
	if p.Len() == 0 {
		return make([]*blockchain.Transaction, 0)
	}
	if p.Len() < l {
		l = p.Len()
	}
	txns := make([]*blockchain.Transaction, l)
	for i := 0; i < l; i++ {
		t := p.GetN(i)
		txns[i] = t
		p.Delete(t)
	}
	return txns
}
