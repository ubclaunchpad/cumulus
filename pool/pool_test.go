package pool

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/ubclaunchpad/cumulus/blockchain"
)

func TestGetAndSetTransaction(t *testing.T) {
	p := New()
	bc, b := blockchain.NewValidTestChainAndBlock()
	if p.Len() != 0 {
		t.FailNow()
	}
	tr := b.Transactions[1]
	if !p.Set(tr, bc) {
		t.FailNow()
	}

	if p.Len() != 1 {
		t.FailNow()
	}

	r := p.Get(tr.Input.Hash)
	if r != tr {
		t.FailNow()
	}

	p.Delete(tr)
	if p.Len() != 0 {
		t.FailNow()
	}
}

func TestSetBadTransaction(t *testing.T) {
	p := New()
	bc := blockchain.NewTestBlockChain()
	if p.Set(blockchain.NewTestTransaction(), bc) {
		t.FailNow()
	}
}

func TestUpdatePool(t *testing.T) {
	p := New()
	bc, legitBlk := blockchain.NewValidTestChainAndBlock()
	badBlock := blockchain.NewTestBlock()
	if p.Update(badBlock, bc) {
		t.FailNow()
	}

	for _, tr := range legitBlk.Transactions[1:] {
		p.Set(tr, bc)
	}
	if p.Len() == 0 {
		t.FailNow()
	}
	if p.Len() != len(legitBlk.Transactions[1:]) {
		t.FailNow()
	}

	if !p.Update(legitBlk, bc) {
		t.FailNow()
	}
	if p.Len() != 0 {
		t.FailNow()
	}
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
	p.SetUnsafe(tr)
	for i := 0; i < numTxns; i++ {
		p.SetUnsafe(blockchain.NewTestTransaction())
	}
	if p.GetIndex(tr) != 0 {
		t.FailNow()
	}
	for i := 0; i < numTxns; i++ {
		if p.GetIndex(p.Order[i].Transaction) != i {
			t.FailNow()
		}
	}
}

func TestNextBlock(t *testing.T) {
	p := New()
	chain, _ := blockchain.NewValidTestChainAndBlock()
	nBlks := len(chain.Blocks)
	lastBlk := chain.Blocks[nBlks-1]
	numTxns := 1000
	for i := 0; i < numTxns; i++ {
		p.SetUnsafe(blockchain.NewTestTransaction())
	}
	b := p.NextBlock(chain, blockchain.NewWallet().Public(), 1<<18)
	assert.NotNil(t, b)
	assert.True(t, b.Len() < 1<<18)
	assert.True(t, b.Len() > 0)

	// The difference is off by one thanks to cloud transaction.
	assert.Equal(t, len(b.Transactions), numTxns-p.Len()+1)
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
	t2 := blockchain.NewTestTransaction()
	t1.Input.Hash = t2.Input.Hash
	p.SetUnsafe(t1)
	p.SetUnsafe(t2)
	assert.Equal(t, p.Peek(), t2)
	assert.Equal(t, p.Len(), 1)
}
