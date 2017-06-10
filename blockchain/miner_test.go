package blockchain

import (
	"fmt"
	"testing"
)

func TestMine(t *testing.T) {
	var mh MiningHeader
	bc, b := newValidChainAndBlock()
	mh.SetMiningHeader(b.LastBlock, bc.Head, MaxDifficulty)
	fmt.Println(mh.Mine())
}
