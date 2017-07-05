package consensus

import (
	"math/big"

	log "github.com/Sirupsen/logrus"
	"github.com/ubclaunchpad/cumulus/blockchain"
)

// MinedBlockCode is returned from ValidateMinedBlock
type MinedBlockCode uint32

const (
	// ValidNewBlock is returned when a mined block is valid
	ValidNewBlock MinedBlockCode = iota
	// BadBlock is returned when an invalid block was mined
	BadBlock MinedBlockCode = iota
	// BadNonce is returned when the hash of the block is not less than the
	// target
	BadNonce MinedBlockCode = iota
	// BadTarget is returned when the target that was used while mining is
	// not equal to the current network target
	BadTarget MinedBlockCode = iota
	// BadBlockReward is returned when the block's rewards is not equal to the
	// current reward of the network
	BadBlockReward MinedBlockCode = iota
	// BadCloudBase is returned when the address of the recipient of the
	// CloudBase transaction is not valid
	BadCloudBase MinedBlockCode = iota
)

const (
	// blockRewardHalvingRate is the number of blocks that need to be mined
	// before the blockReward is halved
	blockRewardHalvingRate int = 210000
)

var (
	// BlockReward is the current reward for mining a block
	BlockReward uint64 = 25
	// CurrentDifficulty is the current hashing difficulty of the network
	CurrentDifficulty = blockchain.MinTarget
)

// HalveReward halves the current blockReward if the size of the BlockChain is a
// multiple of the blockRewardHalvingRate
func HalveReward(bc *blockchain.BlockChain) {
	if len(bc.Blocks)%blockRewardHalvingRate == 0 {
		BlockReward /= 2
	}
}

// CurrentTarget returns the current target based on the CurrentDifficulty
func CurrentTarget() blockchain.Hash {
	return blockchain.BigIntToHash(
		new(big.Int).Div(
			blockchain.MaxTarget,
			CurrentDifficulty,
		),
	)
}

// ValidMinedBlock validates that the mined block conforms to the
// consensus rules of Cumulus
func ValidMinedBlock(
	cb blockchain.Address,
	bc *blockchain.BlockChain,
	b *blockchain.Block) (bool, MinedBlockCode) {

	// Check if the block is valid
	if valid, code := bc.ValidBlock(b); !valid {
		log.Errorf("Invalid block, BlockCode: %d", code)
		return false, BadBlock
	}

	// Check if the CloudBase transaction reward is equal to the network's
	// current block reward
	if b.GetCloudBaseTransaction().Outputs[0].Amount != BlockReward {
		log.Error("Invalid Block Reward")
		return false, BadBlockReward
	}

	// Check if the CloudBase transaction recipient is equal to the address
	// of the miner that mined the block
	if b.GetCloudBaseTransaction().Outputs[0].Recipient != cb {
		log.Error("Invalid CloudBase address")
		return false, BadCloudBase
	}

	// Check if the block's target is equal to the network's current target
	if b.Target != CurrentTarget() {
		log.Error("Invalid Target")
		return false, BadTarget
	}

	// Verify proof of work
	if !blockchain.HashSum(b).LessThan(b.Target) {
		log.Error("Invalid Nonce, no proof of work")
		return false, BadNonce
	}

	return true, ValidNewBlock
}
