package blockchain

import (
	"os"
	"reflect"
	"testing"

	log "github.com/Sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

const blockchainFileName = "blockchain.json"

func TestMain(t *testing.T) {
	log.SetLevel(log.DebugLevel)
}

func TestSaveAndLoad(t *testing.T) {
	bc1, _ := NewValidTestChainAndBlock()
	assert.Nil(t, bc1.Save("blockchainTestFile.json"))
	bc2, err := Load("blockchainTestFile.json")
	assert.Nil(t, err)
	assert.Equal(t, bc1.Head, bc2.Head)
	assert.Equal(t, bc1.Blocks, bc2.Blocks)
	assert.Nil(t, os.Remove("blockchainTestFile.json"))
}

func TestGetBlock(t *testing.T) {
	bc, _ := NewValidTestChainAndBlock()
	b, _ := bc.GetBlockByLastBlockHash(bc.Blocks[1].LastBlock)

	if !reflect.DeepEqual(b, bc.Blocks[1]) {
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
