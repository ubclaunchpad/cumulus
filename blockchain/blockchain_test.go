package blockchain

import (
	"bytes"
	"reflect"
	"testing"

	log "github.com/Sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestMain(t *testing.T) {
	log.SetLevel(log.DebugLevel)
}

func TestEncodeDecodeBlockChain(t *testing.T) {
	b1 := NewTestBlockChain()

	buf := bytes.NewBuffer(make([]byte, 0, b1.Len()))

	b1.Encode(buf)
	b2 := DecodeBlockChain(buf)

	if HashSum(b1) != HashSum(b2) {
		t.Fail()
	}
}

func TestCopyBlock(t *testing.T) {
	bc, _ := NewValidTestChainAndBlock()
	b, _ := bc.CopyBlockByLastBlockHash(bc.Blocks[1].LastBlock)

	if !reflect.DeepEqual(b, bc.Blocks[1]) {
		t.FailNow()
	}

	// Enforce copy.
	if b == bc.Blocks[1] {
		t.FailNow()
	}
}

func TestWalletRepr(t *testing.T) {
	w := NewWallet()
	assert.Equal(t, len(w.Public().Repr()), 40)
}

func TestAppend(t *testing.T) {
	bc, block := NewValidTestChainAndBlock()
	bc.AppendBlock(block)
	if bc.Head != HashSum(block) || bc.LastBlock() != block {
		t.Fail()
	}
}

func TestRollBack(t *testing.T) {
	bc, block1 := NewValidTestChainAndBlock()
	block2 := NewTestBlock()
	initialHead := bc.LastBlock()
	bc.AppendBlock(block1)
	if bc.Head != HashSum(block1) || bc.LastBlock() != block1 {
		t.FailNow()
	}
	bc.AppendBlock(block2)
	if bc.Head != HashSum(block2) {
		t.FailNow()
	}
	rollbackBlock1 := bc.RollBack()
	if bc.Head != HashSum(block1) || bc.LastBlock() != block1 || rollbackBlock1 != block2 {
		t.FailNow()
	}
	rollbackBlock2 := bc.RollBack()
	if bc.Head != HashSum(initialHead) || bc.LastBlock() != initialHead || rollbackBlock2 != block1 {
		t.FailNow()
	}
}
