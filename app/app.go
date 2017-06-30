package app

import (
	"bufio"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	log "github.com/Sirupsen/logrus"
	"github.com/google/uuid"

	"github.com/ubclaunchpad/cumulus/blockchain"
	"github.com/ubclaunchpad/cumulus/conf"
	"github.com/ubclaunchpad/cumulus/conn"
	"github.com/ubclaunchpad/cumulus/console"
	"github.com/ubclaunchpad/cumulus/msg"
	"github.com/ubclaunchpad/cumulus/peer"
)

var (
	config  *conf.Config
	chain   *blockchain.BlockChain
	logFile = os.Stdout
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
		log.Info("Saving blockchain...")
		// TODO
		os.Exit(0)
	}()

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
		err := conn.Listen(fmt.Sprintf("%s:%d", cfg.Interface, cfg.Port), peer.ConnectionHandler)
		if err != nil {
			log.WithError(err).Fatalf("Failed to listen on %s:%d", cfg.Interface, cfg.Port)
		}
	}()

	// If the console flag was passed, redirect logs to a file and run the console
	if cfg.Console {
		logFile, err := os.OpenFile("logfile", os.O_WRONLY|os.O_CREATE, 0755)
		if err != nil {
			log.WithError(err).Fatal("Failed to redirect logs to log file")
		}
		log.Warn("Redirecting logs to logfile")

		logWriter := bufio.NewWriter(logFile)
		log.SetOutput(logWriter)
		go console.Run()
	}

	// If a target peer was supplied, connect to it and try discover and connect
	// to its peers
	if len(cfg.Target) > 0 {
		peerInfoRequest := msg.Request{
			ID:           uuid.New().String(),
			ResourceType: msg.ResourcePeerInfo,
		}

		log.Infof("Dialing target %s", cfg.Target)
		c, err := conn.Dial(cfg.Target)
		if err != nil {
			log.WithError(err).Fatalf("Failed to connect to target")
		}
		peer.ConnectionHandler(c)
		p := peer.PStore.Get(c.RemoteAddr().String())
		p.Request(peerInfoRequest, peer.PeerInfoHandler)
	}

	// Try maintain as close to peer.MaxPeers connections as possible while this
	// peer is running
	go peer.MaintainConnections()
	select {} // Hang main thread. Everything happens in goroutines from here
}

// RequestHandler is called every time a peer sends us a request message expect
// on peers whos PushHandlers have been overridden.
func RequestHandler(req *msg.Request) msg.Response {
	res := msg.Response{ID: req.ID}

	switch req.ResourceType {
	case msg.ResourcePeerInfo:
		res.Resource = peer.PStore.Addrs()
		log.Debugf("Returning PeerInfo %s", res.Resource)
	case msg.ResourceBlock, msg.ResourceTransaction:
		res.Error = msg.NewProtocolError(msg.NotImplemented,
			"Block and Transaction requests are not yet implemented on this peer")
	default:
		res.Error = msg.NewProtocolError(msg.InvalidResourceType,
			"Invalid resource type")
	}

	return res
}

// PushHandler is called every time a peer sends us a Push message expect on
// peers whos PushHandlers have been overridden.
func PushHandler(push *msg.Push) {
	switch push.ResourceType {
	case msg.ResourceBlock:
	case msg.ResourceTransaction:
	default:
		// Invalid resource type. Ignore
	}
}
