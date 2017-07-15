package app

import "github.com/ubclaunchpad/cumulus/blockchain"

// User holds basic user information.
type User struct {
	// Wallet holds the users wallets (currently just support 1).
	Wallet blockchain.Wallet
}

var currentUser *User

func init() {
	// Temporary to create a new user for testing.
	currentUser = NewUser()
}

// NewUser creates a new user
func NewUser() *User {
	return &User{
		Wallet: blockchain.NewWallet(),
	}
}

// GetCurrentUser returns the current user.
func GetCurrentUser() *User {
	return currentUser
}
