package app

import "github.com/ubclaunchpad/cumulus/blockchain"

// User holds basic user information.
type User struct {
	blockchain.Wallet
	Name      string
	BlockSize uint32
}

// NewUser creates a new user
func NewUser() *User {
	return &User{
		Wallet:    blockchain.NewWallet(),
		BlockSize: blockchain.DefaultBlockSize,
	}
}

// getCurrentUser gets the current user.
func getCurrentUser() *User {
	// TODO: Check for local user information on disk,
	// If doesnt exist, create new user.
	return NewUser()
}
