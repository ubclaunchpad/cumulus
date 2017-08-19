package miner

import (
	"math"
	"sync"

	"github.com/ubclaunchpad/cumulus/blockchain"
	"github.com/ubclaunchpad/cumulus/common/util"
	"github.com/ubclaunchpad/cumulus/consensus"
)

// MinerState represents the state of the miner
type MinerState int

const (
	// Paused represents the MinerState where the miner is not running but the
	// previously running mining job can be resumed or stopped.
	Paused = iota
	// Stopped represents the MinerState where the miner is not mining anything.
	Stopped
	// Running represents the MinerState where the miner is actively mining.
	Running
)

const (
	// MiningSuccessful is returned when the miner mines a block.
	MiningSuccessful = iota
	// MiningNeverStarted is returned when the block header is invalid.
	MiningNeverStarted
	// MiningHalted is returned when the app halts the miner.
	MiningHalted
)

var (
	// checkState signals to the miner to check for a new mining state when true.
	checkState bool
	// checkStateLock is a read/write lock to change the checkState flag.
	checkStateLock = &sync.RWMutex{}
	// minerState represents the state of the miner at any given time
	minerState MinerState
	// minerStateLock is a read/write lock to check the minerState variable
	minerStateLock = &sync.RWMutex{}
	// stop signals to the miner to abort the current mining job immediately.
	stop = make(chan bool)
	// resume signals to the miner that it can continue mining from its previous
	// state.
	resume = make(chan bool)
	// pause signals to the miner to pause mining and wait for a stop or resume
	// signal.
	pause = make(chan bool)
)

// MiningResult contains the result of the mining operation.
type MiningResult struct {
	Complete bool
	Info     int
}

// Mine continuously increases the nonce and tries to verify the proof of work
// until the puzzle is solved.
func Mine(b *blockchain.Block) *MiningResult {
	setStateChanged(false)
	setState(Running)

	miningHalted := &MiningResult{
		Complete: false,
		Info:     MiningHalted,
	}

	for !VerifyProofOfWork(b) {
		// Check if we should keep mining.
		if stateChanged() {
			select {
			case <-pause:
				setState(Paused)
				select {
				case <-resume:
					setState(Running)
				case <-stop:
					setState(Stopped)
					return miningHalted
				case <-pause:
					panic("Miner already paused")
				}
			case <-stop:
				setState(Stopped)
				return miningHalted
			case <-resume:
				panic("Miner already running")
			}
		}

		// Check if we should reset the nonce.
		if b.Nonce == math.MaxUint64 {
			b.Nonce = 0
		}

		// Timestamp and increase the nonce.
		b.Time = util.UnixNow()
		b.Nonce++
	}

	setState(Stopped)
	return &MiningResult{
		Complete: true,
		Info:     MiningSuccessful,
	}
}

// StopMining causes the miner to abort the current mining job immediately.
func StopMining() {
	if State() == Running {
		setStateChanged(true)
	}
	stop <- true
}

// PauseIfRunning pauses the current mining job if it is current running. Returns
// true if the miner was running and false otherwise.
func PauseIfRunning() bool {
	minerStateLock.RLock()
	defer minerStateLock.RUnlock()

	if minerState == Running {
		checkStateLock.Lock()
		checkState = true
		checkStateLock.Unlock()
		pause <- true
		return true
	}
	return false
}

// ResumeMining causes the miner to continue mining from a paused state.
func ResumeMining() {
	resume <- true
}

// setStateChanged synchronously sets the checkState variable to the given value.
func setStateChanged(check bool) {
	checkStateLock.Lock()
	defer checkStateLock.Unlock()
	checkState = check
}

// stateChanged synchronously returns wheter or not the miner state has changed
// since it was last checked by the miner.
func stateChanged() bool {
	checkStateLock.RLock()
	defer checkStateLock.RUnlock()
	return checkState
}

// setState synchronously sets the current state of the miner to the given state.
func setState(state MinerState) {
	minerStateLock.Lock()
	defer minerStateLock.Unlock()
	minerState = state
}

// State synchronously returns the current state of the miner.
func State() MinerState {
	minerStateLock.RLock()
	defer minerStateLock.RUnlock()
	return minerState
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
		Amount:    consensus.CurrentBlockReward(bc),
		Recipient: cb.Repr(),
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
