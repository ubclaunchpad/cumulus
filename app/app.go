package app

import (
	"errors"
	"fmt"
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
	logFile          = os.Stdout
	blockQueue       = make(chan *blockchain.Block, blockQueueSize)
	transactionQueue = make(chan *blockchain.Transaction, transactionQueueSize)
	quitChan         = make(chan bool)
)

const (
	blockQueueSize       = 100
	transactionQueueSize = 100
)

// App contains information about a running instance of a Cumulus node
type App struct {
	CurrentUser *User
	PeerStore   *peer.PeerStore
	Chain       *blockchain.BlockChain
	Pool        *pool.Pool
}

// Run sets up and starts a new Cumulus node with the
// given configuration. This should only be called once (except in tests)
func Run(cfg conf.Config) {
	log.Info("Starting Cumulus node")
	config := &cfg

	addr := fmt.Sprintf("%s:%d", config.Interface, config.Port)
	user := getCurrentUser()
	a := App{
		PeerStore:   peer.NewPeerStore(addr),
		CurrentUser: user,
		Chain:       getLocalChain(user),
		Pool:        getLocalPool(),
	}

	// Set logging level
	if cfg.Verbose {
		log.SetLevel(log.DebugLevel)
	}

	// We'll need to wait on at least 2 goroutines (Listen and
	// MaintainConnections) to start before returning
	wg := &sync.WaitGroup{}
	wg.Add(2)

	// Start a goroutine that waits for program termination. Before the program
	// exits it will flush logs and save the blockchain.
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		log.Info("Saving blockchain and flushing logs...")
		// TODO
		logFile.Sync()
		logFile.Close()
		os.Exit(0)
	}()

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
	log.Infof("Starting listener on %s:%d", cfg.Interface, cfg.Port)
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

	// If the console flag was passed, redirect logs to a file and run the console
	// NOTE: if the log file already exists we will exit with a fatal error here!
	// This should stop people from running multiple Cumulus nodes that will try
	// to log to the same file.
	if cfg.Console {
		logFile, err := os.OpenFile("logfile", os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0755)
		if err != nil {
			log.WithError(err).Fatal("Failed to redirect logs to file")
		}
		log.Info("Redirecting logs to logfile")
		log.SetOutput(logFile)
		go RunConsole(&a)
	}

	if len(config.Target) > 0 {
		// Connect to the target and discover its peers.
		a.ConnectAndDiscover(cfg.Target)
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
}

// RequestHandler is called every time a peer sends us a request message expect
// on peers whos PushHandlers have been overridden.
func (a *App) RequestHandler(req *msg.Request) msg.Response {
	res := msg.Response{ID: req.ID}

	// Build some error types.
	typeErr := msg.NewProtocolError(msg.InvalidResourceType,
		"Invalid resource type")
	notFoundErr := msg.NewProtocolError(msg.ResourceNotFound,
		"Resource not found.")

	switch req.ResourceType {
	case msg.ResourcePeerInfo:
		res.Resource = a.PeerStore.Addrs()
	case msg.ResourceBlock:
		// Block is requested by number.
		blockNumber, ok := req.Params["blockNumber"].(uint32)
		if ok {
			// If its ok, we make try to a copy of it.
			// TODO: Make this CopyBlockByHash.
			blk, err := a.Chain.CopyBlockByIndex(blockNumber)
			if err != nil {
				// Bad index parameter.
				res.Error = notFoundErr
			} else {
				res.Resource = blk
			}
		} else {
			// No index parameter.
			res.Error = notFoundErr
		}
	default:
		// Return err by default.
		res.Error = typeErr
	}

	return res
}

// PushHandler is called every time a peer sends us a Push message except on
// peers whos PushHandlers have been overridden.
func (a *App) PushHandler(push *msg.Push) {
	switch push.ResourceType {
	case msg.ResourceBlock:
		blk, ok := push.Resource.(*blockchain.Block)
		if ok {
			log.Info("Adding block to work queue.")
			blockQueue <- blk
		} else {
			log.Error("Could not cast resource to block.")
		}
	case msg.ResourceTransaction:
		txn, ok := push.Resource.(*blockchain.Transaction)
		if ok {
			log.Info("Adding transaction to work queue.")
			transactionQueue <- txn
		} else {
			log.Error("Could not cast resource to transaction.")
		}
	default:
		// Invalid resource type. Ignore
	}
}

