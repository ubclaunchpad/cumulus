package blockchain

import (
	"encoding/json"
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
	w.Balance = txn.GetTotalOutput()

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

	// Make a fake blockchain that contains transactions that this transaction
	// references as it's inputs
	block := &Block{
		Transactions: []*Transaction{
			&Transaction{
				TxBody: TxBody{
					Outputs: []TxOutput{
						TxOutput{
							Amount:    txn.TxBody.Outputs[0].Amount,
							Recipient: txn.Sender.Repr(),
						},
					},
				},
			},
		},
	}
	bc := &BlockChain{
		Blocks: []*Block{block},
	}

	w.Balance = txn.GetTotalOutput()
	w.SetAllPending([]*Transaction{txn})

	// Drop all pending
	result, _ := w.IsPending(txn)
	assert.True(t, result)
	w.DropAllPending([]*Transaction{txn}, bc)
	result, _ = w.IsPending(txn)
	assert.False(t, result)
}

func TestGetWalletBalances(t *testing.T) {
	w := NewWallet()
	txn := NewTestTransaction()
	w.Balance = txn.GetTotalOutput()
	w.SetAllPending([]*Transaction{txn})

	assert.Equal(t, w.Balance, txn.GetTotalOutput())
	assert.Equal(t, w.GetEffectiveBalance(), uint64(0))
}

func TestUnmarshalJSON(t *testing.T) {
	w1 := NewWallet()
	walletBytes, err := json.Marshal(w1)
	assert.Nil(t, err)
	var w2 Wallet
	err = w2.UnmarshalJSON(walletBytes)
	assert.Nil(t, err)
	assert.Equal(t, w1, &w2)
}

func TestUpdate(t *testing.T) {
	bc, wallets := NewValidBlockChainFixture()
	assert.Nil(t, wallets["bob"].Update(bc.Blocks[2], bc))
	assert.Equal(t, wallets["bob"].Balance, uint64(1))
}

func TestRefresh(t *testing.T) {
	bc, wallets := NewValidBlockChainFixture()
	assert.Nil(t, wallets["bob"].Refresh(bc))
	assert.Equal(t, wallets["bob"].Balance, uint64(1))
}
