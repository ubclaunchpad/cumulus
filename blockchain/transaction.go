package blockchain

import (
	"crypto/sha256"
	"encoding/binary"
)

// TxHashPointer is a reference to a transaction on the blockchain.
type TxHashPointer struct {
	blockNumber uint32
	hash        Hash
}

// Marshal converts a TxHashPointer to a byte slice
func (thp TxHashPointer) Marshal() []byte {
	buf := []byte{}
	binary.LittleEndian.PutUint32(buf, thp.blockNumber)
	for _, b := range thp.hash {
		buf = append(buf, b)
	}
	return buf
}

// TxOutput defines an output to a transaction
type TxOutput struct {
	amount    uint64
	recipient Wallet
}

// Marshal converts a TxOutput to a byte slice
func (to TxOutput) Marshal() []byte {
	buf := []byte{}
	binary.LittleEndian.PutUint64(buf, to.amount)
	buf = append(buf, to.recipient.Marshal()...)
	return buf
}

// TxBody contains all relevant information about a transaction
type TxBody struct {
	sender  Wallet
	input   TxHashPointer
	outputs []TxOutput
}

// Marshal converts a TxBody to a byte slice
func (tb TxBody) Marshal() []byte {
	buf := tb.sender.Marshal()
	buf = append(buf, tb.input.Marshal()...)
	for _, out := range tb.outputs {
		buf = append(buf, out.Marshal()...)
	}
	return buf
}

// Transaction contains a TxBody and a signature verifying it
type Transaction struct {
	TxBody
	sig Signature
}

// Marshal converts a Transaction to a byte slice
func (t *Transaction) Marshal() []byte {
	buf := t.TxBody.Marshal()
	buf = append(buf, t.sig.Marshal()...)
	return buf
}

// Hash returns the SHA256 hash of a transaction
func (t *Transaction) Hash() Hash {
	return sha256.Sum256(t.Marshal())
}