// getLocalChain returns an instance of the blockchain.
func getLocalChain(user *User) *blockchain.BlockChain {
	// TODO: Look for local chain on disk. If doesn't exist, go rummaging
	// around on the internets for one.
	bc := blockchain.BlockChain{
		Blocks: make([]*blockchain.Block, 0),
		Head:   blockchain.NilHash,
	}

	genisisBlock := blockchain.Genesis(user.HotWallet.Wallet.Public(),
		consensus.CurrentTarget(), consensus.StartingBlockReward,
		make([]byte, 0))

	bc.Blocks = append(bc.Blocks, genisisBlock)
	bc.Head = blockchain.HashSum(genisisBlock)
	return &bc
}

// getLocalPool returns an instance of the pool.
func getLocalPool() *pool.Pool {
	// TODO: Look for local pool on disk. If doesn't exist,  make a new one.
	return pool.New()
}

// HandleWork continually collects new work from existing work channels.
func (a *App) HandleWork() {
	log.Debug("Worker waiting for work.")
	for {
		select {
		case work := <-transactionQueue:
			a.HandleTransaction(work)
		case work := <-blockQueue:
			a.HandleBlock(work)
		case <-quitChan:
			return
		}
	}
}

// HandleTransaction handles new instance of TransactionWork.
func (a *App) HandleTransaction(txn *blockchain.Transaction) {
	validTransaction := a.Pool.Push(txn, a.Chain)
	if validTransaction {
		log.Debug("added transaction to pool from address: " + txn.Sender.Repr())
	} else {
		log.Debug("bad transaction rejected from sender: " + txn.Sender.Repr())
	}
}

// HandleBlock handles new instance of BlockWork.
func (a *App) HandleBlock(blk *blockchain.Block) {
	validBlock := a.Pool.Update(blk, a.Chain)

	if validBlock {
		// Append to the chain before requesting the next block so that the block
		// numbers make sense.
		a.Chain.AppendBlock(blk)
		address := a.CurrentUser.Wallet.Public()
		blk := a.Pool.NextBlock(a.Chain, address, a.CurrentUser.BlockSize)
		if miner.IsMining() {
			miner.RestartMiner(a.Chain, blk)
		}
		log.Debug("added blk number %d to chain", blk.BlockNumber)
	}
}

// SyncBlockchain updates the local copy of the blockchain by requesting missing
// blocks from peers. Returns error if we are not connected to any peers.
func (a *App) SyncBlockchain() error {
	prevHead := a.Chain.Blocks[len(a.Chain.Blocks)-1]
	newBlockChan := make(chan *blockchain.Block)
	errChan := make(chan *msg.ProtocolError)

	// Define a handler for responses to our block requests
	blockResponseHandler := func(blockResponse *msg.Response) {
		if blockResponse.Error != nil {
			errChan <- blockResponse.Error
			return
		}

		block, ok := blockResponse.Resource.(blockchain.Block)
		if !ok {
			newBlockChan <- nil
			return
		}

		newBlockChan <- &block
	}

	// Continually request the block after the latest block in our chain until
	// we are totally up to date
	for {
		lastBlock := a.Chain.Blocks[len(a.Chain.Blocks)-1]

		reqParams := map[string]interface{}{
			"lastBlockHash": blockchain.HashSum(lastBlock),
		}

		blockRequest := msg.Request{
			ID:           uuid.New().String(),
			ResourceType: msg.ResourceBlock,
			Params:       reqParams,
		}

		// Get a list of peer addresses. We will use these to retreive peers from
		// the PeerStore that we will then send block requests to
		peerAddrs := a.PeerStore.Addrs()
		if len(peerAddrs) == 0 {
			return errors.New("SyncBlockchain failed: no peers to request blocks from")
		}

		// Pick a peer to send the request to (this won't always be the same one)
		p := a.PeerStore.Get(peerAddrs[0])
		p.Request(blockRequest, blockResponseHandler)

		// Wait for response
		select {
		case newBlock := <-newBlockChan:
			if newBlock == nil {
				// We received a response with no error but an invalid resource
				// Try again
				continue
			}

			valid, validationCode := consensus.VerifyBlock(a.Chain, newBlock)
			if !valid {
				// There is something wrong with this block. Try again
				fields := log.Fields{"validationCode": validationCode}
				log.WithFields(fields).Debug(
					"SyncBlockchain received invalid block")
				continue
			}

			// Valid block. Append it to the chain
			a.Chain.AppendBlock(newBlock)

			if (&newBlock.BlockHeader).Equal(&prevHead.BlockHeader) {
				// Our blockchain is up to date
				return nil
			}

		case err := <-errChan:
			if err.Code == msg.ResourceNotFound {
				// Our chain might be out of sync, roll it back by one block
				// and request the next block
				prevHead = lastBlock
				a.Chain.Blocks = a.Chain.Blocks[:len(a.Chain.Blocks)-1]
			}

			// Some other protocol error occurred. Try again
			log.WithError(err).Debug("SyncBlockchain received error response")
		}
	}
}
