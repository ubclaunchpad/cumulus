package pool

import "testing"

func TestGetAndPutTransaction(t *testing.T) {
	p := New()
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
	p := New()
	b := newBlock()
	for _, tr := range b.Transactions {
		p.PutTransaction(tr)
		if len(p.GetTransaction(tr.Input.Hash)) == 0 {
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
	p := New()
	for _, tr := range b.Transactions {
		p.PutTransaction(tr)
	}
	newBlk, err := p.GetNewBlock()
	if err != nil {
		t.Fail()
	}
	for _, tr := range newBlk.Transactions {
		if exists, _ := b.ContainsTransaction(tr); exists {
			t.Fail()
		}
	}
}
