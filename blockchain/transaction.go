package blockchain

import "encoding/binary"

const (
	// TxHashPointerLen is the length in bytes of a hash pointer.
	TxHashPointerLen = 32/8 + HashLen
	// TxOutputLen is the length in bytes of a transaction output.
	TxOutputLen = 64/8 + AddrLen
)

// TxHashPointer is a reference to a transaction on the blockchain.
type TxHashPointer struct {
	BlockNumber uint32
	Hash        Hash
}

// Marshal converts a TxHashPointer to a byte slice
func (thp TxHashPointer) Marshal() []byte {
	buf := make([]byte, 4, TxHashPointerLen)
	binary.LittleEndian.PutUint32(buf, thp.BlockNumber)
	buf = append(buf, thp.Hash.Marshal()...)
	return buf
}

// TxOutput defines an output to a transaction
type TxOutput struct {
	Amount    uint64
	Recipient Address
}

// Marshal converts a TxOutput to a byte slice
func (to TxOutput) Marshal() []byte {
	buf := make([]byte, 8, TxOutputLen)
	binary.LittleEndian.PutUint64(buf, to.Amount)
	buf = append(buf, to.Recipient.Marshal()...)
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
	return AddrLen + TxHashPointerLen + len(tb.Outputs)*TxOutputLen
}

// Marshal converts a TxBody to a byte slice
func (tb TxBody) Marshal() []byte {
	buf := make([]byte, 0, tb.Len())
	buf = append(buf, tb.Sender.Marshal()...)
	buf = append(buf, tb.Input.Marshal()...)
	for _, out := range tb.Outputs {
		buf = append(buf, out.Marshal()...)
	}
	return buf
}

// Transaction contains a TxBody and a signature verifying it
type Transaction struct {
	TxBody
	Sig Signature
}

// Len returns the length in bytes of a transaction
func (t *Transaction) Len() int {
	return t.TxBody.Len() + SigLen
}

// Marshal converts a Transaction to a byte slice
func (t *Transaction) Marshal() []byte {
	buf := make([]byte, 0, t.Len())
	buf = append(buf, t.TxBody.Marshal()...)
	buf = append(buf, t.Sig.Marshal()...)
	return buf
}
