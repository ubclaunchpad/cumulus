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
type Address [AddrLen]byte

// Marshal converts an Address to a byte slice.
func (a Address) Marshal() []byte {
	buf := make([]byte, AddrLen)
	for i, b := range a {
		buf[i] = b
	}
	return buf
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
	addr := Address{}
	x := w.PublicKey.X.Bytes()
	y := w.PublicKey.Y.Bytes()

	if len(x) != CoordLen || len(y) != CoordLen {
		// Invalid wallet
		return NilAddr
	}

	for i, b := range x {
		addr[i] = b
	}
	for i, b := range y {
		addr[CoordLen+i] = b
	}

	return addr
}

// Sign returns a signature of the digest.
func (w *wallet) Sign(digest Hash, random io.Reader) (Signature, error) {
	r, s, err := ecdsa.Sign(random, w.key(), digest.Marshal())
	return Signature{r, s}, err
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
