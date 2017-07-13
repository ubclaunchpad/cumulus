package miner

import (
	"testing"
	"time"

	"github.com/ubclaunchpad/cumulus/blockchain"
	c "github.com/ubclaunchpad/cumulus/common/constants"
	"github.com/ubclaunchpad/cumulus/consensus"
)

func TestMine(t *testing.T) {
	bc, b := blockchain.NewValidTestChainAndBlock()
	tempMaxTarget := blockchain.MaxTarget

	// Set min difficulty to be equal to the target so that the block validation
	// passes
	blockchain.MaxTarget = c.MaxUint256

	// Set target to be as easy as possible so that we find a hash
	// below the target straight away (2**256 - 1)
	b.Target = blockchain.BigIntToHash(blockchain.MaxTarget)
	b.Time = uint32(time.Now().Unix())
	mineResult := Mine(bc, b)
	blockchain.MaxTarget = tempMaxTarget

	if !mineResult {
		t.Fail()
	}
}

func TestMineBadBlock(t *testing.T) {
	bc, _ := blockchain.NewValidTestChainAndBlock()
	mineResult := Mine(bc, nil)

	if mineResult {
		t.Fail()
	}
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
		c.MaxUint256,
	)

	if !VerifyProofOfWork(b) {
		t.Fail()
	}
}
