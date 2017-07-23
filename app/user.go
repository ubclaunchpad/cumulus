package app

import "github.com/ubclaunchpad/cumulus/blockchain"

// User holds basic user information.
type User struct {
	// Account holds the users wallet(s).
	HotWallet
	// UserBlockSize is the maximum size of a block in bytes when marshaled
	// as specified by the user.
	BlockSize uint32
}

// HotWallet is a representation of the users wallet.
type HotWallet struct {
	Name string
	blockchain.Wallet
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
		HotWallet: HotWallet{
			Wallet: blockchain.NewWallet(),
			Name:   "default",
		},
		BlockSize: defaultBlockSize,
	}
}

// GetCurrentUser returns the current user.
func GetCurrentUser() *User {
	return currentUser
}
