package app

import (
	"sync"
	"testing"

	log "github.com/Sirupsen/logrus"
	bc "github.com/ubclaunchpad/cumulus/blockchain"
	"github.com/ubclaunchpad/cumulus/pool"
)

var (
	legitBlock       *bc.Block
	legitTransaction *bc.Transaction
	realWorker       GenericWorker
	txnWork          TransactionWork
	mockResponder    MockResponder
	badTxnWork       TransactionWork
	goodTxnWork      TransactionWork
	badBlkWork       BlockWork
	goodBlkWork      BlockWork
)

type MockResponder struct {
	Result   bool
	NumCalls int
	*sync.Mutex
}

func (r *MockResponder) Send(ok bool) {
	r.Result = ok
	r.NumCalls++
}

func (r *MockResponder) Lock() {
	r.Mutex.Lock()
}

func (r *MockResponder) Unlock() {
	r.Mutex.Unlock()
}

func init() {
	log.SetLevel(log.DebugLevel)
}

func reset() {
	tpool = pool.New()
	chain, legitBlock = bc.NewValidChainAndBlock()
	legitTransaction = legitBlock.Transactions[0]
	realWorker = NewWorker(7)
	mockResponder = MockResponder{
		Mutex:  &sync.Mutex{},
		Result: false,
	}
	goodTxnWork = TransactionWork{
		Transaction: legitTransaction,
		Responder:   &mockResponder,
	}
	badTxnWork = TransactionWork{
		Transaction: bc.NewTransaction(),
		Responder:   &mockResponder,
	}
	goodBlkWork = BlockWork{
		Block:     legitBlock,
		Responder: &mockResponder,
	}
	badBlkWork = BlockWork{
		Block:     bc.NewBlock(),
		Responder: &mockResponder,
	}
	QuitChan = make(chan bool)
}

func TestNewWorker(t *testing.T) {
	reset()
	if realWorker.ID != 7 {
		t.FailNow()
	}
}

func TestHandleTransactionOK(t *testing.T) {
	reset()
	realWorker.HandleTransaction(goodTxnWork)
	if mockResponder.Result != true {
		t.FailNow()
	}
}

func TestHandleTransactionNotOK(t *testing.T) {
	reset()
	realWorker.HandleTransaction(badTxnWork)
	if mockResponder.Result != false {
		t.FailNow()
	}
}

func TestHandleBlockNotOK(t *testing.T) {
	reset()
	realWorker.HandleBlock(goodBlkWork)
	if mockResponder.Result != true {
		t.FailNow()
	}
}

func TestHandleBlockOK(t *testing.T) {
	reset()
	realWorker.HandleBlock(badBlkWork)
	if mockResponder.Result != false {
		t.FailNow()
	}
}
