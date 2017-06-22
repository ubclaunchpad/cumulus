package blockchain

// ValidTransaction checks whether a transaction is valid, assuming the
import (
	"crypto/ecdsa"
	"math/big"
	"time"
)

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
	// BadTime is returned when the block contains an invalid time
	BadTime BlockCode = iota
	// BadTarget is returned when the block contains an invalid target
	BadTarget BlockCode = iota
	// BadBlockNumber is returned when block number is not one greater than
	// previous block.
	BadBlockNumber BlockCode = iota
	// BadHash is returned when the block contains incorrect hash.
	BadHash BlockCode = iota
	// DoubleSpend is returned when two transactions in the block share inputs,
	// but outputs > inputs.
	DoubleSpend BlockCode = iota
)

var (
	// MinDifficulty is the minimum difficulty
	MinDifficulty = new(big.Int).Sub(BigExp(2, 232), big.NewInt(1))
	// MaxDifficulty is the maximum difficulty value
	MaxDifficulty = big.NewInt(1)
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

	// Check that the target is within the min and max difficulty levels
	target := HashToBigInt(b.Target)
	if target.Cmp(MinDifficulty) == 1 || target.Cmp(MaxDifficulty) == -1 {
		return false, BadTarget
	}

	// Check that time is not greater than current time or equal to 0
	if uint32(b.Time) == 0 || uint32(b.Time) > uint32(time.Now().Unix()) {
		return false, BadTime
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

// BigExp returns an big int pointer with the result set to base**exp,
// if exp <= 0, the result is 1
func BigExp(base, exp int) *big.Int {
	return new(big.Int).Exp(big.NewInt(int64(base)), big.NewInt(int64(exp)), nil)
}
