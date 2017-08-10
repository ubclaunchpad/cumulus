package blockchain

import (
	"encoding/binary"
	"io"

	"github.com/ubclaunchpad/cumulus/common/util"
)

// TxHashPointer is a reference to a transaction on the blockchain.
type TxHashPointer struct {
	BlockNumber uint32
	Hash        Hash
	Index       uint32
}

// Marshal converts a TxHashPointer to a byte slice
func (thp TxHashPointer) Marshal() []byte {
	var buf []byte
	buf = util.AppendUint32(buf, thp.BlockNumber)
	buf = append(buf, thp.Hash.Marshal()...)
	buf = util.AppendUint32(buf, thp.Index)
	return buf
}

// TxOutput defines an output to a transaction
type TxOutput struct {
	Amount    uint64
	Recipient string
}

// Marshal converts a TxOutput to a byte slice
func (to TxOutput) Marshal() []byte {
	buf := make([]byte, 8)
	binary.LittleEndian.PutUint64(buf, to.Amount)
	buf = append(buf, []byte(to.Recipient)...)
	return buf
}

// TxBody contains all relevant information about a transaction
type TxBody struct {
	Sender  Address
	Input   TxHashPointer
	Outputs []TxOutput
}

// Len returns the length of a transaction body
func (tb TxBody) Len() int {
	return len(tb.Marshal())
}

// Marshal converts a TxBody to a byte slice
func (tb TxBody) Marshal() []byte {
	var buf []byte
	buf = append(buf, tb.Sender.Marshal()...)
	buf = append(buf, tb.Input.Marshal()...)
	for _, out := range tb.Outputs {
		buf = append(buf, out.Marshal()...)
	}
	return buf
}

// Sign returns a signed Transaction from a TxBody
func (tb TxBody) Sign(w Wallet, r io.Reader) (*Transaction, error) {
	digest := HashSum(tb)
	sig, err := w.Sign(digest, r)
	return &Transaction{tb, sig}, err
}

// Transaction contains a TxBody and a signature verifying it
type Transaction struct {
	TxBody
	Sig Signature
}

// Len returns the length in bytes of a transaction
func (t *Transaction) Len() int {
	return len(t.Marshal())
}

// Marshal converts a Transaction to a byte slice
func (t *Transaction) Marshal() []byte {
	var buf []byte
	buf = append(buf, t.TxBody.Marshal()...)
	buf = append(buf, t.Sig.Marshal()...)
	return buf
}

// InputsEqualOutputs returns true if t.Inputs == other.Outputs, as well
// as the difference between the two (outputs - inputs).
func (t *Transaction) InputsEqualOutputs(other ...*Transaction) bool {
	var inAmount uint64
	for _, otherTransaction := range other {
		for _, output := range otherTransaction.Outputs {
			inAmount += output.Amount
		}
	}

	var outAmount uint64
	for _, output := range t.Outputs {
		outAmount += output.Amount
	}

	return (int(outAmount) - int(inAmount)) == 0
}
