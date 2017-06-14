package miner

import (
	"math"
	"testing"
	"time"

	"github.com/ubclaunchpad/cumulus/blockchain"
)

func TestVerifyMiningHeader(t *testing.T) {
	bc, b := blockchain.NewValidChainAndBlock()
	mh := MiningHeader{b.LastBlock, bc.Head, *blockchain.MaxDifficulty, 0, 0}

	if mh.VerifyMiningHeader() != false {
		t.Fail()
	}

	mh = MiningHeader{b.LastBlock, bc.Head, *blockchain.MaxDifficulty, math.MaxUint32, 0}

	if mh.VerifyMiningHeader() != false {
		t.Fail()
	}

	mh = MiningHeader{b.LastBlock, bc.Head, *blockchain.MinHash, uint32(time.Now().Unix()), 0}

	if mh.VerifyMiningHeader() != false {
		t.Fail()
	}

	mh = MiningHeader{b.LastBlock, bc.Head, *blockchain.MaxHash, uint32(time.Now().Unix()), 0}

	if mh.VerifyMiningHeader() != false {
		t.Fail()
	}

}
