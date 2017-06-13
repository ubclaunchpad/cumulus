package pool

import (
	"testing"

	"github.com/ubclaunchpad/cumulus/blockchain"
)

func TestGetAndSetTransaction(t *testing.T) {
	p := New()
	bc, b := blockchain.NewValidChainAndBlock()
	if p.Len() != 0 {
		t.Fail()
	}
	tr := b.Transactions[0]
	if !p.Set(tr, bc) {
		t.Fail()
	}

	if p.Len() != 1 {
		t.Fail()
	}

	r, ok := p.Get(tr.Input.Hash)
	if !ok {
		t.Fail()
	}
	if r != tr {
		t.Fail()
	}

	p.Delete(tr)
	if p.Len() != 0 {
		t.Fail()
	}
}

func TestUpdatePool(t *testing.T) {
	p := New()
	bc, legitBlk := blockchain.NewValidChainAndBlock()
	badBlock := blockchain.NewBlock()
	if p.Update(badBlock, bc) {
		t.Fail()
	}

	for _, tr := range legitBlk.Transactions {
		p.Set(tr, bc)
		if _, ok := p.Get(tr.Input.Hash); !ok {
			t.Fail()
		}
	}
	if p.Len() == 0 {
		t.Fail()
	}
	if p.Len() != len(legitBlk.Transactions) {
		t.Fail()
	}

	if !p.Update(legitBlk, bc) {
		t.Fail()
	}
	if p.Len() != 0 {
		t.Fail()
	}
}

func TestGetNewBlock(t *testing.T) {
	b := blockchain.NewBlock()
	bc := blockchain.NewBlockChain()
	p := New()
	for _, tr := range b.Transactions {
		p.Set(tr, bc)
	}
	newBlk, err := p.GetBlock()
	if err != nil {
		t.Fail()
	}
	for _, tr := range newBlk.Transactions {
		if exists, _ := b.ContainsTransaction(tr); exists {
			t.Fail()
		}
	}
}
