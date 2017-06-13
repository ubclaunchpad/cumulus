package pool

import "testing"

func TestGetAndPutTransaction(t *testing.T) {
	p := New()
	bc := newBlockChain()
	if p.Len() != 0 {
		t.Fail()
	}
	tr := newTransaction()
	p.Set(tr, bc)

	if p.Len() != 1 {
		t.Fail()
	}

	p.Delete(tr)
	if p.Len() != 0 {
		t.Fail()
	}
}

func TestUpdatePool(t *testing.T) {
	p := New()
	bc := newBlockChain()
	b := newBlock()
	for _, tr := range b.Transactions {
		p.Set(tr, bc)
		if _, ok := p.Get(tr.Input.Hash); ok {
			t.Fail()
		}
	}
	if p.Len() != len(b.Transactions) {
		t.Fail()
	}

	p.Update(b)
	if p.Len() != 0 {
		t.Fail()
	}
}

func TestGetNewBlock(t *testing.T) {
	b := newBlock()
	bc := newBlockChain()
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
