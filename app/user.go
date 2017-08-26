package app

import (
	"errors"

	"github.com/ubclaunchpad/cumulus/consensus"

	"github.com/ubclaunchpad/cumulus/blockchain"
	"github.com/ubclaunchpad/cumulus/msg"

	crand "crypto/rand"
)

// User holds basic user information.
type User struct {
	blockchain.Wallet
	Name      string
	BlockSize uint32
}

// NewUser creates a new user
func NewUser() *User {
	return &User{
		Wallet:    *blockchain.NewWallet(),
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
	// Four steps must occur.
	wallet := a.CurrentUser.Wallet
	pool := a.Pool

	// A legitimate transaction must be built.
	tbody := blockchain.TxBody{
		Sender: wallet.Public(),
		// TODO: Collect inputs.
		Inputs: []blockchain.TxHashPointer{},
		Outputs: []blockchain.TxOutput{
			blockchain.TxOutput{
				Recipient: to,
				Amount:    amount,
			},
		},
	}

	// The transaction must be signed.
	if txn, err := tbody.Sign(a.CurrentUser.Wallet, crand.Reader); err == nil {

		// The transaction must be broadcasted to the peers.
		if err := wallet.SetPending(txn); err != nil {
			return err
		}

		// The transaction must be added to the pool.
		code := pool.Push(txn, a.Chain)
		if code != consensus.ValidTransaction {
			return errors.New("transaction validation failed")
		}

		// The transaction must be broadcasted to the network.
		a.PeerStore.Broadcast(msg.Push{
			ResourceType: msg.ResourceTransaction,
			Resource:     txn,
		})

	} else {
		return err
	}

	return nil
}
