package blockchain

// BlockHeader contains metadata about a block
import (
	"bytes"
	"encoding/json"

	"github.com/ubclaunchpad/cumulus/common/util"
)

// DefaultBlockSize is the default block size, can be augmented by the user.
const DefaultBlockSize = 1 << 18

// BlockHeader contains metadata about a block
type BlockHeader struct {
	// BlockNumber is the position of the block within the blockchain
	BlockNumber uint32
	// LastBlock is the hash of the previous block
	LastBlock Hash
	// Target is the current target
	Target Hash
	// Time is represented as the number of seconds elapsed
	// since January 1, 1970 UTC.
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
	buf = util.AppendUint32(buf, bh.BlockNumber)
	buf = append(buf, bh.LastBlock.Marshal()...)
	buf = append(buf, bh.Target.Marshal()...)
	buf = util.AppendUint32(buf, bh.Time)
	buf = util.AppendUint64(buf, bh.Nonce)
	buf = append(buf, bh.ExtraData...)

	return buf
}

// Equal returns true if all the fields (other than ExtraData) in each of
// the BlockHeaders match, and false otherwise.
func (bh *BlockHeader) Equal(otherHeader *BlockHeader) bool {
	return bh.BlockNumber == otherHeader.BlockNumber &&
		bh.LastBlock == otherHeader.LastBlock &&
		bh.Target == otherHeader.Target &&
		bh.Time == otherHeader.Time &&
		bh.Nonce == otherHeader.Nonce
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

// DecodeBlockJSON returns a block read from the given marshalled block, or an
// error if blockBytes cannot be decoded as JSON.
func DecodeBlockJSON(blockBytes []byte) (*Block, error) {
	var b Block
	dec := json.NewDecoder(bytes.NewReader(blockBytes))
	dec.UseNumber()
	err := dec.Decode(&b)
	return &b, err
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
