package app

import "github.com/ubclaunchpad/cumulus/blockchain"

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
var BlockWorkQueue = make(chan BlockWork, BlockQueueSize)

// TransactionWorkQueue is a queue of transactions to process.
var TransactionWorkQueue = make(chan TransactionWork, TransactionQueueSize)

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
