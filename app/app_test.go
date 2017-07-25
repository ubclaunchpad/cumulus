package app

import (
	"testing"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/ubclaunchpad/cumulus/blockchain"
	"github.com/ubclaunchpad/cumulus/msg"
	"github.com/ubclaunchpad/cumulus/peer"
	"github.com/ubclaunchpad/cumulus/pool"
)

func init() {
	log.SetLevel(log.InfoLevel)
}

func TestPushHandlerNewBlock(t *testing.T) {
	// Should put a block in the blockWorkQueue.
	a := createNewTestApp()
	_, b := blockchain.NewValidTestChainAndBlock()
	push := msg.Push{
		ResourceType: msg.ResourceBlock,
		Resource:     b,
	}
	a.PushHandler(&push)
	select {
	case work, ok := <-blockWorkQueue:
		assert.True(t, ok)
		assert.Equal(t, work.Block, b)
	}
}

func TestPushHandlerNewTestTransaction(t *testing.T) {
	// Should put a transaction in the transactionWorkQueue.
	a := createNewTestApp()
	txn := blockchain.NewTestTransaction()
	push := msg.Push{
		ResourceType: msg.ResourceTransaction,
		Resource:     txn,
	}
	a.PushHandler(&push)
	select {
	case work, ok := <-transactionWorkQueue:
		assert.True(t, ok)
		assert.Equal(t, work.Transaction, txn)
	}
}

func TestRequestHandlerNewBlockOK(t *testing.T) {
	// Request a new block by hash and verify we get the right one.
	a := createNewTestApp()

	req := createNewTestBlockRequest(a.Chain.Blocks[1].LastBlock)
	resp := a.RequestHandler(req)
	block, ok := resp.Resource.(*blockchain.Block)

	// Assertion time!
	assert.True(t, ok, "resource should contain block")
	assert.Equal(t, block, a.Chain.Blocks[1])
}

func TestRequestHandlerNewBlockBadParams(t *testing.T) {
	a := createNewTestApp()

	// Set up a request.
	hash := "definitelynotahash"
	req := createNewTestBlockRequest(hash)

	resp := a.RequestHandler(req)
	block, ok := resp.Resource.(*blockchain.Block)

	// Make sure request failed.
	assert.False(t, ok, "resource should not contain block")
	assert.Equal(t, resp.Error.Code, msg.ResourceNotFound, resp.Error.Message)
}

func TestRequestHandlerNewBlockBadType(t *testing.T) {
	a := createNewTestApp()

	// Set up a request.
	req := createNewTestBlockRequest("doesntmatter")
	req.ResourceType = 25

	resp := a.RequestHandler(req)
	block, ok := resp.Resource.(*blockchain.Block)

	// Make sure request failed.
	assert.False(t, ok, "resource should not contain block")
	assert.Equal(t, resp.Error.Code, msg.InvalidResourceType, resp.Error.Message)
}

func TestRequestHandlerPeerInfo(t *testing.T) {
	a := createNewTestApp()

	// Set up a request.
	req := createNewTestBlockRequest("doesntmatter")
	req.ResourceType = msg.ResourcePeerInfo

	resp := a.RequestHandler(req)
	res := resp.Resource

	// Make sure request did not fail.
	assert.NotNil(t, res, "resource should contain peer info")
	// Assert peer address returned valid.
}

func TestHandleTransactionOK(t *testing.T) {
	a := createNewTestApp()
	HandleTransaction(goodTxnWork)
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

func TestHandleBlockOK(t *testing.T) {
	reset()
	realWorker.HandleBlock(goodBlkWork)
	if mockResponder.Result != true {
		t.FailNow()
	}
}

func TestHandleBlockNotOK(t *testing.T) {
	reset()
	realWorker.HandleBlock(badBlkWork)
	if mockResponder.Result != false {
		t.FailNow()
	}
}

func TestStartTxn(t *testing.T) {
	reset()
	realWorker.Start()
	TransactionWorkQueue <- goodTxnWork
	time.Sleep(50 * time.Millisecond)
	mockResponder.Lock()
	if !mockResponder.Result {
		t.FailNow()
	}
	mockResponder.Unlock()
}

func TestStartBlk(t *testing.T) {
	reset()
	realWorker.Start()
	BlockWorkQueue <- goodBlkWork
	time.Sleep(50 * time.Millisecond)
	mockResponder.Lock()
	if !mockResponder.Result {
		t.FailNow()
	}
	mockResponder.Unlock()
}

func TestQuitWorker(t *testing.T) {
	reset()
	for i := 0; i < nWorkers; i++ {
		NewWorker(i).Start()
	}

	// Would hang if quit call fails, and travis would fail.
	for i := 0; i < nWorkers; i++ {
		QuitChan <- i
	}
}
