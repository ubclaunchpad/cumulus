package miner

import (
	"testing"
	"time"

	"github.com/ubclaunchpad/cumulus/common/constants"

	"github.com/stretchr/testify/assert"
	"github.com/ubclaunchpad/cumulus/blockchain"
	c "github.com/ubclaunchpad/cumulus/common/constants"
	"github.com/ubclaunchpad/cumulus/common/util"
	"github.com/ubclaunchpad/cumulus/consensus"
)

func TestMine(t *testing.T) {
	_, b := blockchain.NewValidTestChainAndBlock()
	tempMaxTarget := c.MaxTarget

	// Set min difficulty to be equal to the target so that the block validation
	// passes
	c.MaxTarget = c.MaxUint256

	// Set target to be as easy as possible so that we find a hash
	// below the target straight away (2**256 - 1)
	b.Target = blockchain.BigIntToHash(c.MaxTarget)
	b.Time = util.UnixNow()
	mineResult := Mine(b)
	c.MaxTarget = tempMaxTarget

	assert.True(t, mineResult.Complete)
	assert.Equal(t, mineResult.Info, MiningSuccessful)
}

func TestMineHaltMiner(t *testing.T) {
	_, b := blockchain.NewValidTestChainAndBlock()

	// Set target to be as hard as possible so that we stall.
	b.Target = blockchain.BigIntToHash(c.MinTarget)
	b.Time = util.UnixNow()

	// Use a thread to stop the miner a few moments after starting.
	go func() {
		time.Sleep(50 * time.Millisecond)
		StopMining()
	}()

	// Start the miner.
	mineResult := Mine(b)

	assert.False(t, mineResult.Complete)
	assert.Equal(t, mineResult.Info, MiningHalted)
}

func TestCloudBase(t *testing.T) {
	bc, _ := blockchain.NewValidBlockChainFixture()
	w := blockchain.NewWallet()
	bcSize := uint32(len(bc.Blocks))
	b := &blockchain.Block{
		BlockHeader: blockchain.BlockHeader{
			BlockNumber: bcSize,
			LastBlock:   blockchain.HashSum(bc.Blocks[bcSize-1]),
			Target:      consensus.CurrentTarget(),
			Time:        util.UnixNow(),
			Nonce:       0,
		},
		Transactions: make([]*blockchain.Transaction, 0),
	}

	CloudBase(b, bc, w.Public())

	if valid, _ := consensus.VerifyBlock(bc, b); !valid {
		t.Fail()
	}

	if b.Transactions[0].Outputs[0].Amount != consensus.CurrentBlockReward(bc) {
		t.Fail()
	}

	if b.Transactions[0].Outputs[0].Recipient != w.Public().Repr() {
		t.Fail()
	}
}

func TestVerifyProofOfWork(t *testing.T) {
	_, b := blockchain.NewValidTestChainAndBlock()
	b.Target = blockchain.BigIntToHash(
		c.MaxUint256,
	)

	if !VerifyProofOfWork(b) {
		t.Fail()
	}
}

func TestStopPauseMining(t *testing.T) {
	b := blockchain.NewTestBlock()
	b.Target = blockchain.BigIntToHash(constants.MinTarget)

	go Mine(b)
	time.Sleep(time.Second / 2)
	assert.Equal(t, int(State()), int(Running))
	assert.True(t, PauseIfRunning())
	assert.Equal(t, int(State()), int(Paused))
	ResumeMining()
	time.Sleep(time.Second / 2)
	assert.Equal(t, int(State()), int(Running))
	StopMining()
	time.Sleep(time.Second / 2)
	assert.Equal(t, int(State()), int(Stopped))
	consensus.CurrentDifficulty = constants.MinTarget
}
