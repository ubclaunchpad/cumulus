package miner

import (
	"math/big"
	"testing"
	"time"

	"github.com/ubclaunchpad/cumulus/blockchain"
)

func TestValidMine(t *testing.T) {
	bc, b := blockchain.NewValidChainAndBlock()
	tempMinDifficulty := blockchain.MinDifficulty
	blockchain.MinDifficulty = new(big.Int).Sub(
		blockchain.BigExp(2, 256),
		big.NewInt(1),
	)

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

	if !verifyProofOfWork(b) {
		t.Fail()
	}
}
