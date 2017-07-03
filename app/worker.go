package app

import log "github.com/Sirupsen/logrus"

// nWorkers is how many workers this node has.
const nWorkers = 10

// Workers is a list of workers.
var workers [nWorkers]*AppWorker

// WorkerQueue is the channel of workers handling work.
var WorkerQueue chan AppWorker

// QuitChan is the channel we use to kill the workers.
var QuitChan chan int

// Worker is an interface for the basic app worker tasks.
type Worker interface {
	HandleTransaction(work TransactionWork)
	HandleBlock(work BlockWork)
}

// AppWorker implements the basic worker.
type AppWorker struct {
	ID int
	log.FieldLogger
}

// NewWorker returns a new AppWorker object.
func NewWorker(id int) AppWorker {
	return AppWorker{
		ID:          id,
		FieldLogger: log.WithField("id", id),
	}
}

// Start continually collects new work from existing work channels.
func (w AppWorker) Start() {
	go func() {
		for {
			// Wait for work.
			log.WithFields(log.Fields{
				"id": w.ID,
			}).Debug("Worker waiting for work")
			select {
			case work := <-TransactionWorkQueue:
				w.FieldLogger.Debug("Worker handling new transaction work")
				w.HandleTransaction(work)
			case work := <-BlockWorkQueue:
				w.FieldLogger.Debug("Worker handling new block work")
				w.HandleBlock(work)
			case <-QuitChan:
				w.FieldLogger.Debug("Worker quitting")
				return
			}
		}
	}()
}

// HandleTransaction handles new instance of TransactionWork.
func (w *AppWorker) HandleTransaction(work TransactionWork) {
	ok := tpool.Set(work.Transaction, chain)

	// Respond to the request if a response method was provided.
	if work.Responder != nil {
		work.Responder.Lock()
		defer work.Responder.Unlock()
		work.Responder.Send(ok)
	}
}

// HandleBlock handles TransactionWork.
func (w *AppWorker) HandleBlock(work BlockWork) {
	ok, _ := chain.ValidBlock(work.Block)
	if ok {
		chain.AppendBlock(work.Block, work.Miner)
	}

	// Respond to the request if a response method was provided.
	if work.Responder != nil {
		work.Responder.Lock()
		defer work.Responder.Unlock()
		work.Responder.Send(ok)
	}
}
