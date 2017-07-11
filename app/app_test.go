package app

import (
	"testing"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/ubclaunchpad/cumulus/blockchain"
	"github.com/ubclaunchpad/cumulus/msg"
)

func init() {
	log.SetLevel(log.InfoLevel)
}

func TestKillWorkers(t *testing.T) {
	intializeQueues()
	time.Sleep(20 * time.Millisecond)
	for i := 0; i < nWorkers; i++ {
		if workers[i] != nil {
			t.FailNow()
		}
	}
	initializeWorkers()
	time.Sleep(20 * time.Millisecond)
	for i := 0; i < nWorkers; i++ {
		if workers[i] == nil {
			t.FailNow()
		}
	}
	killWorkers()
	time.Sleep(20 * time.Millisecond)
	for i := 0; i < nWorkers; i++ {
		if workers[i] != nil {
			t.FailNow()
		}
	}
}

func TestInitializeNode(t *testing.T) {
	initializeNode()
	if tpool == nil {
		t.FailNow()
	}
	if BlockWorkQueue == nil {
		t.FailNow()
	}
	if TransactionWorkQueue == nil {
		t.FailNow()
	}
	killWorkers()
}

func TestPushHandlerNewBlock(t *testing.T) {
	intializeQueues()
	_, b := blockchain.NewValidTestChainAndBlock()
	push := msg.Push{
		ResourceType: msg.ResourceBlock,
		Resource:     b,
	}
	PushHandler(&push)
	select {
	case work, ok := <-BlockWorkQueue:
		if !ok {
			t.FailNow()
		}
		if work.Block != b {
			t.FailNow()
		}
	}
	// Add more here...
}

func TestPushHandlerNewTestTransaction(t *testing.T) {
	intializeQueues()
	txn := blockchain.NewTestTransaction()
	push := msg.Push{
		ResourceType: msg.ResourceTransaction,
		Resource:     txn,
	}
	PushHandler(&push)
	select {
	case work, ok := <-TransactionWorkQueue:
		if !ok {
			t.FailNow()
		}
		if work.Transaction != txn {
			t.FailNow()
		}
	}
	// Add more here...
}

func TestRequestHandlerNewBlock(t *testing.T) {
	intializeQueues()
	push := msg.Request{ResourceType: msg.ResourceBlock}
	RequestHandler(&push)
	select {
	case _, ok := <-BlockWorkQueue:
		if !ok {
			t.FailNow()
		}
	}
	// Add more here...
}

func TestRequestHandlerNewTestTransaction(t *testing.T) {
	intializeQueues()
	push := msg.Request{ResourceType: msg.ResourceTransaction}
	RequestHandler(&push)
	select {
	case _, ok := <-TransactionWorkQueue:
		if !ok {
			t.FailNow()
		}
	}
	// Add more here...
}
