package app

import "github.com/ubclaunchpad/cumulus/blockchain"

// User holds basic user information.
type User struct {
	HotWallet
	BlockSize uint32
}

// HotWallet is a representation of the users wallet.
type HotWallet struct {
	Name string
	blockchain.Wallet
}

// NewUser creates a new user
func NewUser() *User {
	return &User{
		HotWallet: HotWallet{
			Wallet: blockchain.NewWallet(),
			Name:   "default",
		},
		BlockSize: blockchain.DefaultBlockSize,
	}
}
