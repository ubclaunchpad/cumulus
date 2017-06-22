package blockchain

import (
	"bytes"
	"testing"
)

func TestEncodeDecodeBlock(t *testing.T) {
	b1 := NewTestBlock()

	buf := bytes.NewBuffer(make([]byte, 0, b1.Len()))

	b1.Encode(buf)
	b2 := DecodeBlock(buf)

	if HashSum(b1) != HashSum(b2) {
		t.Fail()
	}
}

func TestContainsTransaction(t *testing.T) {
	b := NewTestBlock()

	if exists, _ := b.ContainsTransaction(b.Transactions[0]); !exists {
		t.Fail()
	}
}
