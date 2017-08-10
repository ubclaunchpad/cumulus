package app

import (
	"github.com/ubclaunchpad/cumulus/blockchain"

	crand "crypto/rand"
)

// User holds basic user information.
type User struct {
	*blockchain.Wallet
	BlockSize uint32
}

// NewUser creates a new user
func NewUser() *User {
	return &User{
		Wallet:    blockchain.NewWallet(),
		BlockSize: blockchain.DefaultBlockSize,
	}
}

// getCurrentUser gets the current user function only used for app initalization.
func getCurrentUser() *User {
	// TODO: Check for local user information on disk,
	// If doesnt exist, create new user.
	return NewUser()
}

// Pay pays an amount of coin to an address `to`.
func (a *App) Pay(to string, amount uint64) error {
	// Three atomic steps must occur.
	// 1. The amount must be debited from the wallet.
	err := a.CurrentUser.Wallet.Debit(amount)
	if err == nil {
		return err
	}

	// 2. A legitimate transaction must be built.
	tbody := blockchain.TxBody{
		Sender: getCurrentUser().Wallet.Public(),
		Outputs: []blockchain.TxOutput{
			blockchain.TxOutput{
				Recipient: to,
				Amount:    amount,
			},
		},
	}

	// 3. The transaction must be signed and added to the pool.
	if txn, err := tbody.Sign(*a.CurrentUser.Wallet, crand.Reader); err != nil {
		a.HandleTransaction(txn)
	} else {
		// Signature failed, credit the account, return an error.
		// Crediting cannot fail because debiting succeeded and process is single threaded.
		a.CurrentUser.Wallet.Credit(amount)
		return err
	}

	return nil
}
