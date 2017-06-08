package blockchain

import (
	"testing"
)

func TestSetMiningHeader(t *testing.T) {
	var mh MiningHeader
	bc, b := newValidChainAndBlock()
	mh.SetMiningHeader(b.LastBlock, bc.Head, HashToCompact(MaxDifficulty))
}

func TestComputeHash(t *testing.T) {
	var version uint32 = 1
	lastBlockHash := HexStringToHash("00000000000008a3a41b85b8b29ad444def299fee21793cd8b9e567eab02cd81")
	rootHash := HexStringToHash("2b12fcf1b09288fcaff797d71e950e71ae42b91e8bdb2304758dfcffc2b620e3")
	var time uint32 = 1305998791
	var target uint32 = 440711666
	var nonce uint32 = 2504433986
	mh := MiningHeader{version, lastBlockHash, rootHash, time, target, nonce}
	hash := mh.DoubleHashSum()
	verifiedHashResult := HexStringToHash("00000000000000001e8d6829a8a21adc5d38d0a473b144b6765798e61f98bd1d")
	if !CompareTo(hash, verifiedHashResult, EqualTo) {
		t.Fail()
	}
}
