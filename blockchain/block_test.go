package blockchain

import (
	"bytes"
	"testing"
)

func TestEncodeDecodeBlock(t *testing.T) {
	b1 := newBlock()

	buf := bytes.NewBuffer(make([]byte, 0, b1.Len()))

	b1.Encode(buf)
	b2 := DecodeBlock(buf)

	if HashSum(b1) != HashSum(b2) {
		t.Fail()
	}
}
