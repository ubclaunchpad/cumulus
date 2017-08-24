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
