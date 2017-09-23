package blockchain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTxBodyLen(t *testing.T) {
	txBody := NewTestTxBody()
	senderLen := AddrLen
	inputLen := len(txBody.Inputs) * (2*(32/8) + HashLen)
	outputLen := len(txBody.Outputs) * (64/8 + ReprLen)
	txBodyLen := senderLen + inputLen + outputLen

	assert.Equal(t, txBody.Len(), txBodyLen)
}

func TestTransactionLen(t *testing.T) {
	tx := NewTestTransaction()
	senderLen := AddrLen
	inputLen := len(tx.TxBody.Inputs) * (2*(32/8) + HashLen)
	outputLen := len(tx.TxBody.Outputs) * (64/8 + ReprLen)
	txBodyLen := senderLen + inputLen + outputLen
	txLen := txBodyLen + SigLen

	assert.Equal(t, tx.Len(), txLen)
}

func TestTransactionEqual(t *testing.T) {
	t1 := NewTestTransaction()
	t2 := NewTestTransaction()
	assert.True(t, t1.Equal(t1))
	assert.False(t, t1.Equal(t2))
}

func TestTransactionGetTotalOutput(t *testing.T) {
	tx := NewTestTransaction()
	tx.Outputs = []TxOutput{
		TxOutput{
			Recipient: tx.Outputs[0].Recipient,
			Amount:    5,
		},
	}
	assert.Equal(t, tx.GetTotalOutput(), uint64(5))
}

func TestInputSet(t *testing.T) {
	txn := NewTestTransaction()
	inSet := txn.InputSet()
	assert.Equal(t, inSet.Size(), len(txn.Inputs))
	for _, tx := range txn.Inputs {
		assert.True(t, inSet.Has(tx))
	}
}

func TestInputIntersection(t *testing.T) {
	txn := NewTestTransaction()
	ixn := txn.InputIntersection(txn)
	assert.Equal(t, ixn.Size(), len(txn.Inputs))
	for _, tx := range txn.Inputs {
		assert.True(t, ixn.Has(tx))
	}
	assert.True(t, txn.InputsIntersect(txn))
}

func TestGetTotalOutputFor(t *testing.T) {
	bc, wallets := NewValidBlockChainFixture()

	// We know alice gets sent 3 coins in block 1, txn 1.
	t1 := bc.Blocks[1].Transactions[1]
	actual := t1.GetTotalOutputFor(wallets["alice"].Public().Repr())
	assert.Equal(t, actual, uint64(3))

	// We know bob gets sent 1 coins in block 2, txn 1.
	t2 := bc.Blocks[2].Transactions[1]
	actual = t2.GetTotalOutputFor(wallets["bob"].Public().Repr())
	assert.Equal(t, actual, uint64(1))
}
