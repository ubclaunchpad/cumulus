package app

import (
	"testing"

	log "github.com/Sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/ubclaunchpad/cumulus/blockchain"
	"github.com/ubclaunchpad/cumulus/msg"
)

func init() {
	log.SetLevel(log.DebugLevel)
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
	case blk, ok := <-blockQueue:
		assert.True(t, ok)
		assert.Equal(t, blk, b)
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
	case tr, ok := <-transactionQueue:
		assert.True(t, ok)
		assert.Equal(t, tr, txn)
	}
}

// TODO: Enable once block request by hash implemented.
// func TestRequestHandlerNewBlockOK(t *testing.T) {
// 	// Request a new block by hash and verify we get the right one.
// 	a := createNewTestApp()

// 	req := createNewTestBlockRequest(a.Chain.Blocks[1].LastBlock)
// 	resp := a.RequestHandler(req)
// 	block, ok := resp.Resource.(*blockchain.Block)

// 	assert.True(t, ok, "resource should contain block")
// 	assert.Equal(t, block, a.Chain.Blocks[1])
// }

func TestRequestHandlerNewBlockBadParams(t *testing.T) {
	a := createNewTestApp()

	// Set up a request.
	hash := "definitelynotahash"
	req := createNewTestBlockRequest(hash)

	resp := a.RequestHandler(req)
	_, ok := resp.Resource.(*blockchain.Block)

	// Make sure request failed.
	assert.False(t, ok, "resource should not contain block")
	assert.Equal(t, msg.ResourceNotFound, int(resp.Error.Code), resp.Error.Message)
}

func TestRequestHandlerNewBlockBadType(t *testing.T) {
	a := createNewTestApp()

	// Set up a request.
	req := createNewTestBlockRequest("doesntmatter")
	req.ResourceType = 25

	resp := a.RequestHandler(req)
	_, ok := resp.Resource.(*blockchain.Block)

	// Make sure request failed.
	assert.False(t, ok, "resource should not contain block")
	assert.Equal(t, msg.InvalidResourceType, int(resp.Error.Code), resp.Error.Message)
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
	bc, blk := blockchain.NewValidTestChainAndBlock()
	a.Chain = bc
	txn := blk.Transactions[1]
	a.HandleTransaction(txn)
	assert.False(t, a.Pool.Empty())
	assert.Equal(t, a.Pool.Peek(), txn)
}

func TestHandleTransactionNotOK(t *testing.T) {
	a := createNewTestApp()
	a.HandleTransaction(blockchain.NewTestTransaction())
	assert.True(t, a.Pool.Empty())
}

func TestHandleBlockOK(t *testing.T) {
	a := createNewTestApp()
	i := 0

	// TODO: Start miner.

	for i < 1000 {
		a.Pool.SetUnsafe(blockchain.NewTestTransaction())
		i++
	}

	bc, blk := blockchain.NewValidTestChainAndBlock()
	a.Chain = bc
	a.HandleBlock(blk)
	assert.Equal(t, blk, a.Chain.Blocks[2])

	// TODO: Assert miner restarted.
	// TODO: Assert pool appropriately emptied.
}

func TestHandleBlockNotOK(t *testing.T) {
	a := createNewTestApp()
	i := 0

	// TODO: Start miner.
	for i < 1000 {
		a.Pool.SetUnsafe(blockchain.NewTestTransaction())
		i++
	}

	a.HandleBlock(blockchain.NewTestBlock())
	// TODO: Assert miner not restarted.
	// TODO: Assert pool untouched.
}

func TestGetLocalPool(t *testing.T) {
	assert.NotNil(t, getLocalPool())
}

func TestGetLocalChain(t *testing.T) {
	assert.NotNil(t, getLocalChain())
}

func TestHandleBlock(t *testing.T) {
	a := createNewTestApp()
	go a.HandleWork()
	blockQueue <- blockchain.NewTestBlock()
	assert.Equal(t, len(blockQueue), 0)
}

func TestHandleTransaction(t *testing.T) {
	a := createNewTestApp()
	go a.HandleWork()
	transactionQueue <- blockchain.NewTestTransaction()
	assert.Equal(t, len(transactionQueue), 0)
}
