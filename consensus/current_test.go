package consensus

import (
	"testing"

	"github.com/ubclaunchpad/cumulus/blockchain"
	c "github.com/ubclaunchpad/cumulus/common/constants"
)

func TestCurrentTarget(t *testing.T) {
	if blockchain.HashToBigInt(CurrentTarget()).Cmp(c.MaxTarget) != 0 {
		t.Fail()
	}
}

func TestCurrentBlockReward(t *testing.T) {
	bc, _ := blockchain.NewValidBlockChainFixture()
	if CurrentBlockReward(bc) != StartingBlockReward {
		t.Fail()
	}
	for i := len(bc.Blocks); i < blockRewardHalvingRate; i++ {
		bc.AppendBlock(new(blockchain.Block))
	}
	if CurrentBlockReward(bc) != StartingBlockReward/2 {
		t.Fail()
	}
	for i := 0; i < blockRewardHalvingRate; i++ {
		bc.AppendBlock(new(blockchain.Block))
	}
	if CurrentBlockReward(bc) != StartingBlockReward/4 {
		t.Fail()
	}
}
