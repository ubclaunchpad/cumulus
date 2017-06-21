package miner

import (
	"math"
	"math/big"
	"testing"
	"time"

	"fmt"

	"github.com/ubclaunchpad/cumulus/blockchain"
)

func TestVerifyMiningHeader(t *testing.T) {
	bc, b := blockchain.NewValidChainAndBlock()

	// Set time to min value, with a valid target
	mh := MiningHeader{b.LastBlock, bc.Head, blockchain.BigIntToHash(BigExp(2, 128)), 0, 0}

	if mh.VerifyMiningHeader() != false {
		t.Fail()
	}

	// Set time to max value, with a valid target
	mh = MiningHeader{b.LastBlock, bc.Head, blockchain.BigIntToHash(BigExp(2, 128)), math.MaxUint32, 0}

	if mh.VerifyMiningHeader() != false {
		t.Fail()
	}

	// Set target to a value larger than min difficulty, with a valid time
	mh = MiningHeader{b.LastBlock, bc.Head, blockchain.BigIntToHash(BigExp(2, 255)), uint32(time.Now().Unix()), 0}

	if mh.VerifyMiningHeader() != false {
		t.Fail()
	}

	// Set target to a value less than max difficulty, with a valid time
	mh = MiningHeader{b.LastBlock, bc.Head, blockchain.BigIntToHash(big.NewInt(0)), uint32(time.Now().Unix()), 0}

	if mh.VerifyMiningHeader() != false {
		t.Fail()
	}
}

func TestVerifyProofOfWork(t *testing.T) {
	bc, b := blockchain.NewValidChainAndBlock()
	mh := new(MiningHeader).SetMiningHeader(b.LastBlock, bc.Head, blockchain.BigIntToHash(new(big.Int).Sub(BigExp(2, 256), big.NewInt(1))))

	// Test mining header with extremely easy target (2**256 - 1)
	if !mh.VerifyProofOfWork() {
		t.Fail()
	}
}

func TestBigExp(t *testing.T) {
	a := big.NewInt(1)
	b := BigExp(0, 0)

	if a.Cmp(b) != 0 {
		t.Fail()
	}

	a = big.NewInt(1)
	b = BigExp(10, -2)
	fmt.Println(b)

	if a.Cmp(b) != 0 {
		t.Fail()
	}

	a = new(big.Int).Exp(big.NewInt(int64(2)), big.NewInt(int64(256)), big.NewInt(0))
	b = BigExp(2, 256)

	if a.Cmp(b) != 0 {
		t.Fail()
	}
}
