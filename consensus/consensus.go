package consensus

import (
	"crypto/ecdsa"
	"math"
	"reflect"

	"gopkg.in/fatih/set.v0"

	log "github.com/Sirupsen/logrus"
	"github.com/ubclaunchpad/cumulus/blockchain"
	c "github.com/ubclaunchpad/cumulus/common/constants"
)

// VerifyTransaction tests whether a transaction valid.
func VerifyTransaction(bc *blockchain.BlockChain,
	t *blockchain.Transaction) (bool, TransactionCode) {

	// Check if the transaction is equal to nil
	if t == nil {
		return false, NilTransaction
	}

	// Find the transaction input in the chain (by hash)
	inputs, err := bc.GetAllInputs(t)
	if err != nil || len(inputs) == 0 {
		return false, NoInputTransactions
	}

	// Check that output to sender in input is equal to outputs in t
	out := t.GetTotalOutput()
	in, err := t.GetTotalInput(bc)
	if out != in {
		return false, Overspend
	}

	// Verify signature of t.
	hash := blockchain.HashSum(t.TxBody)
	if !ecdsa.Verify(t.Sender.Key(), hash.Marshal(), t.Sig.R, t.Sig.S) {
		return false, BadSig
	}

	// Test if inputs already spent elsewhere.
	start, _ := bc.GetBlockRange(t)
	if t.InputsSpentElsewhere(bc, start) {
		return false, Respend
	}

	return true, ValidTransaction
}

// VerifyCloudBase returns true if a transaction is a valid CloudBase transaction
// and false otherwise
func VerifyCloudBase(bc *blockchain.BlockChain,
	t *blockchain.Transaction) (bool, CloudBaseTransactionCode) {

	// Check if the CloudBase transaction is equal to nil.
	if t == nil {
		return false, NilCloudBaseTransaction
	}

	// Check that the sender address is nil.
	if !reflect.DeepEqual(t.Sender, blockchain.NilAddr) {
		return false, BadCloudBaseSender
	}

	// Check that the input is 0 (only one input to CB).
	input := t.TxBody.Inputs
	if len(input) != 1 ||
		input[0].BlockNumber != 0 ||
		input[0].Hash != blockchain.NilHash ||
		input[0].Index != 0 {
		return false, BadCloudBaseInput
	}

	// Search for the block associated with the CloudBase transaction. If the
	// transaction is not found, then it is to be added to the next block in the
	// blockchain.
	var i int
	for i = 0; i < len(bc.Blocks); i++ {
		if b, _ := bc.Blocks[i].ContainsTransaction(t); b {
			break
		}
	}

	// Determine the reward associated with that specific block.
	timesHalved := float64(((i + 1) / blockchain.BlockRewardHalvingRate))
	reward := blockchain.StartingBlockReward / uint64(math.Pow(float64(2), timesHalved))

	// Check that the output is properly set.
	if len(t.Outputs) != 1 || t.Outputs[0].Recipient == blockchain.NilAddr.Repr() {
		return false, BadCloudBaseOutput
	}

	// Check that the reward is properly set.
	if t.Outputs[0].Amount != reward {
		return false, BadCloudBaseReward
	}

	// Assert that the signature is equal to nil.
	if !reflect.DeepEqual(t.Sig, blockchain.NilSig) {
		return false, BadCloudBaseSig
	}

	return true, ValidCloudBaseTransaction
}

