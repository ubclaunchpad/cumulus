package app

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	log "github.com/Sirupsen/logrus"
	"github.com/google/uuid"

	"github.com/ubclaunchpad/cumulus/blockchain"
	"github.com/ubclaunchpad/cumulus/conf"
	"github.com/ubclaunchpad/cumulus/conn"
	"github.com/ubclaunchpad/cumulus/msg"
	"github.com/ubclaunchpad/cumulus/peer"
	"github.com/ubclaunchpad/cumulus/pool"
)

var (
	config  *conf.Config
	chain   *blockchain.BlockChain
	logFile = os.Stdout
	// A reference to the transaction pool
	tpool *pool.Pool
)

// Run sets up and starts a new Cumulus node with the
// given configuration.
func Run(cfg conf.Config) {
	log.Info("Starting Cumulus node")
	config = &cfg

	// Set logging level
	if cfg.Verbose {
		log.SetLevel(log.DebugLevel)
	}

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
	// stream in. We should first initalize our pool, workers to handle
	// incoming messages.
	initializeNode()

	// Set Peer default Push and Request handlers. These functions will handle
	// request and push messages from all peers we connect to unless overridden
	// for specific peers by calls like p.SetRequestHandler(someHandler)
	peer.SetDefaultPushHandler(PushHandler)
	peer.SetDefaultRequestHandler(RequestHandler)

	// Start listening on the given interface and port so we can receive
	// conenctions from other peers
	log.Infof("Starting listener on %s:%d", cfg.Interface, cfg.Port)
	peer.ListenAddr = fmt.Sprintf("%s:%d", cfg.Interface, cfg.Port)
	go func() {
		address := fmt.Sprintf("%s:%d", cfg.Interface, cfg.Port)
		err := conn.Listen(address, peer.ConnectionHandler)
		if err != nil {
			log.WithError(
				err,
			).Fatalf("Failed to listen on %s:%d", cfg.Interface, cfg.Port)
		}
	}()

	// If the console flag was passed, redirect logs to a file and run the console
	if cfg.Console {
		logFile, err := os.OpenFile("logfile", os.O_WRONLY|os.O_CREATE, 0755)
		if err != nil {
			log.WithError(err).Fatal("Failed to redirect logs to log file")
		}
		log.Warn("Redirecting logs to logfile")
		log.SetOutput(logFile)
		go RunConsole()
	}

	if len(config.Target) > 0 {
		// Connect to the target and discover its peers.
		ConnectAndDiscover(cfg.Target)
	}

	// Try maintain as close to peer.MaxPeers connections as possible while this
	// peer is running
	go peer.MaintainConnections()

	// Request the blockchain.
	if chain == nil {
		log.Info("Request blockchain from peers not yet implemented.")
		initializeChain()
	}

	// Return to command line.
	select {} // Hang main thread. Everything happens in goroutines from here
}

// ConnectAndDiscover tries to connect to a target and discover its peers.
func ConnectAndDiscover(target string) {
	peerInfoRequest := msg.Request{
		ID:           uuid.New().String(),
		ResourceType: msg.ResourcePeerInfo,
	}

	log.Infof("Dialing target %s", target)
	c, err := conn.Dial(target)
	if err != nil {
		log.WithError(err).Fatalf("Failed to connect to target")
	}
	peer.ConnectionHandler(c)
	p := peer.PStore.Get(c.RemoteAddr().String())
	p.Request(peerInfoRequest, peer.PeerInfoHandler)
}

// RequestHandler is called every time a peer sends us a request message except
// on peers whos RequestHandlers have been overridden.
func RequestHandler(req *msg.Request) msg.Response {
	res := msg.Response{ID: req.ID}

	// Build some error types.
	typeErr := msg.NewProtocolError(msg.InvalidResourceType,
		"Invalid resource type")
	notFoundErr := msg.NewProtocolError(msg.ResourceNotFound,
		"Resource not found.")

	switch req.ResourceType {
	case msg.ResourcePeerInfo:
		res.Resource = peer.PStore.Addrs()
	case msg.ResourceBlock:
		// Block is requested by number.
		blockNumber, ok := req.Params["blockNumber"].(uint32)
		if ok {
			// If its ok, we make try to a copy of it.
			blk, err := chain.CopyBlockByIndex(blockNumber)
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
func PushHandler(push *msg.Push) {
	switch push.ResourceType {
	case msg.ResourceBlock:
		blk, ok := push.Resource.(*blockchain.Block)
		if ok {
			log.Info("Adding block to work queue.")
			BlockWorkQueue <- BlockWork{blk, nil}
		} else {
			log.Error("Could not cast resource to block.")
		}
	case msg.ResourceTransaction:
		txn, ok := push.Resource.(*blockchain.Transaction)
		if ok {
			log.Info("Adding transaction to work queue.")
			TransactionWorkQueue <- TransactionWork{txn, nil}
		} else {
			log.Error("Could not cast resource to transaction.")
		}
	default:
		// Invalid resource type. Ignore
	}
}

// initializeNode creates a transaction pool, workers and queues to handle
// incoming messages.
func initializeNode() {
	tpool = pool.New()
	intializeQueues()
	initializeWorkers()
}

// intializeQueues makes all necessary queues.
func intializeQueues() {
	BlockWorkQueue = make(chan BlockWork, BlockQueueSize)
	TransactionWorkQueue = make(chan TransactionWork, TransactionQueueSize)
	QuitChan = make(chan int)
}

// initializeWorkers kicks off workers to handle incoming requests.
func initializeWorkers() {
	for i := 0; i < nWorkers; i++ {
		log.WithFields(log.Fields{"id": i}).Debug("Starting worker. ")
		worker := NewWorker(i)
		worker.Start()
		workers[i] = &worker
	}
}

// initializeChain creates the blockchain for the node.
func initializeChain() {
	chain, _ = blockchain.NewValidTestChainAndBlock()
	// TODO: Check if chain exists on disk.
	// TODO: If not, request chain from peers.
}

// killWorkers kills all workers.
func killWorkers() {
	for i := 0; i < nWorkers; i++ {
		QuitChan <- i
		workers[i] = nil
	}
}
