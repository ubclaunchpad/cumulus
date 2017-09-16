package blockchain

import (
	"os"
	"reflect"
	"testing"

	log "github.com/Sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestMain(t *testing.T) {
	log.SetLevel(log.DebugLevel)
}

func TestNew(t *testing.T) {
	assert.NotNil(t, New())
}

func TestMarshal(t *testing.T) {
	b := NewTestBlock()
	assert.NotNil(t, b.Marshal())
}

func TestLen(t *testing.T) {
	bc, _ := NewValidTestChainAndBlock()
	assert.True(t, bc.Len() > 0)
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

func TestAppendBlock(t *testing.T) {
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

func TestGetInputTransaction(t *testing.T) {
	bc := NewTestBlockChain()
	txnPtr := &TxHashPointer{
		BlockNumber: 0,
		Index:       0,
	}
	assert.NotNil(t, bc.GetInputTransaction(txnPtr))
}

func TestGetAllInputs(t *testing.T) {
	bc := NewTestBlockChain()
	txn := NewTestTransaction()
	_, err := bc.GetAllInputs(txn)
	assert.EqualError(t, err, "Input transaction not found")
	bc, b := NewValidTestChainAndBlock()
	expectedResult := []*Transaction{
		bc.Blocks[1].Transactions[1],
	}
	result, err := bc.GetAllInputs(b.Transactions[1])
	assert.Nil(t, err)
	assert.Equal(t, expectedResult, result)
}

func TestBlockchainContainsTransaction(t *testing.T) {
	bc, txn := NewValidChainAndTxn()
	found, blockIndex, txnIndex := bc.ContainsTransaction(bc.Blocks[1].Transactions[0], 0, 2)
	assert.True(t, found)
	assert.Equal(t, blockIndex, uint32(1))
	assert.Equal(t, txnIndex, uint32(0))
	found, _, _ = bc.ContainsTransaction(txn, 0, 2)
	assert.False(t, found)
}

func TestGetBlockByLastBlockHash(t *testing.T) {
	bc := NewTestBlockChain()
	block, err := bc.GetBlockByLastBlockHash(bc.Blocks[1].BlockHeader.LastBlock)
	assert.Nil(t, err)
	assert.Equal(t, block, bc.Blocks[1])
	_, err = bc.GetBlockByLastBlockHash(NewTestHash())
	assert.EqualError(t, err, "No such block")
}
