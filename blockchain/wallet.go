package blockchain

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"fmt"
	"math/big"
)

// The curve we use for our ECC crypto.
var curve = elliptic.P256()

// Wallet represents a Cumulus wallet address in the blockchain.
type Wallet ecdsa.PublicKey

// Signature represents a signature of a transaction.
type Signature struct {
	X big.Int
	Y big.Int
}

// Marshal converts a signature to a byte slice
func (s *Signature) Marshal() []byte {
	return append(s.X.Bytes(), s.Y.Bytes()...)
}

// New creates a new Wallet backed by a ECC key pair. Uses system entropy.
func newWallet() (*Wallet, error) {
	k, err := ecdsa.GenerateKey(curve, rand.Reader)
	if err != nil {
		return nil, err
	}
	pk := Wallet(k.PublicKey)
	return &pk, nil
}

// String returns a human-readable string representation of a wallet
func (w *Wallet) String() string {
	return fmt.Sprintf("%x-%x", w.X, w.Y)
}

// Marshal converts the Wallet to a byte slice
func (w *Wallet) Marshal() []byte {
	return elliptic.Marshal(curve, w.X, w.Y)
}

// Equals checks whether two wallets are the same.
func (w *Wallet) Equals(other *Wallet) bool {
	return w.X.Cmp(other.X) == 0 && w.Y.Cmp(other.Y) == 0
}
