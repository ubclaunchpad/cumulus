package miner

import (
	"math"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/ubclaunchpad/cumulus/blockchain"
	"github.com/ubclaunchpad/cumulus/consensus"
)

// Mine continuously increases the nonce and tries to verify the proof of work
// until the puzzle is solved
func Mine(bc *blockchain.BlockChain, b *blockchain.Block) bool {
	if valid, _ := bc.ValidBlock(b); !valid {
		log.Error("Invalid block")
		return false
	}

	for !VerifyProofOfWork(b) {
		if b.Nonce == math.MaxUint64 {
			b.Nonce = 0
		}
		b.Time = uint32(time.Now().Unix())
		b.Nonce++
	}
	return true
}

// CloudBase prepends the cloudbase transaction to the front of a list of
// transactions in a block that is to be added to the blockchain
func CloudBase(
	b *blockchain.Block,
	bc *blockchain.BlockChain,
	cb blockchain.Address) *blockchain.Block {
	// Create a cloudbase transaction by setting all inputs to 0
	cbInput := blockchain.TxHashPointer{
		BlockNumber: 0,
		Hash:        blockchain.NilHash,
		Index:       0,
	}
	// Set the transaction amount to the BlockReward
	// TODO: Add transaction fees
	cbReward := blockchain.TxOutput{
		Amount:    consensus.BlockReward,
		Recipient: cb,
	}
	cbTxBody := blockchain.TxBody{
		Sender:  blockchain.NilAddr,
		Input:   cbInput,
		Outputs: []blockchain.TxOutput{cbReward},
	}
	cbTx := blockchain.Transaction{
		TxBody: cbTxBody,
		Sig:    blockchain.NilSig,
	}

	b.Transactions = append([]*blockchain.Transaction{&cbTx}, b.Transactions...)

	// Increment the input index of every transaction that has an input in the
	// new block
	for _, tx := range b.Transactions[1:] {
		if tx.Input.BlockNumber == uint32(len(bc.Blocks)) {
			tx.Input.Index++
		}
	}

	return b
}

// VerifyProofOfWork computes the hash of the MiningHeader and returns true if
// the result is less than the target
func VerifyProofOfWork(b *blockchain.Block) bool {
	return blockchain.HashSum(b).LessThan(b.Target)
}
