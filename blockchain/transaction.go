package blockchain

import (
	"encoding/binary"
	"io"
	"math"

	"github.com/ubclaunchpad/cumulus/common/util"
	"gopkg.in/fatih/set.v0"
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
	Inputs  []TxHashPointer
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
	for _, in := range tb.Inputs {
		buf = append(buf, in.Marshal()...)
	}
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

// GetTotalOutput sums the output amounts from the transaction.
func (t *Transaction) GetTotalOutput() uint64 {
	result := uint64(0)
	for _, out := range t.Outputs {
		result += out.Amount
	}
	return result
}

// GetTotalOutputFor sums the outputs referenced to a specific recipient.
// recipient is an address checksum hex string.
func (t *Transaction) GetTotalOutputFor(recipient string) uint64 {
	result := uint64(0)
	for _, out := range t.Outputs {
		if out.Recipient == recipient {
			result += out.Amount
		}
	}
	return result
}

// GetTotalInput sums the input amounts from the transaction.
// Requires the blockchain for lookups.
func (t *Transaction) GetTotalInput(bc *BlockChain) (uint64, error) {

	result := uint64(0)
	// This is a bit crazy; filter all input transactions
	// by this senders address and sum the outputs.
	inputs, err := bc.GetAllInputs(t)
	if err != nil {
		return 0, err
	}
	for _, in := range inputs {
		result += in.GetTotalOutputFor(t.Sender.Repr())
	}
	return result, nil
}

// GetBlockRange returns the start and end block indexes for the inputs
// to a transaction.
func (bc *BlockChain) GetBlockRange(t *Transaction) (uint32, uint32) {
	min := uint32(math.MaxUint32)
	max := uint32(0)
	for _, in := range t.Inputs {
		if in.BlockNumber < min {
			min = in.BlockNumber
		}
		if in.BlockNumber > max {
			max = in.BlockNumber
		}
	}
	return min, max
}

// InputsIntersect returns true if the inputs of t intersect with those of other.
func (t *Transaction) InputsIntersect(other *Transaction) bool {
	return !t.InputIntersection(other).IsEmpty()
}

// InputIntersection returns the intersection of the inputs of t and other.
func (t *Transaction) InputIntersection(other *Transaction) set.Interface {
	return set.Intersection(t.InputSet(), other.InputSet())
}

// InputSet returns the transaction inputs as a set object.
func (t *Transaction) InputSet() *set.Set {
	a := make([]interface{}, len(t.Inputs))
	for i, v := range t.Inputs {
		a[i] = v
	}
	return set.New(a...)
}

// InputsSpentElsewhere returns true if inputs purported to be only spent
// on transaction t have been spent elsewhere after block index `start`.
func (t *Transaction) InputsSpentElsewhere(bc *BlockChain, start uint32) bool {
	// Get the set of inputs for t.
	inSet := t.InputSet()

	// Look at each transaction in the chain from start on.
	for _, b := range bc.Blocks[start:] {
		for _, txn := range b.Transactions {

			// If the inputs to t intersect with the inputs to txn...
			if !set.Intersection(inSet, txn.InputSet()).IsEmpty() {

				// ... and the sender is the same, then we have a respend.
				if txn.Sender.Repr() == t.Sender.Repr() {
					return true
				}
			}
		}
	}

	// If we made it through all the transactions, without finding
	// inputs respent anywhere, then we're good.
	return false
}