// VerifyGenesisBlock checks whether a block is a valid genesis block.
func VerifyGenesisBlock(bc *blockchain.BlockChain,
	gb *blockchain.Block) (bool, GenesisBlockCode) {

	// Check if the genesis block is equal to nil.
	if gb == nil {
		return false, NilGenesisBlock
	}

	// Check if the genesis block's block number is equal to 0.
	if gb.BlockHeader.BlockNumber != 0 ||
		(len(bc.Blocks) > 0 && bc.Blocks[0] != gb) {
		return false, BadGenesisBlockNumber
	}

	// Check if the genesis block's last block hash is equal to 0.
	if blockchain.HashToBigInt(gb.BlockHeader.LastBlock).Cmp(c.Big0) != 0 {
		return false, BadGenesisLastBlock
	}

	// Check if the size of the transaction list is equal to 1.
	if len(gb.Transactions) != 1 {
		return false, BadGenesisTransactions
	}

	// Check if the transaction is a valid cloud base transaction.
	if valid, code := VerifyCloudBase(bc, gb.Transactions[0]); !valid {
		log.Errorf("Invalid CloudBase, CloudBaseTransactionCode: %d", code)
		return false, BadGenesisCloudBaseTransaction
	}

	// Check that the target is within the min and max difficulty levels.
	// TODO: CurrentTarget() is used because the difficulty is static for v1 of
	// cumulus. Once the difficulty is dynamic, the target would need to be
	// compared to the target calculated for that specific block.
	target := blockchain.HashToBigInt(gb.Target)
	if target.Cmp(c.MaxTarget) == 1 ||
		target.Cmp(c.MinTarget) == -1 ||
		target.Cmp(blockchain.HashToBigInt(CurrentTarget())) != 0 {
		return false, BadGenesisTarget
	}

	// Check that time is not greater than current time or equal to 0.
	// TODO: check if time is the current time
	if gb.Time == 0 {
		return false, BadGenesisTime
	}

	return true, ValidGenesisBlock
}

// VerifyBlock checks whether a block is valid.
func VerifyBlock(bc *blockchain.BlockChain,
	b *blockchain.Block) (bool, BlockCode) {

	// Check if the block is equal to nil.
	if b == nil {
		return false, NilBlock
	}

	// Check if the block is the genesis block.
	if b.BlockHeader.BlockNumber == 0 {
		if valid, code := VerifyGenesisBlock(bc, b); !valid {
			log.Errorf("Invalid GenesisBlock, GenesisBlockCode: %d", code)
			return false, BadGenesisBlock
		}
		return true, ValidBlock
	}

	// Check that block number is valid.
	lastBlockIdx := b.BlockNumber - 1
	if int(lastBlockIdx) != len(bc.Blocks)-1 {
		return false, BadBlockNumber
	}

	// Check that block number is one greater than last block
	lastBlock := bc.Blocks[lastBlockIdx]
	if lastBlock.BlockNumber != lastBlockIdx {
		return false, BadBlockNumber
	}

	// Check that the first transaction is a CloudBase transaction
	if valid, code := VerifyCloudBase(bc, b.GetCloudBaseTransaction()); !valid {
		log.Errorf("Invalid CloudBase, CloudBaseTransactionCode: %d", code)
		return false, BadCloudBaseTransaction
	}

	// Verify every Transaction in the block.
	for _, t := range b.Transactions[1:] {
		if valid, code := VerifyTransaction(bc, t); !valid {
			log.Errorf("Invalid Transaction, TransactionCode: %d", code)
			return false, BadTransaction
		}
	}

	// Check that the target is within the min and max difficulty levels and
	// that the target is correct.
	// TODO: CurrentTarget() is used because the difficulty is static for v1 of
	// cumulus. Once the difficulty is dynamic, the target would need to be
	// compared to the target calculated for that specific block.
	target := blockchain.HashToBigInt(b.Target)
	if target.Cmp(blockchain.HashToBigInt(CurrentTarget())) != 0 {
		return false, BadTarget
	}

	// Check that time is not greater than current time or equal to 0
	// TODO: check if time is the current time
	if b.Time == 0 {
		return false, BadTime
	}

	// Check that hash of last block is correct
	if blockchain.HashSum(lastBlock) != b.LastBlock {
		return false, BadHash
	}

	// Verify proof of work
	if !blockchain.HashSum(b).LessThan(b.Target) {
		return false, BadNonce
	}

	// Check for multiple transactions referencing same input transaction,
	// where the sender is the same.
	inputSets := map[blockchain.Address]*set.Set{}
	for _, txn := range b.Transactions {

		// We'll inspect these inputs next.
		nextInputSet := txn.InputSet()

		// If the sender already exists in the map, check for
		// a non-empty intersection in inputs.
		if inSet, ok := inputSets[txn.Sender]; ok {
			if !set.Intersection(inSet, nextInputSet).IsEmpty() {
				return false, DoubleSpend
			}

			// No intersection, but more inputs to add to sender.
			inputSets[txn.Sender].Merge(nextInputSet)

		} else {
			// First time seeing sender, give them inputs.
			inputSets[txn.Sender] = nextInputSet
		}
	}

	return true, ValidBlock
}
