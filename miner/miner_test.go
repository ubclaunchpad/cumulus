package miner

import (
	"math/big"
	"testing"
	"time"

	"github.com/ubclaunchpad/cumulus/blockchain"
)

func TestMine(t *testing.T) {
	bc, b := blockchain.NewValidChainAndBlock()
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
