package blockchain

// ValidTransaction checks whether a transaction is valid, assuming the
import (
	"crypto/ecdsa"
	"math/big"
	"time"
)

// TransactionCode is returned from ValidTransaction.
type TransactionCode uint32

// CloudBaseTransactionCode is returned from ValidCloudBaseTransaction
type CloudBaseTransactionCode uint32

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
	// NilTransaction is returned when the transaction pointer is nil
	NilTransaction TransactionCode = iota
)

const (
	// ValidCloudBaseTransaction is returned when a transaction is a valid
	// CloudBase transaction.
	ValidCloudBaseTransaction CloudBaseTransactionCode = iota
	// BadCloudBaseSender is returned when the sender address in the CloudBase
	// transaction is not a NilAddr.
	BadCloudBaseSender CloudBaseTransactionCode = iota
	// BadCloudBaseInput is returned when all the fields inf the  CloudBase
	// transaction input are not equal to 0.
	BadCloudBaseInput CloudBaseTransactionCode = iota
	// BadCloudBaseOutput is returned when the CloudBase transaction output is
	// invalid.
	BadCloudBaseOutput CloudBaseTransactionCode = iota
	// BadCloudBaseSig is returned when the CloudBase transaction signature is
	// not equal to NilSig.
	BadCloudBaseSig CloudBaseTransactionCode = iota
	// NilCloudBaseTransaction is returned when the CloudBase transaction
	// pointer is nil
	NilCloudBaseTransaction CloudBaseTransactionCode = iota
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
	// BadCloudBaseTransaction is returned when a block does not have a
	// CloudBase transaction as the first transaction in its list of
	// transactions.
	BadCloudBaseTransaction BlockCode = iota
	// NilBlock is returned when the block pointer is nil
	NilBlock BlockCode = iota
)

var (
	// MaxTarget is the minimum difficulty
	MaxTarget = new(big.Int).Sub(BigExp(2, 232), big.NewInt(1))
	// MinTarget is the maximum difficulty value
	MinTarget = big.NewInt(1)
)

// ValidTransaction tests whether a transaction valid.
func (bc *BlockChain) ValidTransaction(t *Transaction) (bool, TransactionCode) {

	// Check if the transaction is equal to nil
	if t == nil {
		return false, NilTransaction
	}

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

// ValidCloudBase returns true if a transaction is a valid CloudBase transaction
// and false otherwise
func ValidCloudBase(t *Transaction) (bool, CloudBaseTransactionCode) {

	// Check if the CloudBase transaction is equal to nil
	if t == nil {
		return false, NilCloudBaseTransaction
	}

	// Check that the sender address is nil
	if t.Sender != NilAddr {
		return false, BadCloudBaseSender
	}

	// Check that the input is 0
	if t.TxBody.Input.BlockNumber != 0 ||
		t.TxBody.Input.Hash != NilHash ||
		t.Input.Index != 0 {
		return false, BadCloudBaseInput
	}

	// Check that the output is set and that the reward is != 0
	if len(t.Outputs) == 0 ||
		len(t.Outputs) > 1 ||
		t.Outputs[0].Amount == 0 ||
		t.Outputs[0].Recipient == NilAddr {
		return false, BadCloudBaseOutput
	}

	// Asser that the signature is equal to nil
	if t.Sig != NilSig {
		return false, BadCloudBaseSig
	}

	return true, ValidCloudBaseTransaction
}

// ValidBlock checks whether a block is valid.
func (bc *BlockChain) ValidBlock(b *Block) (bool, BlockCode) {

	// Check if the block is equal to nil
	if b == nil {
		return false, NilBlock
	}

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

	// Check that the first transaction is a CloudBase transaction
	if valid, _ := ValidCloudBase(b.Transactions[0]); !valid {
		return false, BadCloudBaseTransaction
	}

	// Verify every Transaction in the block.
	for _, t := range b.Transactions[1:] {
		if valid, _ := bc.ValidTransaction(t); !valid {
			return false, BadTransaction
		}
	}

	// Check that the target is within the min and max difficulty levels
	target := HashToBigInt(b.Target)
	if target.Cmp(MaxTarget) == 1 || target.Cmp(MinTarget) == -1 {
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
