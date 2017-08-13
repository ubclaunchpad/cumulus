package app

import (
	"testing"
	"time"

	"github.com/ubclaunchpad/cumulus/conn"
	"github.com/ubclaunchpad/cumulus/peer"

	"github.com/ubclaunchpad/cumulus/miner"

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
	case blk, ok := <-a.blockQueue:
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
	case tr, ok := <-a.transactionQueue:
		assert.True(t, ok)
		assert.Equal(t, tr, txn)
	}
}

func TestRequestHandlerNewBlockOK(t *testing.T) {
	// Request a new block by hash and verify we get the right one.
	a := createNewTestApp()

	req := createNewTestBlockRequest(a.Chain.Blocks[1].LastBlock)
	resp := a.RequestHandler(req)
	block, ok := resp.Resource.(*blockchain.Block)

	assert.True(t, ok, "resource should contain block")
	assert.Equal(t, block, a.Chain.Blocks[1])
}

func TestRequestHandlerNewBlockBadParams(t *testing.T) {
	a := createNewTestApp()

	// Set up a request.
	hash := "definitelynotahash"
	req := createNewTestBlockRequest(hash)

	resp := a.RequestHandler(req)
	_, ok := resp.Resource.(*blockchain.Block)

	// Make sure request failed.
	assert.False(t, ok, "resource should not contain block")
	assert.Equal(t, msg.BadRequest, int(resp.Error.Code), resp.Error.Message)
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

func TestHandleValidBlock(t *testing.T) {
	a := createNewTestApp()
	i := 0

	// TODO: Start miner.

	for i < 1000 {
		a.Pool.PushUnsafe(blockchain.NewTestTransaction())
		i++
	}

	bc, blk := blockchain.NewValidTestChainAndBlock()
	a.Chain = bc
	a.HandleBlock(blk)
	assert.Equal(t, blk, a.Chain.Blocks[2])

	// TODO: Assert miner restarted.
	// TODO: Assert pool appropriately emptied.
}

func TestHandleInvalidBlock(t *testing.T) {
	a := createNewTestApp()
	i := 0

	// TODO: Start miner.
	for i < 1000 {
		a.Pool.PushUnsafe(blockchain.NewTestTransaction())
		i++
	}

	a.HandleBlock(blockchain.NewTestBlock())
	// TODO: Assert miner not restarted.
	// TODO: Assert pool untouched.
}

func TestGetLocalPool(t *testing.T) {
	assert.NotNil(t, getLocalPool())
}

func TestCreateBlockchain(t *testing.T) {
	assert.NotNil(t, createBlockchain(NewUser()))
}

func TestHandleBlock(t *testing.T) {
	a := createNewTestApp()
	go a.HandleWork()
	time.Sleep(50 * time.Millisecond)
	a.blockQueue <- blockchain.NewTestBlock()
	assert.Equal(t, len(a.blockQueue), 0)
}

func TestHandleTransaction(t *testing.T) {
	a := createNewTestApp()
	go a.HandleWork()
	time.Sleep(50 * time.Millisecond)
	a.transactionQueue <- blockchain.NewTestTransaction()
	assert.Equal(t, len(a.transactionQueue), 0)
}

func TestRunMiner(t *testing.T) {
	a := createNewTestApp()
	assert.False(t, miner.IsMining())
	go a.RunMiner()
	time.Sleep(time.Second)
	assert.True(t, miner.IsMining())
	a.RestartMiner()
	time.Sleep(time.Second)
	assert.True(t, miner.IsMining())
	miner.StopMining()
	assert.False(t, miner.IsMining())
	a.RestartMiner()
	assert.False(t, miner.IsMining())
}

func TestMakeBlockRequest(t *testing.T) {
	bc := conn.NewBufConn(false, true)
	a := createNewTestApp()
	p := peer.New(bc, a.PeerStore, "")
	a.PeerStore.Add(p)
	done := make(chan bool, 1)

	rh := func(res *msg.Response) {}
	bc.OnWrite = func(b []byte) {
		done <- true
	}

	err := a.makeBlockRequest(a.Chain.LastBlock(), rh)
	assert.Nil(t, err)
	assert.True(t, <-done)
}

func TestMakeBlockRequestNoPeers(t *testing.T) {
	a := createNewTestApp()
	rh := func(res *msg.Response) {}
	err := a.makeBlockRequest(a.Chain.LastBlock(), rh)
	assert.NotNil(t, err)
}

func TestHandleBlockResponse(t *testing.T) {
	a := createNewTestApp()
	newBlockChan := make(chan *blockchain.Block, 1)
	errChan := make(chan *msg.ProtocolError, 1)
	newBlockChan <- a.Chain.RollBack()
	changed, upToDate := a.handleBlockResponse(newBlockChan, errChan)
	assert.True(t, changed)
	assert.False(t, upToDate)
	newBlockChan <- a.Chain.Blocks[1]
	changed, upToDate = a.handleBlockResponse(newBlockChan, errChan)
	assert.False(t, changed)
	assert.False(t, upToDate)
	errChan <- msg.NewProtocolError(msg.UpToDate, "")
	changed, upToDate = a.handleBlockResponse(newBlockChan, errChan)
	assert.False(t, changed)
	assert.True(t, upToDate)
	errChan <- msg.NewProtocolError(msg.ResourceNotFound, "")
	changed, upToDate = a.handleBlockResponse(newBlockChan, errChan)
	assert.True(t, changed)
	assert.False(t, upToDate)
	assert.Equal(t, len(a.Chain.Blocks), 1)
}

func TestHandleWork(t *testing.T) {
	a := createNewTestApp()
	go a.HandleWork()
	a.blockQueue <- a.Chain.RollBack()
	time.Sleep(time.Second)
	assert.Equal(t, len(a.Chain.Blocks), 2)
	a.quitChan <- true
	a.blockQueue <- a.Chain.RollBack()
	time.Sleep(time.Second)
	assert.Equal(t, len(a.Chain.Blocks), 1)
}

func TestPay(t *testing.T) {
	amt := uint64(5)
	a := createNewTestApp()
	err := a.Pay("badf00d", amt)

	// Fail with low balance.
	assert.NotNil(t, err)

	a.CurrentUser.Wallet.SetBalance(amt)
	err = a.Pay("badf00d", amt)

	// Fail with bad inputs.
	assert.NotNil(t, err)
}
