package blockchain

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	crand "crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"math/big"

	c "github.com/ubclaunchpad/cumulus/common/constants"
	"github.com/ubclaunchpad/cumulus/moj"
)

const (
	// CoordLen is the length in bytes of coordinates with our ECC curve.
	CoordLen = 32
	// AddrLen is the length in bytes of addresses.
	AddrLen = 2 * CoordLen
	// SigLen is the length in bytes of signatures.
	SigLen = AddrLen
	// AddressVersion is the version of the address shortening protocol.
	AddressVersion = 0
)

var (
	// The curve we use for our ECC crypto.
	curve = elliptic.P256()
	// NilSig is a signature representing a failed Sign operation
	NilSig = Signature{c.Big0, c.Big0}
	// NilAddr is an address representing no address
	NilAddr = Address{c.Big0, c.Big0}
)

// Address represents a wallet that can be a recipient in a transaction.
type Address struct {
	X, Y *big.Int
}

// Repr returns a string representation of the address. We follow
// ethereums protocol, replacing Keccak hash with SHA256.
// Where pr is the private key,
//    	A(pr) = SHA256(ECDSAPUBLICKEY(pr))[96:255],
// Resources:
// http://gavwood.com/paper.pdf (fig 213)
func (a Address) Repr() string {
	// 1. Concatenate X and Y and version the result.
	concat := a.Marshal()
	prefix := append([]byte{AddressVersion}, concat...)

	// 2. Perform SHA-256 on result.
	hash := sha256.Sum256(prefix) // 256 bit

	return hex.EncodeToString(hash[96/8 : 256/8])
}

// Emoji returns the users address as a string of emojis.
func (a Address) Emoji() string {
	result, _ := moj.EncodeHex(a.Repr())
	return result
}

// Marshal converts an Address to a byte slice.
func (a Address) Marshal() []byte {
	buf := make([]byte, 0, AddrLen)
	xBytes := a.X.Bytes()
	yBytes := a.Y.Bytes()

	if len(xBytes) < CoordLen {
		for i := len(xBytes); i < CoordLen; i++ {
			xBytes = append(xBytes, 0)
		}
	}

	if len(yBytes) < CoordLen {
		for i := len(yBytes); i < CoordLen; i++ {
			yBytes = append(yBytes, 0)
		}
	}

	buf = append(buf, xBytes...)
	buf = append(buf, yBytes...)
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
	rBytes := s.R.Bytes()
	sBytes := s.S.Bytes()

	if len(rBytes) < CoordLen {
		for i := len(rBytes); i < CoordLen; i++ {
			rBytes = append(rBytes, 0)
		}
	}

	if len(sBytes) < CoordLen {
		for i := len(sBytes); i < CoordLen; i++ {
			rBytes = append(sBytes, 0)
		}
	}

	buf = append(buf, rBytes...)
	buf = append(buf, sBytes...)
	return buf
}

// NewWallet produces a new Wallet that can sign transactionsand has a
// public Address.
func NewWallet() Wallet {
	priv, _ := ecdsa.GenerateKey(curve, crand.Reader)
	return (*wallet)(priv)
}
