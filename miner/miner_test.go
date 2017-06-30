package miner

import (
	"math/big"
	"testing"
	"time"

	"github.com/ubclaunchpad/cumulus/blockchain"
)

func TestMine(t *testing.T) {
	bc, b := blockchain.NewValidChainAndBlock()
	tempMinDifficulty := blockchain.MinDifficulty

	// Set min difficulty to be equal to the target so that the block validation
	// passes
	blockchain.MinDifficulty = new(big.Int).Sub(
		blockchain.BigExp(2, 256),
		big.NewInt(1),
	)

	// Set target to be as easy as possible so that we find a hash
	// below the target straight away (2**256 - 1)
	b.Target = blockchain.BigIntToHash(blockchain.MinDifficulty)
	b.Time = uint32(time.Now().Unix())
	mineResult := Mine(bc, b)
	blockchain.MinDifficulty = tempMinDifficulty

	if bc == nil {
		t.Fail()
	}

	if !mineResult {
		t.Fail()
	}
}

func TestVerifyProofOfWork(t *testing.T) {
	_, b := blockchain.NewValidChainAndBlock()
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
