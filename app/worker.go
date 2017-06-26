package app

import log "github.com/Sirupsen/logrus"

// nWorkers is how many workers this node has.
const nWorkers = 10

// Workers is a list of workers.
var workers [nWorkers]*GenericWorker

// Worker is an interface for the basic app worker tasks.
type Worker interface {
	HandleTransaction(work TransactionWork)
	HandleBlock(work BlockWork)
}

// GenericWorker implements the basic worker.
type GenericWorker struct {
	ID int
}

// NewWorker returns a new TransactionWorker object.
func NewWorker(id int) GenericWorker {
	return GenericWorker{
		ID: id,
	}
}

// Start continually collects new work from existing work channels.
func (w GenericWorker) Start() {
	go func() {
		for {
			// Wait for work.
			log.WithFields(log.Fields{
				"id": w.ID,
			}).Debug("Worker waiting for work.")
			select {
			case work := <-TransactionWorkQueue:
				log.WithFields(log.Fields{
					"id": w.ID,
				}).Debug("Worker handling new transaction work.")
				w.HandleTransaction(work)
			case work := <-BlockWorkQueue:
				log.WithFields(log.Fields{
					"id": w.ID,
				}).Debug("Worker handling new block work.")
				w.HandleBlock(work)
			case <-QuitChan:
				log.WithFields(log.Fields{
					"id": w.ID,
				}).Debug("Worker quitting.")
				return
			}
		}
	}()
}

// NB: We're currently imposing a validation layer at the app level
// using methods like SetUnsafe to get the transaction into the
// pool, etc. Figure out a nicer way to do this, ie: where *should*
// we validate?

// HandleTransaction handles new instance of TransactionWork.
func (w *GenericWorker) HandleTransaction(work TransactionWork) {
	ok, _ := chain.ValidTransaction(work.Transaction)
	if ok {
		tpool.SetUnsafe(work.Transaction)
	}

	// Respond to the request if a response method was provided.
	if work.Responder != nil {
		work.Responder.Lock()
		work.Responder.Send(ok)
		work.Responder.Unlock()
	}
}

// HandleBlock handles TransactionWork.
func (w *GenericWorker) HandleBlock(work BlockWork) {
	ok, _ := chain.ValidBlock(work.Block)
	if ok {
		chain.AppendBlock(work.Block, work.Miner)
	}

	// Respond to the request if a response method was provided.
	if work.Responder != nil {
		work.Responder.Lock()
		work.Responder.Send(ok)
		work.Responder.Unlock()
	}
}
