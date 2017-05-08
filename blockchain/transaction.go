package blockchain

import (
	"encoding/binary"
)

// TxHashPointer is a reference to a transaction on the blockchain.
type TxHashPointer struct {
	blockNumber uint32
	hash        []byte
}

func (thp TxHashPointer) Marshal() []byte {
	buf := []byte{}
	binary.LittleEndian.PutUint32(buf, thp.blockNumber)
	return append(buf, thp.hash...)
}

// TxBody contains all relevant information about a transaction.
type TxBody struct {
	sender  Wallet
	inputs  []TxHashPointer
	outputs []TxHashPointer
}

func (tb *TxBody) Marshal() []byte {
	buf := tb.sender.Marshal()
	for _, input := range tb.inputs {
		buf = append(buf, input.Marshal()...)
	}
	for _, output := range tb.outputs {
		buf = append(buf, output.Marshal()...)
	}
	return buf
}

// Transaction contains a TxBody and a signature verifying it.
type Transaction struct {
	TxBody
	hash Hash
	sig  Signature
}

func (t *Transaction) Marshal() []byte {
	buf := t.TxBody.Marshal()
	buf = append(buf, t.hash...)
	buf = append(buf, t.sig.X.Bytes()...)
	buf = append(buf, t.sig.Y.Bytes()...)
	return buf
}
