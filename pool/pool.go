package pool

import "github.com/ubclaunchpad/cumulus/blockchain"

// Pool is a set of valid Transactions.
type Pool struct {
	OrderMap        map[blockchain.Hash]int
	OrderReverseMap map[int]blockchain.Hash
	Transactions    map[blockchain.Hash]*blockchain.Transaction
}

// New initializes a new pool.
func New() *Pool {
	return &Pool{
		OrderMap:        make(map[blockchain.Hash]int, 0),
		OrderReverseMap: map[int]blockchain.Hash{},
		Transactions:    map[blockchain.Hash]*blockchain.Transaction{},
	}
}

// Len returns the number of transactions in the Pool.
func (p *Pool) Len() int {
	return len(p.Transactions)
}

// Get returns the transction with input transaction Hash h.
func (p *Pool) Get(h blockchain.Hash) *blockchain.Transaction {
	return p.Transactions[h]
}

// GetN returns the Nth transaction in the pool.
func (p *Pool) GetN(N int) *blockchain.Transaction {
	return p.Transactions[p.OrderReverseMap[N]]
}

// Set inserts a transaction into the pool, returning
// true if the Transaction was inserted (was valid).
func (p *Pool) Set(t *blockchain.Transaction, bc *blockchain.BlockChain) bool {
	ok, _ := bc.ValidTransaction(t)
	if ok {
		p.OrderMap[t.Input.Hash] = p.Len()
		p.OrderReverseMap[p.Len()] = t.Input.Hash
		p.Transactions[t.Input.Hash] = t
	}
	return ok
}

// Delete removes a transaction from the Pool.
func (p *Pool) Delete(t *blockchain.Transaction) {
	i := p.OrderMap[t.Input.Hash]
	if i < p.Len() {
		delete(p.Transactions, t.Input.Hash)
		delete(p.OrderMap, t.Input.Hash)
		delete(p.OrderReverseMap, i)
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

// GetTxns returns the the largest of l or size of pool transactions.
// It selects the highest priority transactions.
func (p *Pool) GetTxns(l int) []*blockchain.Transaction {
	var txns []*blockchain.Transaction
	if p.Len() == 0 {
		return txns
	}
	if p.Len() < l {
		l = p.Len()
	}
	i := 0
	for i < l {
		txns = append(txns, p.GetN(i))
		i++
	}
	return txns
}
