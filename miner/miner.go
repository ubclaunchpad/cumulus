package miner

import (
	"math"
	"time"

	"github.com/ubclaunchpad/cumulus/blockchain"
)

// Mine continuously increases the nonce and tries to verify the proof of work
// until the puzzle is solved
func Mine(bc *blockchain.BlockChain, b *blockchain.Block) bool {
	if valid, _ := bc.ValidBlock(b); !valid {
		return false
	}

	for !verifyProofOfWork(b) {
		if b.Nonce == math.MaxUint64 {
			b.Nonce = 0
		}
		b.Time = uint32(time.Now().Unix())
		b.Nonce++
	}
	return true
}

// verifyProofOfWork computes the hash of the MiningHeader and returns true if
// the result is less than the target
func verifyProofOfWork(b *blockchain.Block) bool {
	return blockchain.HashSum(b).LessThan(b.Target)
}
