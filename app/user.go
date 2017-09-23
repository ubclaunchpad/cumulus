package app

import (
	"encoding/json"
	"errors"
	"fmt"
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

// Public returns the public key of the given user
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

// LoadUser attempts to read user info from the file with the given name in the
// current working directory in JSON format. On success this returns
// a pointer to a new user constructed from the information in the file.
// If an error occurrs it is returned.
func LoadUser(fileName string) (*User, error) {
	userFile, err := os.OpenFile(fileName, os.O_RDONLY|os.O_CREATE, 0644)
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
	wallet := a.CurrentUser.Wallet
	pool := a.Pool

	// Collect input transactions who's total output to the sender is >= the
	// given amount
	inputTxns, totalInput, err := a.collectInputsForTxn(wallet.Public().Repr(), amount)
	if err != nil {
		return err
	}

	// A legitimate transaction must be built.
	tbody := blockchain.TxBody{
		Sender: wallet.Public(),
		Inputs: *inputTxns,
		Outputs: []blockchain.TxOutput{
			blockchain.TxOutput{
				Recipient: to,
				Amount:    amount,
			},
		},
	}

	// Any change left over gets sent back to the sender
	if totalInput > amount {
		tbody.Outputs = append(tbody.Outputs, blockchain.TxOutput{
			Amount:    totalInput - amount,
			Recipient: wallet.Public().Repr(),
		})
	}

	// The transaction must be signed.
	if txn, err := tbody.Sign(*a.CurrentUser.Wallet, crand.Reader); err == nil {
		// The transaction must be added to the pool.
		code := pool.Push(txn, a.Chain)
		if code != consensus.ValidTransaction {
			return fmt.Errorf("Transaction validation failed with code %d", code)
		}

		// The transaction must be added to the wallet's pending transcations
		if err := wallet.SetPending(txn); err != nil {
			return err
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

// collectInputsForTxn returns a list of input transactions for a new transaction
// from the given sender of the given amount, and the total value of all the
// inputs returned. Returns an error if there are not enough transactions to the
// given sender in the blockchain to make a new transaction of the given amount.
func (a *App) collectInputsForTxn(sender string, amount uint64) (*[]blockchain.TxHashPointer,
	uint64, error) {

	a.Chain.RLock()
	defer a.Chain.RUnlock()

	total := uint64(0)
	inputs := make([]blockchain.TxHashPointer, 0)

	// Iterate through the blockchain in reverse, searching for transactions to
	// sender.
	for i := len(a.Chain.Blocks) - 1; i > 0; i-- {
		block := a.Chain.Blocks[i]
		txns := *block.GetTransactionsTo(sender)

		// Add all the transactions to sender to our list of potential input
		// transactions until the total is greater that or equal to the amount
		// for the transaction we want to send.
		for i, txn := range txns {
			outputToSender := txn.GetTotalOutputFor(sender)
			txnPtr := blockchain.TxHashPointer{
				BlockNumber: block.BlockNumber,
				Hash:        blockchain.HashSum(block),
				Index:       uint32(i),
			}
			if outputToSender >= amount {
				inputs = []blockchain.TxHashPointer{txnPtr}
				return &inputs, outputToSender, nil
			}
			inputs = append(inputs, txnPtr)
			total += outputToSender
			if total >= amount {
				return &inputs, total, nil
			}
		}
	}
	return nil, total, errors.New("Insufficient funds")
}
