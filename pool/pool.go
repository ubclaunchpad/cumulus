package pool

import (
	"time"

	log "github.com/Sirupsen/logrus"

	"github.com/ubclaunchpad/cumulus/blockchain"
	"github.com/ubclaunchpad/cumulus/common/util"
	"github.com/ubclaunchpad/cumulus/consensus"
	"github.com/ubclaunchpad/cumulus/miner"
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

// Empty returns true if the pool has exactly zero transactions in it.
func (p *Pool) Empty() bool {
	return p.Len() == 0
}

// Get returns the tranasction with input transaction Hash h.
func (p *Pool) Get(h blockchain.Hash) *blockchain.Transaction {
	return p.ValidTransactions[h].Transaction
}

// GetN returns the Nth transaction in the pool.
func (p *Pool) GetN(N int) *blockchain.Transaction {
	return p.Order[N].Transaction
}

// GetIndex returns the index of the transaction in the ordering.
func (p *Pool) GetIndex(t *blockchain.Transaction) int {
	target := p.ValidTransactions[t.Input.Hash].Time
	return getIndex(p.Order, target, 0, p.Len()-1)
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

// Push inserts a transaction into the pool, returning
// true if the Transaction was inserted (was valid).
func (p *Pool) Push(t *blockchain.Transaction, bc *blockchain.BlockChain) bool {
	if ok, err := consensus.VerifyTransaction(bc, t); ok {
		p.set(t)
		return true
	} else {
		log.Debug(err)
		return false
	}
}

// PushUnsafe adds a transaction to the pool without validation.
func (p *Pool) PushUnsafe(t *blockchain.Transaction) {
	p.set(t)
}

// Silently adds a transaction to the pool.
// Deletes a transaction if it exists from the same
// input hash.
func (p *Pool) set(t *blockchain.Transaction) {
	if txn, ok := p.ValidTransactions[t.Input.Hash]; ok {
		p.Delete(txn.Transaction)
	}
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
		p.Order = append(p.Order[0:i], p.Order[i+1:]...)
		delete(p.ValidTransactions, t.Input.Hash)
	}
}

// Update updates the Pool by removing the Transactions found in the
// Block. If the Block is found invalid wrt bc, then false is returned and no
// Transactions are removed from the Pool.
func (p *Pool) Update(b *blockchain.Block, bc *blockchain.BlockChain) bool {
	if ok, _ := consensus.VerifyBlock(bc, b); !ok {
		return false
	}
	for _, t := range b.Transactions {
		p.Delete(t)
	}
	return true
}

// Pop returns the next transaction and removes it from the pool.
func (p *Pool) Pop() *blockchain.Transaction {
	if p.Len() > 0 {
		next := p.GetN(0)
		p.Delete(next)
		return next
	}
	return nil
}

// Peek returns the next transaction and does not remove it from the pool.
func (p *Pool) Peek() *blockchain.Transaction {
	if p.Len() > 0 {
		return p.GetN(0)
	}
	return nil
}

// NextBlock produces a new block from the pool for mining. The block returned
// may not contain transactions if there are none left in the transaction pool.
func (p *Pool) NextBlock(chain *blockchain.BlockChain,
	address blockchain.Address, size uint32) *blockchain.Block {
	var txns []*blockchain.Transaction

	// Hash the last block in the chain.
	ix := len(chain.Blocks) - 1
	lastHash := blockchain.HashSum(chain.Blocks[ix])

	// Build a new block for mining.
	b := &blockchain.Block{
		BlockHeader: blockchain.BlockHeader{
			BlockNumber: uint32(len(chain.Blocks)),
			LastBlock:   lastHash,
			Time:        util.UnixNow(),
			Nonce:       0,
		}, Transactions: txns,
	}

	// Prepend the cloudbase transaction for this miner.
	miner.CloudBase(b, chain, address)

	// Try to grab as many transactions as the block will allow.
	// Test each transaction to see if we break size before adding.
	for p.Len() > 0 {
		nextSize := p.Peek().Len()
		if b.Len()+nextSize < int(size) {
			b.Transactions = append(b.Transactions, p.Pop())
		} else {
			break
		}
	}

	return b
}
