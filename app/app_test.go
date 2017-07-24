package app

import (
	"testing"

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

func createNewBlockRequest(lastBlock interface{}) *msg.Request {
	params := make(map[string]interface{}, 1)
	params["lastBlock"] = lastBlock
	return &msg.Request{
		ResourceType: msg.ResourceBlock,
		Params:       params,
	}
}

func createNewApp() *App {
	chain, _ := blockchain.NewValidTestChainAndBlock()
	return &App{
		PeerStore:   peer.NewPeerStore("127.0.0.1:8000"),
		CurrentUser: NewUser(),
		Chain:       chain,
		Pool:        pool.New(),
	}
}

func TestPushHandlerNewBlock(t *testing.T) {
	// Should put a block in the blockWorkQueue.
	a := createNewApp()
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
	a := createNewApp()
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
	a := createNewApp()

	req := createNewBlockRequest(a.Chain.Blocks[1].LastBlock)
	resp := a.RequestHandler(req)
	block, ok := resp.Resource.(*blockchain.Block)

	// Assertion time!
	assert.True(t, ok, "resource should contain block")
	assert.Equal(t, block, a.Chain.Blocks[1])
}

func TestRequestHandlerNewBlockBadParams(t *testing.T) {
	a := createNewApp()

	// Set up a request.
	hash := "definitelynotahash"
	req := createNewBlockRequest(hash)

	resp := a.RequestHandler(req)
	block, ok := resp.Resource.(*blockchain.Block)

	// Make sure request failed.
	assert.False(t, ok, "resource should not contain block")
	assert.Equal(t, resp.Error.Code, msg.ResourceNotFound, resp.Error.Message)
}

func TestRequestHandlerNewBlockBadType(t *testing.T) {
	a := createNewApp()

	// Set up a request.
	req := createNewBlockRequest("doesntmatter")
	req.ResourceType = 25

	resp := a.RequestHandler(req)
	block, ok := resp.Resource.(*blockchain.Block)

	// Make sure request failed.
	assert.False(t, ok, "resource should not contain block")
	assert.Equal(t, resp.Error.Code, msg.InvalidResourceType, resp.Error.Message)
}

func TestRequestHandlerPeerInfo(t *testing.T) {
	a := createNewApp()

	// Set up a request.
	req := createNewBlockRequest("doesntmatter")
	req.ResourceType = msg.ResourcePeerInfo

	resp := a.RequestHandler(req)
	res := resp.Resource

	// Make sure request did not fail.
	assert.NotNil(t, res, "resource should contain peer info")
	// Assert peer address returned valid.
}
