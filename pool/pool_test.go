package pool

import (
	"testing"

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

func TestGetTxns(t *testing.T) {
	p := New()
	bc, b := blockchain.NewValidTestChainAndBlock()
	for _, tr := range b.Transactions[1:] {
		if !p.Set(tr, bc) {
			t.FailNow()
		}
	}
	nTxns := len(b.Transactions[1:]) + 12 // arbitrary.
	txns := p.PopTxns(nTxns)
	for _, tr := range txns {
		if ok, _ := b.ContainsTransaction(tr); !ok {
			t.FailNow()
		}
	}
	if p.Len() != 0 {
		t.FailNow()
	}
}

func TestGetNewBlockEmpty(t *testing.T) {
	p := New()
	txns := p.PopTxns(305)
	if len(txns) != 0 {
		t.FailNow()
	}
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
