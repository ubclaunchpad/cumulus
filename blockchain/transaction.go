package blockchain

import "encoding/binary"

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

// TxBody contains all relevant information about a transaction.
type TxBody struct {
	sender Wallet
	input  TxHashPointer
	output TxHashPointer
}

// Marshal converts a TxBody to a byte slice
func (tb TxBody) Marshal() []byte {
	buf := tb.sender.Marshal()
	buf = append(buf, tb.input.Marshal()...)
	buf = append(buf, tb.output.Marshal()...)
	return buf
}

// Transaction contains a TxBody and a signature verifying it.
type Transaction struct {
	TxBody
	hash Hash
	sig  Signature
}

// Marshal converts a Transaction to a byte slice
func (t *Transaction) Marshal() []byte {
	buf := t.TxBody.Marshal()
	for _, b := range t.hash {
		buf = append(buf, b)
	}
	buf = append(buf, t.sig.X.Bytes()...)
	buf = append(buf, t.sig.Y.Bytes()...)
	return buf
}
