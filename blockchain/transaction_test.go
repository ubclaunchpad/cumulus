package blockchain

import (
	"testing"
)

func TestTxBodyLen(t *testing.T) {
	txBody := NewTestTxBody()
	senderLen := AddrLen
	inputLen := 2*(32/8) + HashLen
	outputLen := len(txBody.Outputs) * (64/8 + AddrLen)
	txBodyLen := senderLen + inputLen + outputLen

	if txBody.Len() != txBodyLen {
		t.Fail()
	}
}

func TestTransactionLen(t *testing.T) {
	tx := NewTestTransaction()
	senderLen := AddrLen
	inputLen := 2*(32/8) + HashLen
	outputLen := len(tx.TxBody.Outputs) * (64/8 + AddrLen)
	txBodyLen := senderLen + inputLen + outputLen
	txLen := txBodyLen + SigLen

	if tx.Len() != txLen {
		t.Fail()
	}
}
