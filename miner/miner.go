package miner

import (
	"math"
	"sync"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/ubclaunchpad/cumulus/blockchain"
	"github.com/ubclaunchpad/cumulus/consensus"
)

// CurrentlyMining is a flag to control the miner.
var CurrentlyMining bool

// CurrentlyMiningLock is a read/write lock to change the Mining flag.
var CurrentlyMiningLock sync.RWMutex

const (
	// MiningSuccessful is returned when the miner mines a block.
	MiningSuccessful = iota
	// MiningNeverStarted is returned when the block is invalid.
	MiningNeverStarted = iota
	// MiningHalted is returned when the app halts the miner.
	MiningHalted = iota
)

// MiningResult contains the result of the mining operation.
type MiningResult struct {
	Complete bool
	Info     int
}

// IsMining returns the mining status of the miner.
// Many threads can read this status, only one can write.
func IsMining() bool {
	CurrentlyMiningLock.RLock()
	defer CurrentlyMiningLock.RUnlock()
	return CurrentlyMining
}

// RestartMiner restarts the miner with a new block.
func RestartMiner(bc *blockchain.BlockChain, b *blockchain.Block) {
	StopMiner()
	Mine(bc, b)
}

// Mine continuously increases the nonce and tries to verify the proof of work
// until the puzzle is solved
func Mine(bc *blockchain.BlockChain, b *blockchain.Block) *MiningResult {
	setStart()
	if valid, _ := bc.ValidBlock(b); !valid {
		log.Error("Invalid block")
		return &MiningResult{
			Complete: false,
			Info:     MiningNeverStarted,
		}
	}

	for !VerifyProofOfWork(b) {
		// Check if we should keep mining.
		if !IsMining() {
			return &MiningResult{
				Complete: false,
				Info:     MiningHalted,
			}
		}

		// Check if we should reset the nonce.
		if b.Nonce == math.MaxUint64 {
			b.Nonce = 0
		}

		// Timestamp and increase the nonce.
		b.Time = uint32(time.Now().Unix())
		b.Nonce++
	}

	return &MiningResult{
		Complete: true,
		Info:     MiningSuccessful,
	}
}

func setStart() {
	CurrentlyMiningLock.Lock()
	defer CurrentlyMiningLock.Unlock()
	CurrentlyMining = true
}

// StopMiner stops the miner from mining.
func StopMiner() {
	CurrentlyMiningLock.Lock()
	defer CurrentlyMiningLock.Unlock()
	CurrentlyMining = false
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
