package app

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"os"
	"os/signal"
	"sync"
	"syscall"

	log "github.com/Sirupsen/logrus"
	"github.com/google/uuid"

	"github.com/ubclaunchpad/cumulus/blockchain"
	"github.com/ubclaunchpad/cumulus/conf"
	"github.com/ubclaunchpad/cumulus/conn"
	"github.com/ubclaunchpad/cumulus/consensus"
	"github.com/ubclaunchpad/cumulus/miner"
	"github.com/ubclaunchpad/cumulus/msg"
	"github.com/ubclaunchpad/cumulus/peer"
	"github.com/ubclaunchpad/cumulus/pool"
)

var (
	logFile = os.Stdout
)

const (
	// MaxBlockSize is the maximum size of a block in bytes
	MaxBlockSize = 5000000
	// MinBlockSize is the minimum size of a block in bytes
	MinBlockSize         = 1000
	blockQueueSize       = 100
	transactionQueueSize = 100
	userFileName         = "user.json"
	blockchainFileName   = "blockchain.json"
)

// App contains information about a running instance of a Cumulus node
type App struct {
	CurrentUser      *User
	PeerStore        *peer.PeerStore
	Chain            *blockchain.BlockChain
	Miner            *miner.Miner
	Pool             *pool.Pool
	blockQueue       chan *blockchain.Block
	transactionQueue chan *blockchain.Transaction
	quitChan         chan bool
}

// New returns a new user with the given parameters
func New(user *User, pStore *peer.PeerStore, chain *blockchain.BlockChain, pool *pool.Pool) *App {
	return &App{
		PeerStore:        pStore,
		CurrentUser:      user,
		Chain:            chain,
		Miner:            miner.New(),
		Pool:             pool,
		blockQueue:       make(chan *blockchain.Block, blockQueueSize),
		transactionQueue: make(chan *blockchain.Transaction, transactionQueueSize),
		quitChan:         make(chan bool),
	}
}

