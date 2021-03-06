package app

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/ubclaunchpad/cumulus/blockchain"
	"github.com/ubclaunchpad/cumulus/common/constants"
	"github.com/ubclaunchpad/cumulus/conf"
	"github.com/ubclaunchpad/cumulus/conn"
	"github.com/ubclaunchpad/cumulus/consensus"
	"github.com/ubclaunchpad/cumulus/miner"
	"github.com/ubclaunchpad/cumulus/msg"
	"github.com/ubclaunchpad/cumulus/peer"
	"github.com/ubclaunchpad/cumulus/pool"
)

func TestNew(t *testing.T) {
	ps := peer.NewPeerStore("123")
	user := NewUser()
	chain := blockchain.NewTestBlockChain()
	pool := pool.New()
	a := New(user, ps, chain, pool)
	assert.Equal(t, ps, a.PeerStore)
	assert.Equal(t, user, a.CurrentUser)
	assert.Equal(t, chain, a.Chain)
	assert.Equal(t, pool, a.Pool)
}

func TestPushHandlerNewBlock(t *testing.T) {
	// Should put a block in the blockWorkQueue.
	a := newTestApp()
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
	a := newTestApp()
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
	a := newTestApp()

	req := newTestBlockRequest(a.Chain.Blocks[1].LastBlock)
	resp := a.RequestHandler(req)
	block, ok := resp.Resource.(*blockchain.Block)

	assert.True(t, ok, "resource should contain block")
	assert.Equal(t, block, a.Chain.Blocks[1])
}

func TestRequestHandlerNewBlockBadParams(t *testing.T) {
	a := newTestApp()

	// Set up a request.
	hash := "definitelynotahash"
	req := newTestBlockRequest(hash)

	resp := a.RequestHandler(req)
	_, ok := resp.Resource.(*blockchain.Block)

	// Make sure request failed.
	assert.False(t, ok, "resource should not contain block")
	assert.Equal(t, msg.BadRequest, int(resp.Error.Code), resp.Error.Message)
}

func TestRequestHandlerNewBlockBadType(t *testing.T) {
	a := newTestApp()

	// Set up a request.
	req := newTestBlockRequest("doesntmatter")
	req.ResourceType = 25

	resp := a.RequestHandler(req)
	_, ok := resp.Resource.(*blockchain.Block)

	// Make sure request failed.
	assert.False(t, ok, "resource should not contain block")
	assert.Equal(t, msg.InvalidResourceType, int(resp.Error.Code), resp.Error.Message)
}

func TestRequestHandlerPeerInfo(t *testing.T) {
	a := newTestApp()

	// Set up a request.
	req := newTestBlockRequest("doesntmatter")
	req.ResourceType = msg.ResourcePeerInfo

	resp := a.RequestHandler(req)
	res := resp.Resource

	// Make sure request did not fail.
	assert.NotNil(t, res, "resource should contain peer info")
	// Assert peer address returned valid.
}

func TestHandleTransactionOK(t *testing.T) {
	a := newTestApp()
	bc, blk := blockchain.NewValidTestChainAndBlock()
	a.Chain = bc
	txn := blk.Transactions[1]
	a.HandleTransaction(txn)
	assert.False(t, a.Pool.Empty())
	assert.Equal(t, a.Pool.Peek(), txn)
}

func TestHandleTransactionNotOK(t *testing.T) {
	a := newTestApp()
	a.HandleTransaction(blockchain.NewTestTransaction())
	assert.True(t, a.Pool.Empty())
}

func TestHandleValidBlock(t *testing.T) {
	a := newTestApp()
	i := 0

	// TODO: Start miner.

	for i < 1000 {
		a.Pool.PushUnsafe(blockchain.NewTestTransaction())
		i++
	}

	bc, blk := blockchain.NewValidTestChainAndBlock()
	a.Chain = bc
	a.HandleBlock(blk)
	assert.Equal(t, blk, a.Chain.Blocks[3])

	// TODO: Assert miner restarted.
	// TODO: Assert pool appropriately emptied.
}

