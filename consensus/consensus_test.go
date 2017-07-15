package consensus

import (
	"testing"

	"math/rand"

	"github.com/ubclaunchpad/cumulus/blockchain"
	c "github.com/ubclaunchpad/cumulus/common/constants"
	"github.com/ubclaunchpad/cumulus/common/util"
)

func TestHalveReward(t *testing.T) {
	bc, _ := blockchain.NewValidTestChainAndBlock()
	tempBlockReward := BlockReward

	for i := 0; i < blockRewardHalvingRate-2; i++ {
		HalveReward(bc)
		if BlockReward != tempBlockReward {
			BlockReward = tempBlockReward
			t.Fail()
		}
		bc.AppendBlock(new(blockchain.Block))
	}

	HalveReward(bc)
	if BlockReward != tempBlockReward/2 {
		BlockReward = tempBlockReward
		t.Fail()
	}
	BlockReward = tempBlockReward
}

func TestCurrentTarget(t *testing.T) {
	if blockchain.HashToBigInt(CurrentTarget()).Cmp(blockchain.MaxTarget) != 0 {
		t.Fail()
	}
}

func TestValidMinedBlockBadBlock(t *testing.T) {
	w := blockchain.NewWallet()
	valid, code := ValidMinedBlock(w.Public(), nil, nil)

	if valid {
		t.Fail()
	}

	if code != BadBlock {
		t.Fail()
	}
}

func TestValidMinedBlockBadBlockReward(t *testing.T) {
	bc, b, a := newValidBlockChainAndCloudBaseBlock()

	var r uint64
	for r = RandomUint64(); r == BlockReward; r = RandomUint64() {
	}

	b.Transactions[0].Outputs[0].Amount = r

	valid, code := ValidMinedBlock(a, bc, b)

	if valid {
		t.Fail()
	}

	if code != BadBlockReward {
		t.Fail()
	}

}

func TestValidMinedBlockBadTarget(t *testing.T) {
	bc, b, a := newValidBlockChainAndCloudBaseBlock()
	b.Target = blockchain.NewValidTestTarget()

	valid, code := ValidMinedBlock(a, bc, b)

	if valid {
		t.Fail()
	}

	if code != BadTarget {
		t.Fail()
	}
}

func TestValidMinedBlockBadNonce(t *testing.T) {
	bc, b, a := newValidBlockChainAndCloudBaseBlock()
	b.Target = blockchain.BigIntToHash(c.Big1)
	tempCurrentDifficulty := CurrentDifficulty
	CurrentDifficulty = blockchain.MaxTarget

	valid, code := ValidMinedBlock(a, bc, b)

	CurrentDifficulty = tempCurrentDifficulty

	if valid {
		t.Fail()
	}

	if code != BadNonce {
		t.Fail()
	}
}

func TestValidMinedBlockBadCloudBase(t *testing.T) {
	bc, b, a := newValidBlockChainAndCloudBaseBlock()
	b.Transactions[0].Outputs[0].Recipient = blockchain.NewWallet().Public()
	valid, code := ValidMinedBlock(a, bc, b)

	if valid {
		t.Fail()
	}

	if code != BadCloudBase {
		t.Fail()
	}
}

func TestValidMinedBlock(t *testing.T) {
	tempMaxTarget := blockchain.MaxTarget
	tempCurrentDifficulty := CurrentDifficulty

	blockchain.MaxTarget = c.MaxUint256
	CurrentDifficulty = blockchain.MinTarget

	bc, b, a := newValidBlockChainAndCloudBaseBlock()

	valid, code := ValidMinedBlock(a, bc, b)

	blockchain.MaxTarget = tempMaxTarget
	CurrentDifficulty = tempCurrentDifficulty

	if !valid {
		t.Fail()
	}

	if code != ValidNewBlock {
		t.Fail()
	}
}

func newValidBlockChainAndCloudBaseBlock() (
	*blockchain.BlockChain,
	*blockchain.Block,
	blockchain.Address) {
	bc, _ := blockchain.NewValidBlockChainFixture()
	cbTx, a := blockchain.NewValidCloudBaseTestTransaction()
	bcSize := uint32(len(bc.Blocks))
	b := &blockchain.Block{
		BlockHeader: blockchain.BlockHeader{
			BlockNumber: bcSize,
			LastBlock:   blockchain.HashSum(bc.Blocks[bcSize-1]),
			Target:      CurrentTarget(),
			Time:        util.UnixNow(),
			Nonce:       0,
		},
		Transactions: make([]*blockchain.Transaction, 1),
	}
	b.Transactions[0] = cbTx
	return bc, b, a
}

// Creates a random uint64 value
func RandomUint64() uint64 {
	return uint64(rand.Uint32())<<32 + uint64(rand.Uint32())
}
