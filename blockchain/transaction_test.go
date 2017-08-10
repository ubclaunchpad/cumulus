package blockchain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTxBodyLen(t *testing.T) {
	txBody := NewTestTxBody()
	senderLen := AddrLen
	inputLen := 2*(32/8) + HashLen
	outputLen := len(txBody.Outputs) * (64/8 + ReprLen)
	txBodyLen := senderLen + inputLen + outputLen

	assert.Equal(t, txBody.Len(), txBodyLen)
}

func TestTransactionLen(t *testing.T) {
	tx := NewTestTransaction()
	senderLen := AddrLen
	inputLen := 2*(32/8) + HashLen
	outputLen := len(tx.TxBody.Outputs) * (64/8 + ReprLen)
	txBodyLen := senderLen + inputLen + outputLen
	txLen := txBodyLen + SigLen

	assert.Equal(t, tx.Len(), txLen)
}
