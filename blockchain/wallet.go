package blockchain

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	crand "crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"math/big"

	log "github.com/Sirupsen/logrus"
	c "github.com/ubclaunchpad/cumulus/common/constants"
	"github.com/ubclaunchpad/cumulus/moj"
)

const (
	// CoordLen is the length in bytes of coordinates with our ECC curve.
	CoordLen = 32
	// AddrLen is the length in bytes of addresses.
	AddrLen = 2 * CoordLen
	// ReprLen is the length in bytes of an address checksum.
	ReprLen = 40
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

// Account represents a wallet that we have the ability to sign for.
type Account interface {
	Public() Address
	Sign(digest Hash, random io.Reader) (Signature, error)
}

// Wallet is an account that can sign and hold a balance.
type Wallet struct {
	*ecdsa.PrivateKey
	PendingTxns []*Transaction
	Balance     uint64
}

// Key retreives the underlying private key from a wallet.
func (w *Wallet) key() *ecdsa.PrivateKey {
	return w.PrivateKey
}

// Public returns the public key as byte array, or address, of the wallet.
func (w *Wallet) Public() Address {
	return Address{X: w.PrivateKey.PublicKey.X, Y: w.PrivateKey.PublicKey.Y}
}

// Sign returns a signature of the digest.
func (w *Wallet) Sign(digest Hash, random io.Reader) (Signature, error) {
	r, s, err := ecdsa.Sign(random, w.key(), digest.Marshal())
	return Signature{R: r, S: s}, err
}

// UnmarshalJSON unmarshals the given byte slice into the given wallet, and
// returns an error if one occurs.
func (w *Wallet) UnmarshalJSON(walletBytes []byte) error {
	var walletParams map[string]interface{}
	dec := json.NewDecoder(bytes.NewReader(walletBytes))
	dec.UseNumber()
	if err := dec.Decode(&walletParams); err != nil {
		return err
	}

	// Initialize private key to avoid SIGSEGV
	key, err := ecdsa.GenerateKey(curve, crand.Reader)
	if err != nil {
		return err
	}
	w.PrivateKey = key

	// Get pending transactions
	txnBytes, err := json.Marshal(walletParams["PendingTxns"])
	if err != nil {
		return err
	}
	txnDecoder := json.NewDecoder(bytes.NewReader(txnBytes))
	txnDecoder.UseNumber()
	if err := txnDecoder.Decode(&w.PendingTxns); err != nil {
		return err
	}

	// Get balance
	balanceBytes, err := json.Marshal(walletParams["Balance"])
	if err != nil {
		return err
	}
	if err := json.Unmarshal(balanceBytes, &w.Balance); err != nil {
		return err
	}

	// Get private/public keys
	if err := w.decodeBigInt(walletParams["X"], w.PrivateKey.PublicKey.X); err != nil {
		return err
	}
	if err := w.decodeBigInt(walletParams["Y"], w.PrivateKey.PublicKey.Y); err != nil {
		return err
	}
	if err := w.decodeBigInt(walletParams["D"], w.PrivateKey.D); err != nil {
		return err
	}

	// Add elliptic curve
	w.PrivateKey.PublicKey.Curve = curve
	return nil
}

// decodeBigInt is a helper function for wallet.UnmarshalJSON. It attempts to
// set target to the big.Int decoded from the given param. Returns an error if
// one occurs.
func (w *Wallet) decodeBigInt(param interface{}, target *big.Int) error {
	intBytes, err := json.Marshal(param)
	if err != nil {
		return err
	}
	dec := json.NewDecoder(bytes.NewReader(intBytes))
	dec.UseNumber()
	return dec.Decode(target)
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

// NewWallet produces a new Wallet that can sign transactions and has a
// public Address.
func NewWallet() *Wallet {
	priv, _ := ecdsa.GenerateKey(curve, crand.Reader)
	return &Wallet{
		PrivateKey: priv,
		Balance:    0,
	}
}

// SetAllPending appends transactions to the pending set of transactions.
func (w *Wallet) SetAllPending(txns []*Transaction) {
	for _, t := range txns {
		w.SetPending(t)
	}
}

// SetPending appends one transaction to the pending set of transactions
// if the wallet's effective balance is high enough to accomodate.
func (w *Wallet) SetPending(txn *Transaction) error {
	bal := w.GetEffectiveBalance()
	spend := txn.GetTotalOutput()
	if bal >= spend {
		w.PendingTxns = append(w.PendingTxns, txn)
	} else {
		return fmt.Errorf("Wallet balance is too low %v < %v", bal, spend)
	}
	return nil
}

// DropAllPending drops pending transactions if they appear in txns. Returns an
// error if any of the given transactions are not in the blockchain.
func (w *Wallet) DropAllPending(txns []*Transaction, bc *BlockChain) error {
	for _, t := range txns {
		if p, i := w.IsPending(t); p {
			w.DropPending(i)
		}
	}
	return nil
}

// DropPending drops a single pending transaction by index in the pending list.
func (w *Wallet) DropPending(i int) {
	if i < len(w.PendingTxns) && i >= 0 {
		log.Info("Pending transaction verified")
		w.PendingTxns = append(w.PendingTxns[:i], w.PendingTxns[i+1:]...)
	}
}

// IsPending returns true if the transaction exists in the pending list.
// If true, it also returns the integer index of the transaction.
func (w *Wallet) IsPending(txn *Transaction) (bool, int) {
	for i, t := range w.PendingTxns {
		if t.Equal(txn) {
			return true, i
		}
	}
	return false, -1
}

// GetEffectiveBalance returns the wallet balance less the sum of the pending
// transactions in the wallet.
func (w *Wallet) GetEffectiveBalance() uint64 {
	r := w.Balance
	for _, t := range w.PendingTxns {
		r -= t.GetTotalOutput() - t.GetTotalOutputFor(t.Sender.Repr())
	}
	return r
}

// Update updates the wallet's balance and set of pending transactions based
// on the transaction information in the given block. Returns an error if any
// of the transactions in the given block cannot be found in the blockchain.
func (w *Wallet) Update(block *Block, bc *BlockChain) error {
	// Update wallet balance with any transactions with intputs to or outputs
	// from the given wallet.
	totalOutput := block.GetTotalOutputFor(w.Public().Repr())
	totalInput, err := block.GetTotalInputFrom(w.Public().Repr(), bc)
	if err != nil {
		return err
	}
	w.Balance += totalOutput - totalInput

	// Update pending transactions
	txns := block.GetTransactionsFrom(w.Public().Repr())
	if err := w.DropAllPending(txns, bc); err != nil {
		return err
	}
	return nil
}

// Refresh sets the wallet's balance and  updates it's set of pending transactions
// based on the transaction information in the given blockchain. Returns an
// error if any of the transactions in the blockchain cannot be found.
func (w *Wallet) Refresh(bc *BlockChain) error {
	w.Balance = uint64(0)
	for _, b := range bc.Blocks {
		if err := w.Update(b, bc); err != nil {
			return err
		}
	}
	return nil
}
