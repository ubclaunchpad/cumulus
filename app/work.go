package app

import "github.com/ubclaunchpad/cumulus/blockchain"

const (
	// BlockQueueBuffer is the size of the BlockQueue channel.
	BlockQueueBuffer = 100
	// TransactionQueueBuffer is the size of the BlockQueue channel.
	TransactionQueueBuffer = 100
)

// Responder is used to handle requests who require a response.
type Responder interface {
	Send(ok bool)
	Lock()
	Unlock()
}

// BlockWorkQueue is a queue of blocks to process.
var BlockWorkQueue = make(chan BlockWork, BlockQueueBuffer)

// TransactionWorkQueue is a queue of transactions to process.
var TransactionWorkQueue = make(chan TransactionWork, TransactionQueueBuffer)

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
