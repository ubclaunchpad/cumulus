package app

import (
	"github.com/ubclaunchpad/cumulus/blockchain"
	"github.com/ubclaunchpad/cumulus/miner"

	log "github.com/Sirupsen/logrus"
)

const (
	// BlockQueueSize is the size of the BlockQueue channel.
	BlockQueueSize = 100
	// TransactionQueueSize is the size of the BlockQueue channel.
	TransactionQueueSize = 100
)

// Responder is used to handle requests who require a response.
type Responder interface {
	Send(ok bool)
	Lock()
	Unlock()
}

// BlockWorkQueue is a queue of blocks to process.
var blockWorkQueue = make(chan BlockWork, BlockQueueSize)

// TransactionWorkQueue is a queue of transactions to process.
var transactionWorkQueue = make(chan TransactionWork, TransactionQueueSize)

// QuitChan kills the app worker.
var quitChan = make(chan bool)

// TransactionWork holds a new transaction job, and a Responder for
// sending results.
type TransactionWork struct {
	*blockchain.Transaction
	Responder
}

// BlockWork holds a new block job,  and a Responder for
// sending results.
type BlockWork struct {
	*blockchain.Block
	Responder
}

// HandleWork continually collects new work from existing work channels.
func HandleWork(app *App) {
	log.Debug("Worker waiting for work.")
	for {
		select {
		case work := <-transactionWorkQueue:
			HandleTransaction(app, work)
		case work := <-blockWorkQueue:
			HandleBlock(app, work)
		case <-quitChan:
			return
		}
	}
}

// HandleTransaction handles new instance of TransactionWork.
func HandleTransaction(app *App, work TransactionWork) {
	validTransaction := tpool.Set(work.Transaction, chain)

	// Respond to the request if a response method was provided.
	if work.Responder != nil {
		work.Responder.Lock()
		defer work.Responder.Unlock()
		work.Responder.Send(validTransaction)
	}
}

// HandleBlock handles new instance of BlockWork.
func HandleBlock(app *App, work BlockWork) {
	validBlock := tpool.Update(work.Block, chain)

	if validBlock {
		// Append to the chain before requesting
		// the next block so that the block
		// numbers make sense.
		chain.AppendBlock(work.Block)
		address := app.CurrentUser.Wallet.Public()
		blk := tpool.NextBlock(chain, address, app.CurrentUser.BlockSize)
		if miner.IsMining() {
			miner.RestartMiner(chain, blk)
		}
	}

	// Respond to the request if a response method was provided.
	if work.Responder != nil {
		work.Responder.Lock()
		defer work.Responder.Unlock()
		work.Responder.Send(validBlock)
	}
}
