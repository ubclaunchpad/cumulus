package miner

import (
	"math/big"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/ubclaunchpad/cumulus/blockchain"
	"github.com/ubclaunchpad/cumulus/consensus"
)

func TestMine(t *testing.T) {
	bc, b := blockchain.NewValidTestChainAndBlock()
	tempMaxTarget := blockchain.MaxTarget

	// Set min difficulty to be equal to the target so that the block validation
	// passes
	blockchain.MaxTarget = new(big.Int).Sub(
		blockchain.BigExp(2, 256),
		big.NewInt(1),
	)

	// Set target to be as easy as possible so that we find a hash
	// below the target straight away (2**256 - 1)
	b.Target = blockchain.BigIntToHash(blockchain.MaxTarget)
	b.Time = uint32(time.Now().Unix())
	mineResult := Mine(bc, b)
	blockchain.MaxTarget = tempMaxTarget

	assert.True(t, mineResult.Complete)
	assert.Equal(t, mineResult.Info, MiningSuccessful)
}

func TestMineBadBlock(t *testing.T) {
	bc, _ := blockchain.NewValidTestChainAndBlock()
	mineResult := Mine(bc, nil)

	assert.False(t, mineResult.Complete)
	assert.Equal(t, mineResult.Info, MiningNeverStarted)
}

func TestMineHaltMiner(t *testing.T) {
	bc, b := blockchain.NewValidTestChainAndBlock()

	// Set target to be as hard as possible so that we stall.
	b.Target = blockchain.BigIntToHash(blockchain.MinTarget)
	b.Time = uint32(time.Now().Unix())

	// Use a thread to stop the miner a few moments after starting.
	go func() {
		time.Sleep(50 * time.Millisecond)
		StopMining()
	}()

	// Start the miner.
	mineResult := Mine(bc, b)

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
			Time:        uint32(time.Now().Unix()),
			Nonce:       0,
		},
		Transactions: make([]*blockchain.Transaction, 0),
	}

	CloudBase(b, bc, w.Public())

	if valid, _ := bc.ValidBlock(b); !valid {
		t.Fail()
	}

	if b.Transactions[0].Outputs[0].Amount != consensus.BlockReward {
		t.Fail()
	}

	if b.Transactions[0].Outputs[0].Recipient != w.Public() {
		t.Fail()
	}
}

func TestVerifyProofOfWork(t *testing.T) {
	_, b := blockchain.NewValidTestChainAndBlock()
	b.Target = blockchain.BigIntToHash(
		new(big.Int).Sub(
			blockchain.BigExp(2, 256),
			big.NewInt(1),
		),
	)

	if !VerifyProofOfWork(b) {
		t.Fail()
	}
}
