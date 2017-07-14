package blockchain

// ValidTransaction checks whether a transaction is valid, assuming the
import (
	"crypto/ecdsa"

	log "github.com/Sirupsen/logrus"
	c "github.com/ubclaunchpad/cumulus/common/constants"
	"github.com/ubclaunchpad/cumulus/common/util"
)

// TransactionCode is returned from ValidTransaction.
type TransactionCode uint32

// CloudBaseTransactionCode is returned from ValidCloudBaseTransaction.
type CloudBaseTransactionCode uint32

// GenesisBlockCode is returned from ValidGenesisBlock.
type GenesisBlockCode uint32

// BlockCode is returned from ValidBlock.
type BlockCode uint32

const (
	// ValidTransaction is returned when transaction is valid.
	ValidTransaction TransactionCode = iota
	// NoInputTransaction is returned when transaction has no valid input transaction.
	NoInputTransaction
	// Overspend is returned when transaction outputs exceed transaction inputs.
	Overspend
	// BadSig is returned when the signature verification fails.
	BadSig
	// Respend is returned when inputs have been spent elsewhere in the chain.
	Respend
	// NilTransaction is returned when the transaction pointer is nil.
	NilTransaction
)

const (
	// ValidCloudBaseTransaction is returned when a transaction is a valid
	// CloudBase transaction.
	ValidCloudBaseTransaction CloudBaseTransactionCode = iota
	// BadCloudBaseSender is returned when the sender address in the CloudBase
	// transaction is not a NilAddr.
	BadCloudBaseSender
	// BadCloudBaseInput is returned when all the fields inf the  CloudBase
	// transaction input are not equal to 0.
	BadCloudBaseInput
	// BadCloudBaseOutput is returned when the CloudBase transaction output is
	// invalid.
	BadCloudBaseOutput
	// BadCloudBaseSig is returned when the CloudBase transaction signature is
	// not equal to NilSig.
	BadCloudBaseSig
	// NilCloudBaseTransaction is returned when the CloudBase transaction
	// pointer is nil
	NilCloudBaseTransaction
)

const (
	// ValidGenesisBlock is returned when a block is a valid genesis block.
	ValidGenesisBlock GenesisBlockCode = iota
	// BadGenesisLastBlock is returned when the LastBlock of the genesis block
	// is not equal to 0.
	BadGenesisLastBlock
	// BadGenesisTransactions is returned when the genesis block does not contain
	// exactly 1 transaction, the first CloudBase transaction.
	BadGenesisTransactions
	// BadGenesisCloudBaseTransaction is returned when the transaction in the
	// genesis block is not a valid CloudBase transaction.
	BadGenesisCloudBaseTransaction
	// BadGenesisBlockNumber is returned when the block number in the genesis
	// block is not equal to 0.
	BadGenesisBlockNumber
	// BadGenesisTarget is returned when the genesis block's target is invalid.
	BadGenesisTarget
	// BadGenesisTime is returned when the genesis block's time is invalid.
	BadGenesisTime
	// NilGenesisBlock is returned when the genesis block is equal to nil.
	NilGenesisBlock
)

const (
	// ValidBlock is returned when the block is valid.
	ValidBlock BlockCode = iota
	// BadTransaction is returned when the block contains an invalid
	// transaction.
	BadTransaction
	// BadTime is returned when the block contains an invalid time.
	BadTime
	// BadTarget is returned when the block contains an invalid target.
	BadTarget
	// BadBlockNumber is returned when block number is not one greater than
	// previous block.
	BadBlockNumber
	// BadHash is returned when the block contains incorrect hash.
	BadHash
	// DoubleSpend is returned when two transactions in the block share inputs,
	// but outputs > inputs.
	DoubleSpend
	// BadCloudBaseTransaction is returned when a block does not have a
	// CloudBase transaction as the first transaction in its list of
	// transactions.
	BadCloudBaseTransaction
	// BadGenesisBlock is returned if the block is a genesis block and is
	// invalid.
	BadGenesisBlock
	// NilBlock is returned when the block pointer is nil.
	NilBlock
)

var (
	// MaxTarget is the minimum difficulty
	MaxTarget = util.BigSub(util.BigExp(2, 232), c.Big1)
	// MinTarget is the maximum difficulty value
	MinTarget = c.Big1
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

// ValidGenesisBlock checks whether a block is a valid genesis block.
func (bc *BlockChain) ValidGenesisBlock(gb *Block) (bool, GenesisBlockCode) {

	// Check if the genesis block is equal to nil.
	if gb == nil {
		return false, NilGenesisBlock
	}

	// Check if the genesis block's block number is equal to 0.
	if gb.BlockHeader.BlockNumber != 0 ||
		bc.Blocks[0] != gb {
		return false, BadGenesisBlockNumber
	}

	// Check if the genesis block's last block hash is equal to 0.
	if HashToBigInt(gb.BlockHeader.LastBlock).Cmp(c.Big0) != 0 {
		return false, BadGenesisLastBlock
	}

	// Check if the size of the transaction list is equal to 1.
	if len(gb.Transactions) != 1 {
		return false, BadGenesisTransactions
	}

	// Check if the transaction is a valid cloud base transaction.
	if valid, code := ValidCloudBase(gb.Transactions[0]); !valid {
		log.Errorf("Invalid CloudBase, CloudBaseTransactionCode: %d", code)
		return false, BadGenesisCloudBaseTransaction
	}

	// Check that the target is within the min and max difficulty levels.
	target := HashToBigInt(gb.Target)
	if target.Cmp(MaxTarget) == 1 || target.Cmp(MinTarget) == -1 {
		return false, BadGenesisTarget
	}

	// Check that time is not greater than current time or equal to 0.
	if uint32(gb.Time) == 0 {
		return false, BadGenesisTime
	}

	return true, ValidGenesisBlock
}

// ValidBlock checks whether a block is valid.
func (bc *BlockChain) ValidBlock(b *Block) (bool, BlockCode) {

	// Check if the block is equal to nil.
	if b == nil {
		return false, NilBlock
	}

	// Check if the block is the genesis block.
	if b.BlockHeader.BlockNumber == 0 || bc.Blocks[0] == b {
		if valid, code := bc.ValidGenesisBlock(b); !valid {
			log.Errorf("Invalid GenesisBlock, GenesisBlockCode: %d", code)
			return false, BadGenesisBlock
		}
		return true, ValidBlock
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
	if valid, code := ValidCloudBase(b.Transactions[0]); !valid {
		log.Errorf("Invalid CloudBase, CloudBaseTransactionCode: %d", code)
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
	if uint32(b.Time) == 0 {
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
