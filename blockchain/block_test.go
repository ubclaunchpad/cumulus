package blockchain

import (
	"bytes"
	"testing"
)

func TestEncodeBlock(t *testing.T) {
	b1 := newBlock()
	b2 := Block{}

	buf := bytes.NewBuffer(make([]byte, 0, b1.Len()))

	b1.Encode(buf)
	b2.Decode(buf)

	if b1.Hash() != b2.Hash() {
		t.Fail()
	}
}
