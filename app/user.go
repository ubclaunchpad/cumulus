package app

import (
	"github.com/ubclaunchpad/cumulus/blockchain"
	"github.com/ubclaunchpad/cumulus/msg"

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

	// 1. A legitimate transaction must be built.
	tbody := blockchain.TxBody{
		Sender: getCurrentUser().Wallet.Public(),
		Outputs: []blockchain.TxOutput{
			blockchain.TxOutput{
				Recipient: to,
				Amount:    amount,
			},
		},
	}

	// 2. The transaction must be signed and broadcasted.
	if txn, err := tbody.Sign(*a.CurrentUser.Wallet, crand.Reader); err != nil {
		a.PeerStore.Broadcast(msg.Push{
			ResourceType: msg.ResourceTransaction,
			Resource:     txn,
		})
		a.CurrentUser.Wallet.SetPending(txn)

		// 3. The transaction must be added to the pool.
		a.HandleTransaction(txn)
	} else {
		return err
	}

	return nil
}