func TestHandleInvalidBlock(t *testing.T) {
	a := newTestApp()
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

func TestCreateBlockchain(t *testing.T) {
	assert.NotNil(t, createBlockchain(NewUser()))
}

func TestHandleBlock(t *testing.T) {
	a := newTestApp()
	go a.HandleWork()
	time.Sleep(50 * time.Millisecond)
	a.blockQueue <- blockchain.NewTestBlock()
	assert.Equal(t, len(a.blockQueue), 0)
}

func TestHandleTransaction(t *testing.T) {
	a := newTestApp()
	go a.HandleWork()
	time.Sleep(50 * time.Millisecond)
	a.transactionQueue <- blockchain.NewTestTransaction()
	assert.Equal(t, len(a.transactionQueue), 0)
}

func TestRunMiner(t *testing.T) {
	oldDifficulty := consensus.CurrentDifficulty
	consensus.CurrentDifficulty = constants.MaxTarget
	a := newTestApp()
	go a.RunMiner()
	time.Sleep(time.Second / 2)
	assert.Equal(t, int(a.Miner.State()), int(miner.Running))
	assert.True(t, a.Miner.PauseIfRunning())
	assert.Equal(t, int(a.Miner.State()), int(miner.Paused))
	a.ResumeMiner(false)
	time.Sleep(time.Second / 2)
	assert.Equal(t, int(a.Miner.State()), int(miner.Running))
	a.Miner.PauseIfRunning()
	time.Sleep(time.Second / 2)
	assert.Equal(t, int(a.Miner.State()), int(miner.Paused))
	a.ResumeMiner(true)
	time.Sleep(time.Second / 2)
	assert.Equal(t, int(a.Miner.State()), int(miner.Running))
	a.Miner.StopMining()
	assert.Equal(t, int(a.Miner.State()), int(miner.Stopped))
	consensus.CurrentDifficulty = oldDifficulty
}

func TestMakeBlockRequest(t *testing.T) {
	bc := conn.NewBufConn(false, true)
	a := newTestApp()
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
	a := newTestApp()
	rh := func(res *msg.Response) {}
	err := a.makeBlockRequest(a.Chain.LastBlock(), rh)
	assert.NotNil(t, err)
}

func TestHandleBlockResponse(t *testing.T) {
	a := newTestApp()
	newBlockChan := make(chan *blockchain.Block, 1)
	errChan := make(chan *msg.ProtocolError, 1)

	// Roll the last block off of the chain, add to channel.
	lastBlock := a.Chain.RollBack()
	assert.NotNil(t, lastBlock)
	newBlockChan <- lastBlock

	// Handle the block.
	changed, upToDate := a.handleBlockResponse(newBlockChan, errChan)
	assert.True(t, changed)
	assert.False(t, upToDate)

	// Add an on-chain block to the handler (this shouldn't change the chain).
	newBlockChan <- a.Chain.Blocks[1]
	changed, upToDate = a.handleBlockResponse(newBlockChan, errChan)
	assert.False(t, changed)
	assert.False(t, upToDate)

	// More stuff happens.
	errChan <- msg.NewProtocolError(msg.UpToDate, "")
	changed, upToDate = a.handleBlockResponse(newBlockChan, errChan)
	assert.False(t, changed)
	assert.True(t, upToDate)
	errChan <- msg.NewProtocolError(msg.ResourceNotFound, "")
	changed, upToDate = a.handleBlockResponse(newBlockChan, errChan)
	assert.True(t, changed)
	assert.False(t, upToDate)
	assert.Equal(t, len(a.Chain.Blocks), 2)
}

func TestHandleWork(t *testing.T) {
	a := newTestApp()
	go a.HandleWork()

	// Roll the last block off of the chain, add to channel (expect it added back).
	lastBlock := a.Chain.RollBack()
	assert.Equal(t, len(a.Chain.Blocks), 2)
	a.blockQueue <- lastBlock
	time.Sleep(time.Millisecond * 50)
	assert.Equal(t, len(a.Chain.Blocks), 3)

	// Kill the worker.
	a.quitChan <- true

	// Roll the last block off of the chain, add to channel
	// (expect it not to be added back).
	lastBlock = a.Chain.RollBack()
	assert.Equal(t, len(a.Chain.Blocks), 2)
	a.blockQueue <- lastBlock
	time.Sleep(time.Millisecond * 50)
	assert.Equal(t, len(a.Chain.Blocks), 2)
}

func TestPay(t *testing.T) {
	amt := uint64(5)
	a := newTestApp()
	err := a.Pay("badf00d", amt)

	// Fail with low balance.
	assert.NotNil(t, err)

	a.CurrentUser.Wallet.Balance = amt
	err = a.Pay("badf00d", amt)

	// Fail with bad inputs.
	assert.NotNil(t, err)
}

func TestRun(t *testing.T) {
	cfg := conf.Config{
		Interface: "127.0.0.1",
		Port:      8080,
		Target:    "",
		Verbose:   true,
		Mine:      true,
		Console:   false,
	}
	Run(cfg)
	assert.Nil(t, os.Remove(userFileName))
}
