package miner

import (
	"math"
	"sync"

	"github.com/ubclaunchpad/cumulus/blockchain"
	"github.com/ubclaunchpad/cumulus/common/util"
	"github.com/ubclaunchpad/cumulus/consensus"
)

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

// MiningResult contains the result of the mining operation.
type MiningResult struct {
	Complete bool
	Info     int
}

// MinerState represents the state of the miner
type MinerState int

// Miner represents the state of of the current mining job (or lack thereof).
type Miner struct {
	// state represents the state of the miner at any given time
	state MinerState
	// stateLock is a read/write lock to check the state variable
	stateLock *sync.RWMutex
	// stop signals to the miner to abort the current mining job immediately.
	stop chan bool
	// resume signals to the miner that it can continue mining from its previous
	// state.
	resume chan bool
	// pause signals to the miner to pause mining and wait for a stop or resume
	// signal.
	pause chan bool
}

// New returns a new miner.
func New() *Miner {
	return &Miner{
		state:     Stopped,
		stateLock: &sync.RWMutex{},
		stop:      make(chan bool),
		resume:    make(chan bool),
		pause:     make(chan bool),
	}
}

// Mine continuously increases the nonce and tries to verify the proof of work
// until the puzzle is solved.
func (m *Miner) Mine(b *blockchain.Block) *MiningResult {
	m.setState(Running)

	miningHalted := &MiningResult{
		Complete: false,
		Info:     MiningHalted,
	}

	for !m.VerifyProofOfWork(b) {
		// Check if we should keep mining.
		select {
		case <-m.pause:
			m.setState(Paused)
			select {
			case <-m.resume:
				m.setState(Running)
			case <-m.stop:
				m.setState(Stopped)
				return miningHalted
			case <-m.pause:
				panic("Miner already paused")
			}
		case <-m.stop:
			m.setState(Stopped)
			return miningHalted
		case <-m.resume:
			panic("Miner already running")
		default:
			// No state change - keep mining.
		}

		// Check if we should reset the nonce.
		if b.Nonce == math.MaxUint64 {
			b.Nonce = 0
		}

		// Timestamp and increase the nonce.
		b.Time = util.UnixNow()
		b.Nonce++
	}

	m.setState(Stopped)
	return &MiningResult{
		Complete: true,
		Info:     MiningSuccessful,
	}
}

// setState synchronously sets the current state of the miner to the given state.
func (m *Miner) setState(state MinerState) {
	m.stateLock.Lock()
	defer m.stateLock.Unlock()
	m.state = state
}

// StopMining causes the miner to abort the current mining job immediately.
func (m *Miner) StopMining() {
	m.stop <- true
}

// PauseIfRunning pauses the current mining job if it is current running. Returns
// true if the miner was running and false otherwise.
func (m *Miner) PauseIfRunning() bool {
	m.stateLock.RLock()
	defer m.stateLock.RUnlock()
	if m.state == Running {
		m.pause <- true
		return true
	}
	return false
}

// ResumeMining causes the miner to continue mining from a paused state.
func (m *Miner) ResumeMining() {
	m.resume <- true
}

// State synchronously returns the current state of the miner.
func (m *Miner) State() MinerState {
	m.stateLock.RLock()
	defer m.stateLock.RUnlock()
	return m.state
}

// VerifyProofOfWork computes the hash of the MiningHeader and returns true if
// the result is less than the target
func (m *Miner) VerifyProofOfWork(b *blockchain.Block) bool {
	return blockchain.HashSum(b).LessThan(b.Target)
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
		Inputs:  []blockchain.TxHashPointer{cbInput},
		Outputs: []blockchain.TxOutput{cbReward},
	}
	cbTx := blockchain.Transaction{
		TxBody: cbTxBody,
		Sig:    blockchain.NilSig,
	}

	b.Transactions = append([]*blockchain.Transaction{&cbTx}, b.Transactions...)

	return b
}
