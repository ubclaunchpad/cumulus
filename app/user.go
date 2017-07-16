package app

import "github.com/ubclaunchpad/cumulus/blockchain"

// User holds basic user information.
type User struct {
	// Wallet holds the users wallets (currently just support 1).
	Wallet blockchain.Wallet
	// UserBlockSize is the maximum size of a block in bytes when marshaled
	// as specified by the user.
	BlockSize uint32
}

var currentUser *User

const defaultBlockSize = 1 << 18

func init() {
	// Temporary to create a new user for testing.
	currentUser = NewUser()
}

// NewUser creates a new user
func NewUser() *User {
	return &User{
		Wallet:    blockchain.NewWallet(),
		BlockSize: defaultBlockSize,
	}
}

// GetCurrentUser returns the current user.
func GetCurrentUser() *User {
	return currentUser
}
