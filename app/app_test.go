package app

import (
	"testing"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/ubclaunchpad/cumulus/msg"
)

func init() {
	log.SetLevel(log.DebugLevel)
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
	push := msg.Push{ResourceType: msg.ResourceBlock}
	PushHandler(&push)
	select {
	case _, ok := <-BlockWorkQueue:
		if !ok {
			t.FailNow()
		}
	}
	// Add more here...
}

func TestPushHandlerNewTransaction(t *testing.T) {
	intializeQueues()
	push := msg.Push{ResourceType: msg.ResourceTransaction}
	PushHandler(&push)
	select {
	case _, ok := <-TransactionWorkQueue:
		if !ok {
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

func TestRequestHandlerNewTransaction(t *testing.T) {
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
