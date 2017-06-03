package blockchain

// ValidTransaction checks whether a transaction is valid, assuming the
import "crypto/ecdsa"

// TransactionCode is returned from ValidTransaction.
type TransactionCode uint32

// BlockCode is returned from ValidBlock.
type BlockCode uint32

const (
	// ValidTransaction is returned when transaction is valid.
	ValidTransaction TransactionCode = 0
	// NoInputTransaction is returned when transaction has no valid input transaction.
	NoInputTransaction TransactionCode = 1
	// Overspend is returned when transaction outputs exceed transaction inputs.
	Overspend TransactionCode = 2
	// BadSig is returned when the signature verification fails.
	BadSig TransactionCode = 3
	// Respend is returned when inputs have been spent elsewhere in the chain.
	Respend TransactionCode = 4
)

const (
	// ValidBlock is returned when the block is valid.
	ValidBlock BlockCode = 0
	// BadTransaction is returned when the block contains an invalid
	// transaction.
	BadTransaction BlockCode = 1
	// BadBlockNumber is returned when block number is not one greater than
	// previous block.
	BadBlockNumber BlockCode = 2
	// BadHash is returned when the block contains incorrect hash.
	BadHash BlockCode = 3
	// DoubleSpend is returned when two transactions in the block share inputs,
	// but outputs > inputs.
	DoubleSpend BlockCode = 4
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
	if _, diff := input.InputsEqualOutputs(t); diff != 0 {
		return false, Overspend
	}

	// Verify signature of t
	hash := HashSum(t.TxBody)
	if !ecdsa.Verify(t.Sender.Key(), hash.Marshal(), t.Sig.R, t.Sig.S) {
		return false, BadSig
	}

	// Test if identical transaction already exists in chain.
	endChain := uint32(len(bc.Blocks))
	for i := t.Input.BlockNumber; i < endChain; i++ {
		if exists, _ := bc.Blocks[i].ContainsTransaction(t); exists {
			return false, Respend
		}
	}

	return true, ValidTransaction
}

// ValidBlock checks whether a block is valid.
func (bc *BlockChain) ValidBlock(b *Block) (bool, BlockCode) {
	// Check that block number is one greater than last block
	lastBlock := bc.Blocks[b.BlockNumber-1]
	if lastBlock.BlockNumber != b.BlockNumber-1 {
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
				if _, diff := inputTr.InputsEqualOutputs(trA, trB); diff < 0 {
					return false, DoubleSpend
				}
			}
		}
	}

	return true, ValidBlock
}