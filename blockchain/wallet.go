package blockchain

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"fmt"
	"math/big"
)

var curve = elliptic.P256()

// Wallet represents a Cumulus wallet address in the blockchain.
type Wallet ecdsa.PublicKey

// Hash represents a hash of a transaction.
type Hash []byte

// Signature represents a signature of a transaction.
type Signature struct {
	X big.Int
	Y big.Int
}

// New creates a new Wallet backed by a ECC key pair. Uses system entropy.
func NewWallet() (*Wallet, error) {
	k, err := ecdsa.GenerateKey(curve, rand.Reader)
	if err != nil {
		return nil, err
	}
	pk := Wallet(k.PublicKey)
	return &pk, nil
}

func (w *Wallet) String() string {
	return fmt.Sprintf("%x-%x", w.X, w.Y)
}

func (w *Wallet) Marshal() []byte {
	return elliptic.Marshal(curve, w.X, w.Y)
}

func (w *Wallet) Equals(other *Wallet) bool {
	return w.X.Cmp(other.X) == 0 && w.Y.Cmp(other.Y) == 0
}

func Unmarshal(wallet []byte) *Wallet {
	x, y := elliptic.Unmarshal(curve, wallet)
	return &Wallet{curve, x, y}
}
