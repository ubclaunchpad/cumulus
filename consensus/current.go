package consensus

import (
	"math"
	"math/big"

	"github.com/ubclaunchpad/cumulus/blockchain"
	c "github.com/ubclaunchpad/cumulus/common/constants"
)

const (
	// CoinValue is the transaction amount that represents one Cumulus coin
	CoinValue uint64 = 2 << 32
	// StartingBlockReward is the mining reward that the blockchain will begin
	// with.
	StartingBlockReward uint64 = 25 * CoinValue
	// blockRewardHalvingRate is the number of blocks that need to be mined
	// before the blockReward is halved
	blockRewardHalvingRate int = 210000
)

var (
	// CurrentDifficulty is the current hashing difficulty of the network
	CurrentDifficulty = c.MinTarget
)

// CurrentBlockReward determines the current block reward using the
// the length of the blockchain
func CurrentBlockReward(bc *blockchain.BlockChain) uint64 {
	timesHalved := float64((len(bc.Blocks) / blockRewardHalvingRate))
	return StartingBlockReward / uint64(math.Pow(float64(2), timesHalved))
}

// CurrentTarget returns the current target based on the CurrentDifficulty
func CurrentTarget() blockchain.Hash {
	return blockchain.BigIntToHash(
		new(big.Int).Div(
			c.MaxTarget,
			CurrentDifficulty,
		),
	)
}
