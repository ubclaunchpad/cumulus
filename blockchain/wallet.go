package blockchain

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"io"
	"math/big"
)

const (
	// CoordLen is the length in bytes of coordinates with our ECC curve.
	CoordLen = 32
	// AddrLen is the length in bytes of addresses.
	AddrLen = 2 * CoordLen
	// SigLen is the length in bytes of signatures.
	SigLen = AddrLen
)

var (
	// The curve we use for our ECC crypto.
	curve = elliptic.P256()
	// NilSig is a signature representing a failed Sign operation
	NilSig = Signature{big.NewInt(0), big.NewInt(0)}
	// NilAddr is an address representing no address
	NilAddr = Address{}
)

// Address represents a wallet that can be a recipient in a transaction.
type Address struct {
	X, Y *big.Int
}

// Marshal converts an Address to a byte slice.
func (a Address) Marshal() []byte {
	buf := make([]byte, AddrLen)
	buf = append(buf, a.X.Bytes()...)
	buf = append(buf, a.Y.Bytes()...)
	return buf
}

// Key returns the ECDSA public key representation of the address.
func (a Address) Key() *ecdsa.PublicKey {
	return &ecdsa.PublicKey{
		Curve: curve,
		X:     a.X,
		Y:     a.Y,
	}
}

// Wallet represents a wallet that we have the ability to sign for.
type Wallet interface {
	Public() Address
	Sign(digest Hash, random io.Reader) (Signature, error)
}

// Internal representation of a wallet.
type wallet ecdsa.PrivateKey

// Key retreives the underlying private key from a wallet.
func (w *wallet) key() *ecdsa.PrivateKey {
	return (*ecdsa.PrivateKey)(w)
}

// Public returns the public key as byte array, or address, of the wallet.
func (w *wallet) Public() Address {
	return Address{X: w.PublicKey.X, Y: w.PublicKey.Y}
}

// Sign returns a signature of the digest.
func (w *wallet) Sign(digest Hash, random io.Reader) (Signature, error) {
	r, s, err := ecdsa.Sign(random, w.key(), digest.Marshal())
	return Signature{R: r, S: s}, err
}

// Signature represents a signature of a transaction.
type Signature struct {
	R *big.Int
	S *big.Int
}

// Marshal converts a signature to a byte slice. Should be 64 bytes long.
func (s *Signature) Marshal() []byte {
	buf := make([]byte, 0, SigLen)
	buf = append(buf, s.R.Bytes()...)
	buf = append(buf, s.S.Bytes()...)
	return buf
}
