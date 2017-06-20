package blockchain

// ValidTransaction checks whether a transaction is valid, assuming the
import "crypto/ecdsa"

// TransactionCode is returned from ValidTransaction.
type TransactionCode uint32

// BlockCode is returned from ValidBlock.
type BlockCode uint32

const (
	// ValidTransaction is returned when transaction is valid.
	ValidTransaction TransactionCode = iota
	// NoInputTransaction is returned when transaction has no valid input transaction.
	NoInputTransaction TransactionCode = iota
	// Overspend is returned when transaction outputs exceed transaction inputs.
	Overspend TransactionCode = iota
	// BadSig is returned when the signature verification fails.
	BadSig TransactionCode = iota
	// Respend is returned when inputs have been spent elsewhere in the chain.
	Respend TransactionCode = iota
)

const (
	// ValidBlock is returned when the block is valid.
	ValidBlock BlockCode = iota
	// BadTransaction is returned when the block contains an invalid
	// transaction.
	BadTransaction BlockCode = iota
	// BadBlockNumber is returned when block number is not one greater than
	// previous block.
	BadBlockNumber BlockCode = iota
	// BadHash is returned when the block contains incorrect hash.
	BadHash BlockCode = iota
	// DoubleSpend is returned when two transactions in the block share inputs,
	// but outputs > inputs.
	DoubleSpend BlockCode = iota
)

// ValidTransaction tests whether a transaction valid.
func (bc *BlockChain) ValidTransaction(t *Transaction) (bool, TransactionCode) {

	// Find the transaction input in the chain (by hash)
	var input *Transaction
	input = bc.GetInputTransaction(t)
	if input == nil || HashSum(input) != t.Input.Hash {
		return false, NoInputTransaction
	}

	// Check that output to sender in input is equal to outputs in t
	if !input.InputsEqualOutputs(t) {
		return false, Overspend
	}

	// Verify signature of t
	hash := HashSum(t.TxBody)
	if !ecdsa.Verify(t.Sender.Key(), hash.Marshal(), t.Sig.R, t.Sig.S) {
		return false, BadSig
	}

	// Test if identical transaction already exists in chain.
	end := uint32(len(bc.Blocks))
	start := t.Input.BlockNumber
	if exists, _, _ := bc.ContainsTransaction(t, start, end); exists {
		return false, Respend
	}

	return true, ValidTransaction
}

// ValidBlock checks whether a block is valid.
func (bc *BlockChain) ValidBlock(b *Block) (bool, BlockCode) {
	// Check that block number between 0 and max blocks.
	ix := b.BlockNumber - 1
	if int(ix) > len(bc.Blocks)-1 || ix < 0 {
		return false, BadBlockNumber
	}

	// Check that block number is one greater than last block
	lastBlock := bc.Blocks[ix]
	if lastBlock.BlockNumber != ix {
		return false, BadBlockNumber
	}

	// Verify every Transaction in the block.
	for _, t := range b.Transactions {
		if valid, _ := bc.ValidTransaction(t); !valid {
			return false, BadTransaction
		}
	}

	// Check that hash of last block is correct
	if HashSum(lastBlock) != b.LastBlock {
		return false, BadHash
	}

	// Check for multiple transactions referencing same input transaction.
	for i, trA := range b.Transactions {
		for j, trB := range b.Transactions {
			if (i != j) && (trA.Input.Hash == trB.Input.Hash) {
				inputTr := bc.GetInputTransaction(trA)
				if !inputTr.InputsEqualOutputs(trA, trB) {
					return false, DoubleSpend
				}
			}
		}
	}

	return true, ValidBlock
}