// Run sets up and starts a new Cumulus node with the
// given configuration. This should only be called once (except in tests)
func Run(cfg conf.Config) {
	// Set logging level
	if cfg.Verbose {
		log.SetLevel(log.DebugLevel)
	}

	log.Info("Starting Cumulus node")
	config := &cfg
	addr := fmt.Sprintf("%s:%d", config.Interface, config.Port)

	// Set starting difficulty (TODO: remove this when we have adjustable difficulty)
	consensus.CurrentDifficulty = big.NewInt(2 << 21)

	// Load user info from a file (or create a new user if there isn't one on disk)
	user, err := Load(userFileName)
	if err != nil {
		user = NewUser()
		if err := user.Save(userFileName); err != nil {
			log.WithError(err).Fatal("Failed to save new user info to ", userFileName)
		} else {
			log.Info("Saved new user info to file ", userFileName)
		}
	} else {
		log.Info("Loaded user info from ", userFileName)
	}

	// Load blockchain from a file (or create a new one if there isn't one on disk)
	chain, err := blockchain.Load(blockchainFileName)
	if err != nil {
		genesisBlock := blockchain.Genesis(user.Public(), consensus.CurrentTarget(),
			blockchain.StartingBlockReward, []byte{})
		chain = blockchain.New()
		chain.AppendBlock(genesisBlock)
		if err := user.Wallet.Refresh(chain); err != nil {
			log.WithError(err).Fatal("Failed to set user wallet information " +
				"based on the newly created blockchain")
		}
		log.Info("Created new blockchain with genesis block")
	} else {
		log.Info("Loaded blockchain from ", blockchainFileName)
	}

	// Create new app instance
	a := New(user, peer.NewPeerStore(addr), chain, pool.New())

	// We'll need to wait on at least 2 goroutines (Listen and
	// MaintainConnections) to start before returning
	wg := &sync.WaitGroup{}
	wg.Add(3)

	// Start a goroutine that waits for program termination. Before the program
	// exits it will flush logs and save the blockchain.
	go a.awaitExit(wg)

	// Below we'll connect to peers. After which, requests could begin to
	// stream in. Kick off a worker to handle requests and pushes.
	go a.HandleWork()

	// Set Peer default Push and Request handlers. These functions will handle
	// request and push messages from all peers we connect to unless overridden
	// for specific peers by calls like p.SetRequestHandler(someHandler)
	a.PeerStore.SetDefaultPushHandler(a.PushHandler)
	a.PeerStore.SetDefaultRequestHandler(a.RequestHandler)

	// Start listening on the given interface and port so we can receive
	// conenctions from other peers
	log.Infof("Starting listener on %s", addr)
	a.PeerStore.ListenAddr = addr
	go func() {
		err := conn.Listen(addr, a.PeerStore.ConnectionHandler, wg)
		if err != nil {
			log.WithError(err).Fatalf("Failed to listen on %s", addr)
		}
	}()

	// Try maintain as close to peer.MaxPeers connections as possible while this
	// peer is running
	go a.PeerStore.MaintainConnections(wg)

	// Wait for goroutines to start
	wg.Wait()

	if len(config.Target) > 0 {
		// Connect to the target, discover its peers, and download the blockchain
		a.ConnectAndDiscover(cfg.Target)
	}

	if config.Mine {
		log.Info("Starting miner")
		go a.RunMiner()
	}

	// If the console flag was passed, redirect logs to a file and run the console
	if cfg.Console {
		logFile, err := os.OpenFile("logfile", os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
		if err != nil {
			log.WithError(err).Fatal("Failed to redirect logs to file")
		}
		log.Info("Redirecting logs to logfile")
		log.SetOutput(logFile)
		go RunConsole(a)
	}
}

// ConnectAndDiscover tries to connect to a target and discover its peers.
func (a *App) ConnectAndDiscover(target string) {
	peerInfoRequest := msg.Request{
		ID:           uuid.New().String(),
		ResourceType: msg.ResourcePeerInfo,
	}

	log.Infof("Dialing target %s", target)
	p, err := peer.Connect(target, a.PeerStore)
	if err != nil {
		log.WithError(err).Fatal("Failed to dial target")
	}
	p.Request(peerInfoRequest, a.PeerStore.PeerInfoHandler)

	// Download blockchain
	log.Info("Syncronizing blockchain")
	_, err = a.SyncBlockChain()
	if err != nil {
		log.WithError(err).Fatal("Failed to download blockchain")
	}
	log.Info("Blockchain synchronization complete")
}

// RequestHandler is called every time a peer sends us a request message expect
// on peers whos PushHandlers have been overridden.
func (a *App) RequestHandler(req *msg.Request) msg.Response {
	res := msg.Response{ID: req.ID}

	// Build some error types.
	typeErr := msg.NewProtocolError(msg.InvalidResourceType,
		"Invalid resource type")
	notFoundErr := msg.NewProtocolError(msg.ResourceNotFound,
		"Resource not found")
	badRequestErr := msg.NewProtocolError(msg.BadRequest,
		"Bad request")
	upToDateErr := msg.NewProtocolError(msg.UpToDate,
		"The requested block has not yet been mined")

	switch req.ResourceType {
	case msg.ResourcePeerInfo:
		res.Resource = a.PeerStore.Addrs()
	case msg.ResourceBlock:
		log.Debug("Received block request")

		// Block is requested by block hash.
		hashBytes, err := json.Marshal(req.Params["lastBlockHash"])
		if err != nil {
			log.Debug("Returning response with status code: BadRequest")
			res.Error = badRequestErr
			break
		}

		a.Chain.RLock()
		defer a.Chain.RUnlock()

		var hash blockchain.Hash
		err = json.Unmarshal(hashBytes, &hash)
		if err != nil {
			log.Debug("Returning response with status code: BadRequest")
			res.Error = badRequestErr
			break
		} else if len(a.Chain.Blocks) > 0 && hash == blockchain.HashSum(a.Chain.LastBlock()) {
			log.Debug("Returning response with status code: UpToDate")
			res.Error = upToDateErr
			break
		}

		block, err := a.Chain.GetBlockByLastBlockHash(hash)
		if err != nil {
			log.Debug("Returning response with status code: ResourceNotFound")
			res.Error = notFoundErr
		} else {
			log.Debug("Returning response with block")
			res.Resource = block
		}
	default:
		res.Error = typeErr
	}

	return res
}

// PushHandler is called every time a peer sends us a Push message except on
// peers whos PushHandlers have been overridden.
func (a *App) PushHandler(push *msg.Push) {
	switch push.ResourceType {
	case msg.ResourceBlock:
		blockBytes, err := json.Marshal(push.Resource)
		if err != nil {
			return
		}
		block, err := blockchain.DecodeBlockJSON(blockBytes)
		if err != nil {
			log.WithError(err).Debug("Received invalid block")
			return
		}

		log.Debug("Adding block to work queue")
		a.blockQueue <- block

	case msg.ResourceTransaction:
		txnBytes, err := json.Marshal(push.Resource)
		if err != nil {
			log.WithError(err).Debug("Received invalid transaction")
			return
		}
		var txn blockchain.Transaction
		dec := json.NewDecoder(bytes.NewReader(txnBytes))
		dec.UseNumber()
		if err := dec.Decode(&txn); err != nil {
			log.WithError(err).Debug("Received invalid transaction")
			return
		}
		log.Debug("Adding transaction to work queue")
		a.transactionQueue <- &txn
	default:
		// Invalid resource type. Ignore
	}
}

// createBlockchain returns a new instance of a blockchain with only a genesis
// block.
func createBlockchain(user *User) *blockchain.BlockChain {
	bc := blockchain.New()
	genesisBlock := blockchain.Genesis(user.Wallet.Public(),
		consensus.CurrentTarget(), blockchain.StartingBlockReward, []byte{})

	bc.AppendBlock(genesisBlock)
	return bc
}

// HandleWork continually collects new work from existing work channels.
func (a *App) HandleWork() {
	log.Debug("Worker waiting for work")
	for {
		select {
		case work := <-a.transactionQueue:
			a.HandleTransaction(work)
		case work := <-a.blockQueue:
			a.HandleBlock(work)
		case <-a.quitChan:
			return
		}
	}
}

// HandleTransaction handles new transactions.
func (a *App) HandleTransaction(txn *blockchain.Transaction) {
	a.Chain.RLock()
	defer a.Chain.RUnlock()

	// If the transaction is not already in our pool, propagate it to the
	// network. This will ensure that transactions don't bounce back and forth
	// endlessly between nodes.
	if a.Pool.Get(blockchain.HashSum(txn)) != nil {
		a.PeerStore.Broadcast(msg.Push{
			ResourceType: msg.ResourceTransaction,
			Resource:     txn,
		})
		return
	}

	// We don't have this transaction in our pool, so we can try add it.
	code := a.Pool.Push(txn, a.Chain)
	if code == consensus.ValidTransaction {
		log.Debug("Added transaction to pool from address: " + txn.Sender.Repr())
	} else {
		log.Debug("Bad transaction rejected from sender: " + txn.Sender.Repr())
	}
}

// HandleBlock handles new blocks.
func (a *App) HandleBlock(blk *blockchain.Block) {
	wasMining := a.Miner.PauseIfRunning()

	a.Chain.Lock()
	defer a.Chain.Unlock()

	if blk.BlockNumber < uint32(len(a.Chain.Blocks)) {
		// We already have this block
		return
	}

	validBlock := a.Pool.Update(blk, a.Chain)
	if !validBlock {
		// The block was invalid wrt our chain. Maybe our chain is out of date.
		// Update it and try again.
		chainChanged, err := a.SyncBlockChain()
		if chainChanged {
			// We must update our wallet to reflect the new state of the blockchain
			if err := a.CurrentUser.Wallet.Refresh(a.Chain); err != nil {
				log.WithError(err).Fatal("Failed to update wallet")
			}
		}
		if err != nil {
			log.WithError(err).Error("Error synchronizing blockchain")
			if wasMining {
				a.ResumeMiner(chainChanged)
			}
			return
		}

		validBlock = a.Pool.Update(blk, a.Chain)
		if !validBlock {
			// Synchronizing our chain didn't help, the block is still invalid.
			if wasMining {
				a.ResumeMiner(chainChanged)
			}
			return
		}
	}

	// Append to the chain before requesting the next block so that the block
	// numbers make sense. Then update the user's wallet in case transactions
	// from the block affect it.
	a.Chain.AppendBlock(blk)
	if err := a.CurrentUser.Wallet.Update(blk, a.Chain); err != nil {
		log.WithError(err).Fatal("Attempt to add block with invalid " +
			"transaction(s) to the blockchain")
	}
	if wasMining {
		a.ResumeMiner(true)
	}
	log.Infof("Added block number %d to chain", blk.BlockNumber)
	log.Debug("Chain length: ", len(a.Chain.Blocks))
	return
}

// RunMiner continuously pulls transactions form the transaction pool, uses them to
// create blocks, and mines those blocks. When a block is mined it is added
// to the blockchain and broadcasted into the network. RunMiner returns when
// miner.StopMining() or miner.PauseIfRunning() are called.
func (a *App) RunMiner() {
	log.Debug("Miner started")
	for {
		a.Chain.RLock()

		// Make a new block form the transactions in the transaction pool
		blockToMine := a.Pool.NextBlock(a.Chain, a.CurrentUser.Wallet.Public(),
			a.CurrentUser.BlockSize)

		a.Chain.RUnlock()

		// TODO: update this when we have adjustable difficulty
		blockToMine.Target = consensus.CurrentTarget()
		miningResult := a.Miner.Mine(blockToMine)

		if miningResult.Complete {
			log.Info("Successfully mined block ", blockToMine.BlockNumber)
			a.HandleBlock(blockToMine)
			push := msg.Push{
				ResourceType: msg.ResourceBlock,
				Resource:     blockToMine,
			}
			a.PeerStore.Broadcast(push)
		} else if miningResult.Info == miner.MiningHalted {
			log.Debug("Miner stopped")
			return
		}
	}
}

// ResumeMiner resumes the current mining job if restart is false, otherwise it
// restarts the miner with a new mining job.
func (a *App) ResumeMiner(restart bool) {
	if !restart {
		a.Miner.ResumeMining()
	} else {
		a.Miner.StopMining()
		go a.RunMiner()
	}
}

// SyncBlockChain updates the local copy of the blockchain by requesting missing
// blocks from peers. Returns true if the blockchain changed as a result of
// calling this function, false if it didn't and an error if we are not connected
// to any peers.
func (a *App) SyncBlockChain() (bool, error) {
	newBlockChan := make(chan *blockchain.Block)
	errChan := make(chan *msg.ProtocolError)
	chainChanged := false

	// Define a handler for responses to our block requests
	blockResponseHandler := func(blockResponse *msg.Response) {
		if blockResponse.Error != nil {
			errChan <- blockResponse.Error
			return
		}

		blockBytes, err := json.Marshal(blockResponse.Resource)
		if err != nil {
			newBlockChan <- nil
			return
		}

		block, err := blockchain.DecodeBlockJSON(blockBytes)
		if err != nil {
			newBlockChan <- nil
			return
		}

		newBlockChan <- block
	}

	// Continually request the block after the latest block in our chain until
	// we are totally up to date
	for {
		err := a.makeBlockRequest(a.Chain.LastBlock(), blockResponseHandler)
		if err != nil {
			if a.PeerStore.Size() == 0 {
				// No peers to make the request to
				return chainChanged, err
			}
			continue
		}

		// Wait for response
		change, done := a.handleBlockResponse(newBlockChan, errChan)
		if change {
			chainChanged = true
		}
		if done {
			return chainChanged, nil
		}
	}
}

// makeBlockRequest creates and sends a request to a random peer for the block
// following the given block and returns an error if the request could not be
// created or sent.
func (a *App) makeBlockRequest(currentHead *blockchain.Block,
	responseHandler peer.ResponseHandler) error {
	// In case the blockchain is empty
	currentHeadHash := blockchain.NilHash
	if currentHead != nil {
		currentHeadHash = blockchain.HashSum(currentHead)
	}

	reqParams := map[string]interface{}{
		"lastBlockHash": currentHeadHash,
	}

	blockRequest := msg.Request{
		ID:           uuid.New().String(),
		ResourceType: msg.ResourceBlock,
		Params:       reqParams,
	}

	// Pick a peer to send the request to
	p := a.PeerStore.GetRandom()
	if p == nil {
		return errors.New("SyncBlockchain failed: no peers to request blocks from")
	}
	return p.Request(blockRequest, responseHandler)
}

// handleBlockResponse receives a block or nil from the newBlockChan and attempts
// to validate it and add it to the blockchain, or it handles a protocol error
// from the errChan. Returns whether the blockchain was modified and whether we
// received an UpToDate response.
func (a *App) handleBlockResponse(newBlockChan chan *blockchain.Block,
	errChan chan *msg.ProtocolError) (changed bool, upToDate bool) {
	select {
	case newBlock := <-newBlockChan:
		if newBlock == nil {
			// We received a response with no error but an invalid resource
			// Try again
			log.Debug("Received block response with invalid resource")
			return false, false
		}

		valid, validationCode := consensus.VerifyBlock(a.Chain, newBlock)
		if !valid {
			// There is something wrong with this block. Try again
			fields := log.Fields{"validationCode": validationCode}
			log.WithFields(fields).Debug("SyncBlockchain received invalid block")
			return false, false
		}

		// Valid block. Append it to the chain
		log.Debugf("Adding block %d to blockchain", newBlock.BlockNumber)
		log.Debug("Blockchain length: ", len(a.Chain.Blocks))
		a.Chain.AppendBlock(newBlock)
		return true, false

	case err := <-errChan:
		if err.Code == msg.ResourceNotFound {
			// Our chain might be out of sync, roll it back by one block
			// and request the next block
			log.Debug("Received response with status code: ResourceNotFound")
			a.Chain.RollBack()
			return true, false
		} else if err.Code == msg.UpToDate {
			log.Debug("Received response with status code: UpToDate")
			return false, true
		}
		log.Debug("Received response with unexpected status code: ", err.Code)
		return false, false
	}
}

// awaitExit waits until the interrupt or terminate signal and cleans up before
// cumulus terminates.
func (a *App) awaitExit(wg *sync.WaitGroup) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	wg.Done()
	<-c
	a.onExit()
}

// onExit saves app state to disk before exiting.
func (a *App) onExit() {
	log.Info("Saving app state and flushing logs...")
	if err := a.Chain.Save(blockchainFileName); err != nil {
		log.WithError(err).Error("Error saving blockchain")
	}
	if err := a.CurrentUser.Save(userFileName); err != nil {
		log.WithError(err).Error("Error saving user info")
	}
	logFile.Sync()
	logFile.Close()
	os.Exit(0)
}
