package blockchain

import "testing"

func TestGetAndPutTransaction(t *testing.T) {
	p := NewPool(newWallet().Public())

	if p.Len() != 0 {
		t.Fail()
	}

	tr := newTransaction()

	if p.PutTransaction(tr) {
		t.Fail()
	}

	if p.Len() != 1 {
		t.Fail()
	}

	if p.RemoveTransaction(tr) {
		t.Fail()
	}

	if p.Len() != 0 {
		t.Fail()
	}
}

func TestUpdatePool(t *testing.T) {
	b := newBlock()
	p := NewPool(newWallet().Public())
	for _, tr := range b.Transactions {
		p.PutTransaction(tr)
		if len(p.GetTransactions(tr.Input.Hash)) == 0 {
			t.Fail()
		}
	}
	if p.Len() != len(b.Transactions) {
		t.Fail()
	}

	p.UpdatePool(b)
	if p.Len() != 0 {
		t.Fail()
	}
}

func TestGetNewBlock(t *testing.T) {
	b := newBlock()
	p := NewPool(newWallet().Public())
	for _, tr := range b.Transactions {
		p.PutTransaction(tr)
	}
	newBlk, success := p.GetNewBlock()
	if !success {
		t.Fail()
	}
	for _, tr := range newBlk.Transactions {
		if exists, _ := b.ContainsTransaction(tr); exists {
			t.Fail()
		}
	}
}
