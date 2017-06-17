package pool

import (
	"testing"

	"github.com/ubclaunchpad/cumulus/blockchain"
)

func TestGetAndSetTransaction(t *testing.T) {
	p := New()
	bc, b := blockchain.NewValidChainAndBlock()
	if p.Len() != 0 {
		t.FailNow()
	}
	tr := b.Transactions[0]
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

func TestUpdatePool(t *testing.T) {
	p := New()
	bc, legitBlk := blockchain.NewValidChainAndBlock()
	badBlock := blockchain.NewBlock()
	if p.Update(badBlock, bc) {
		t.FailNow()
	}

	for _, tr := range legitBlk.Transactions {
		p.Set(tr, bc)
	}
	if p.Len() == 0 {
		t.FailNow()
	}
	if p.Len() != len(legitBlk.Transactions) {
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
	bc, b := blockchain.NewValidChainAndBlock()
	for _, tr := range b.Transactions {
		if !p.Set(tr, bc) {
			t.FailNow()
		}
	}
	nTxns := len(b.Transactions) + 12 // arbitrary.
	txns := p.GetTxns(nTxns)
	for _, tr := range txns {
		if ok, _ := b.ContainsTransaction(tr); !ok {
			t.FailNow()
		}
	}
}

func TestGetNewBlockEmpty(t *testing.T) {
	p := New()
	txns := p.GetTxns(305)
	if len(txns) != 0 {
		t.FailNow()
	}
}
