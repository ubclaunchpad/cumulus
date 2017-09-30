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
	if CurrentBlockReward(bc) != blockchain.StartingBlockReward {
		t.Fail()
	}
	for i := len(bc.Blocks); i < blockchain.BlockRewardHalvingRate; i++ {
		bc.AppendBlock(new(blockchain.Block))
	}
	if CurrentBlockReward(bc) != blockchain.StartingBlockReward/2 {
		t.Fail()
	}
	for i := 0; i < blockchain.BlockRewardHalvingRate; i++ {
		bc.AppendBlock(new(blockchain.Block))
	}
	if CurrentBlockReward(bc) != blockchain.StartingBlockReward/4 {
		t.Fail()
	}
}
