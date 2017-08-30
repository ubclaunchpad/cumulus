package app

import (
	"encoding/json"
	"errors"
	"os"

	crand "crypto/rand"

	"github.com/ubclaunchpad/cumulus/blockchain"
	"github.com/ubclaunchpad/cumulus/consensus"
	"github.com/ubclaunchpad/cumulus/msg"
)

// User holds basic user information.
type User struct {
	Wallet    *blockchain.Wallet
	Name      string
	BlockSize uint32
}

// NewUser creates a new user
func NewUser() *User {
	return &User{
		Wallet:    blockchain.NewWallet(),
		BlockSize: blockchain.DefaultBlockSize,
		Name:      "Default User",
	}
}

// Public returns the public key (address) of the given user
func (u *User) Public() blockchain.Address {
	return u.Wallet.Public()
}

// Save writes the user to a file of the given name in the current working
// directory in JSON format. It returns an error if one occurred, or a pointer
// to the file that was written to on success.
func (u *User) Save(fileName string) error {
	userFile, err := os.OpenFile(fileName, os.O_WRONLY|os.O_CREATE, 0775)
	if err != nil {
		return err
	}
	defer userFile.Close()

	userBytes, err := json.Marshal(u)
	if err != nil {
		return err
	}

	if _, err = userFile.Write(userBytes); err != nil {
		return err
	}
	return nil
}

// Load attempts to read user info from the file with the given name in the
// current working directory in JSON format. On success this returns
// a pointer to a new user constructed from the information in the file.
// If an error occurrs it is returned.
func Load(fileName string) (*User, error) {
	userFile, err := os.OpenFile(fileName, os.O_RDONLY|os.O_CREATE, 0775)
	if err != nil {
		return nil, err
	}
	defer userFile.Close()

	dec := json.NewDecoder(userFile)
	dec.UseNumber()

	var u User
	if err := dec.Decode(&u); err != nil {
		return nil, err
	}
	return &u, nil
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
	if txn, err := tbody.Sign(*a.CurrentUser.Wallet, crand.Reader); err == nil {

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
