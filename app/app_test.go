package app

import (
	"testing"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/ubclaunchpad/cumulus/blockchain"
	"github.com/ubclaunchpad/cumulus/msg"
)

func init() {
	log.SetLevel(log.InfoLevel)
}

func createNewBlockRequest(blockNumber interface{}) *msg.Request {
	params := make(map[string]interface{}, 1)
	params["blockNumber"] = blockNumber
	return &msg.Request{
		ResourceType: msg.ResourceBlock,
		Params:       params,
	}
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

func TestRequestHandlerNewBlockOK(t *testing.T) {
	initializeChain()

	// Set up a request (requesting block 0)
	blockNumber := uint32(0)
	req := createNewBlockRequest(blockNumber)

	resp := RequestHandler(req)
	block, ok := resp.Resource.(*blockchain.Block)

	// Assertion time!
	assert.True(t, ok, "resource should contain block")
	assert.Equal(t, block.BlockNumber, blockNumber,
		"block number should be "+string(blockNumber))
}

func TestRequestHandlerNewBlockBadParams(t *testing.T) {
	initializeChain()

	// Set up a request.
	blockNumber := "definitelynotanindex"
	req := createNewBlockRequest(blockNumber)

	resp := RequestHandler(req)
	block, ok := resp.Resource.(*blockchain.Block)

	// Make sure request failed.
	assert.False(t, ok, "resource should not contain block")
	assert.Nil(t, block, "resource should not contain block")
}

func TestRequestHandlerNewBlockBadType(t *testing.T) {
	initializeChain()

	// Set up a request.
	req := createNewBlockRequest("doesntmatter")
	req.ResourceType = 25

	resp := RequestHandler(req)
	block, ok := resp.Resource.(*blockchain.Block)

	// Make sure request failed.
	assert.False(t, ok, "resource should not contain block")
	assert.Nil(t, block, "resource should not contain block")
}

func TestRequestHandlerPeerInfo(t *testing.T) {
	initializeChain()

	// Set up a request.
	req := createNewBlockRequest("doesntmatter")
	req.ResourceType = msg.ResourcePeerInfo

	resp := RequestHandler(req)
	res := resp.Resource

	// Make sure request did not fail.
	assert.NotNil(t, res, "resource should contain peer info")
	// Assert peer address returned valid.
}
