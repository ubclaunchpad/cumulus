package blockchain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewWallet(t *testing.T) {
	w := NewWallet()
	assert.Equal(t, w.GetEffectiveBalance(), uint64(0))
	assert.Equal(t, len(w.PendingTxns), 0)
}

func TestSetAllPending(t *testing.T) {
	w := NewWallet()
	txn := NewTestTransaction()

	// Set the balance approprately to handle the txn.
	w.SetBalance(txn.GetTotalOutput())

	// Set and check.
	w.SetAllPending([]*Transaction{txn})
	result, i := w.IsPending(txn)

	// Make sure transaction is actually pending.
	assert.True(t, result)
	assert.Equal(t, i, 0)
}

func TestDropAllPending(t *testing.T) {
	w := NewWallet()
	txn := NewTestTransaction()
	w.SetBalance(txn.GetTotalOutput())
	w.SetAllPending([]*Transaction{txn})

	// Drop all pending
	result, _ := w.IsPending(txn)
	assert.True(t, result)
	w.DropAllPending([]*Transaction{txn})
	result, _ = w.IsPending(txn)
	assert.False(t, result)
}
