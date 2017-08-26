package pool

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/ubclaunchpad/cumulus/blockchain"
	"github.com/ubclaunchpad/cumulus/consensus"
)

func TestGetAndSetTransaction(t *testing.T) {
	p := New()
	bc, b := blockchain.NewValidTestChainAndBlock()
	assert.Equal(t, p.Size(), 0)

	tr := b.Transactions[1]
	assert.Equal(t, p.Push(tr, bc), consensus.ValidTransaction)
	assert.Equal(t, p.Size(), 1)
	assert.ObjectsAreEqual(tr, p.Get(blockchain.HashSum(tr)))

	p.Delete(tr)
	assert.Equal(t, p.Size(), 0)
}

func TestSetBadTransaction(t *testing.T) {
	p := New()
	bc, _ := blockchain.NewValidTestChainAndBlock()

	// This transaction will have bad inputs.
	txn := blockchain.NewTestTransaction()
	code := p.Push(txn, bc)
	assert.NotEqual(t, code, consensus.ValidTransaction)
}

func TestUpdatePool(t *testing.T) {
	p := New()
	bc, legitBlk := blockchain.NewValidTestChainAndBlock()
	badBlock := blockchain.NewTestBlock()

	// Make sure we cant update with a bad block.
	assert.False(t, p.Update(badBlock, bc))

	for _, tr := range legitBlk.Transactions[1:] {
		p.Push(tr, bc)
	}

	// Assert transactions added.
	assert.NotEqual(t, p.Size(), 0)
	assert.Equal(t, p.Size(), len(legitBlk.Transactions[1:]))

	// Assert we can add update with a legit block and it drains pool.
	assert.True(t, p.Update(legitBlk, bc))
	assert.Equal(t, p.Size(), 0)
}

func TestGetNewBlockEmpty(t *testing.T) {
	p := New()
	txn := p.Pop()
	assert.Nil(t, txn)
}

func TestGetIndex(t *testing.T) {
	p := New()
	numTxns := 1000
	tr := blockchain.NewTestTransaction()
	p.PushUnsafe(tr)
	for i := 0; i < numTxns; i++ {
		p.PushUnsafe(blockchain.NewTestTransaction())
	}
	assert.Equal(t, p.GetIndex(tr), 0)

	for i := 0; i < numTxns; i++ {
		assert.Equal(t, p.GetIndex(p.Order[i].Transaction), i)
	}
}

func TestNextBlock(t *testing.T) {
	p := New()
	chain, _ := blockchain.NewValidTestChainAndBlock()
	nBlks := len(chain.Blocks)
	lastBlk := chain.Blocks[nBlks-1]
	numTxns := 1000
	for i := 0; i < numTxns; i++ {
		p.PushUnsafe(blockchain.NewTestTransaction())
	}
	b := p.NextBlock(chain, blockchain.NewWallet().Public(), 1<<18)

	assert.NotNil(t, b)
	assert.True(t, b.Len() < 1<<18)
	assert.True(t, b.Len() > 0)

	// The difference is off by one thanks to cloud transaction.
	assert.Equal(t, len(b.Transactions), numTxns-p.Size()+1)
	assert.Equal(t, blockchain.HashSum(lastBlk), b.LastBlock)
	assert.Equal(t, uint64(0), b.Nonce)
	assert.Equal(t, uint32(nBlks), b.BlockNumber)
}

func TestPeek(t *testing.T) {
	p := New()
	assert.Nil(t, p.Peek())
}

func TestPop(t *testing.T) {
	p := New()
	assert.Nil(t, p.Pop())
}

func TestSetDedupes(t *testing.T) {
	p := New()
	t1 := blockchain.NewTestTransaction()
	p.PushUnsafe(t1)
	p.PushUnsafe(t1)
	assert.Equal(t, p.Size(), 1)
	assert.Equal(t, p.Peek(), t1)
}
