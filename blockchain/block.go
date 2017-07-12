package blockchain

// BlockHeader contains metadata about a block
import (
	"encoding/binary"
	"encoding/gob"
	"fmt"
	"io"
)

// BlockHeader contains metadata about a block
type BlockHeader struct {
	// BlockNumber is the position of the block within the blockchain
	BlockNumber uint32
	// LastBlock is the hash of the previous block
	LastBlock Hash
	// Target is the current target
	Target Hash
	// Time is represented as the number of seconds elapsed
	// since January 1, 1970 UTC. It increments every second when mining.
	Time uint32
	// Nonce starts at 0 and increments by 1 for every hash when mining
	Nonce uint64
	// ExtraData is an extra field that can be filled with arbitrary data to
	// be stored in the block
	ExtraData []byte
}

// Marshal converts a BlockHeader to a byte slice
func (bh *BlockHeader) Marshal() []byte {
	var buf []byte

	tempBufBlockNumber := make([]byte, 4)
	binary.LittleEndian.PutUint32(tempBufBlockNumber, bh.BlockNumber)

	tempBufTime := make([]byte, 4)
	binary.LittleEndian.PutUint32(tempBufTime, bh.Time)

	tempBufNonce := make([]byte, 8)
	binary.LittleEndian.PutUint64(tempBufNonce, bh.Nonce)

	buf = append(buf, tempBufBlockNumber...)
	buf = append(buf, bh.LastBlock.Marshal()...)
	buf = append(buf, bh.Target.Marshal()...)
	buf = append(buf, tempBufTime...)
	buf = append(buf, tempBufNonce...)
	buf = append(buf, bh.ExtraData...)

	return buf
}

// Len returns the length in bytes of the BlockHeader.
func (bh *BlockHeader) Len() int {
	return len(bh.Marshal())
}

// Block represents a block in the blockchain. Contains transactions and header metadata.
type Block struct {
	BlockHeader
	Transactions []*Transaction
}

// Len returns the length in bytes of the Block.
func (b *Block) Len() int {
	return len(b.Marshal())
}

// Marshal converts a Block to a byte slice.
func (b *Block) Marshal() []byte {
	var buf []byte
	buf = append(buf, b.BlockHeader.Marshal()...)
	for _, t := range b.Transactions {
		buf = append(buf, t.Marshal()...)
	}
	return buf
}

// Encode writes the marshalled block to the given io.Writer
func (b *Block) Encode(w io.Writer) {
	err := gob.NewEncoder(w).Encode(b)
	if err != nil {
		fmt.Println(err.Error())
	}
}

// DecodeBlock reads the marshalled block from the given io.Reader and populates b
func DecodeBlock(r io.Reader) *Block {
	var b Block
	gob.NewDecoder(r).Decode(&b)
	return &b
}

// ContainsTransaction returns true and the transaction itself if the Block
// contains the transaction.
func (b *Block) ContainsTransaction(t *Transaction) (bool, uint32) {
	for i, tr := range b.Transactions {
		if HashSum(t) == HashSum(tr) {
			return true, uint32(i)
		}
	}
	return false, 0
}

// GetCloudBaseTransaction returns the CloudBase transaction within a block
func (b *Block) GetCloudBaseTransaction() *Transaction {
	return b.Transactions[0]
}
