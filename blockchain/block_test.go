package blockchain

import (
	"bytes"
	"testing"

	"github.com/ubclaunchpad/cumulus/common/util"
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

func TestBlockHeaderLen(t *testing.T) {
	bh := &BlockHeader{
		0,
		NewTestHash(),
		NewValidTestTarget(),
		util.UnixNow(),
		0,
		[]byte{0x00, 0x01, 0x02},
	}

	len := 2*(32/8) + 64/8 + 2*HashLen + 3

	if bh.Len() != len {
		t.Fail()
	}

	bh = &BlockHeader{
		0,
		NewTestHash(),
		NewValidTestTarget(),
		util.UnixNow(),
		0,
		[]byte{},
	}

	len = 2*(32/8) + 64/8 + 2*HashLen

	if bh.Len() != len {
		t.Fail()
	}
}

func TestEqual(t *testing.T) {
	block1 := NewTestBlock()
	if !(&block1.BlockHeader).Equal(&block1.BlockHeader) {
		t.FailNow()
	}

	equalBlockHeader := BlockHeader{
		BlockNumber: block1.BlockNumber,
		LastBlock:   block1.LastBlock,
		Target:      block1.Target,
		Time:        block1.Time,
		Nonce:       block1.Nonce,
		ExtraData:   []byte("OneOrTwoOrMaybeEvenThreeOrFour"),
	}

	if !(&block1.BlockHeader).Equal(&equalBlockHeader) {
		t.FailNow()
	}

	block2 := NewTestBlock()
	if (&block1.BlockHeader).Equal(&block2.BlockHeader) {
		t.FailNow()
	}
}
