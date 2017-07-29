package blockchain

import (
	"testing"

	c "github.com/ubclaunchpad/cumulus/common/constants"
)

func TestGenesis(t *testing.T) {
	miner := NewWallet()
	currentTarget := BigIntToHash(c.MaxTarget)
	currentBlockReward := uint64(25)
	gb := Genesis(miner.Public(), currentTarget, currentBlockReward, []byte{})

	// Check if the genesis block is equal to nil
	if gb == nil {
		t.Fail()
	}

	// Check if the genesis block's block number is equal to 0
	if gb.BlockHeader.BlockNumber != 0 {
		t.Fail()
	}

	// Check if the genesis block's last block hash is equal to 0
	if HashToBigInt(gb.BlockHeader.LastBlock).Cmp(c.Big0) != 0 {
		t.Fail()
	}

	// Check iof the size of the transaction list is equal to 1
	if len(gb.Transactions) != 1 {
		t.Fail()
	}
}
