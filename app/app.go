package app

import (
	"fmt"

	log "github.com/Sirupsen/logrus"
	"github.com/google/uuid"

	"github.com/ubclaunchpad/cumulus/blockchain"
	"github.com/ubclaunchpad/cumulus/conf"
	"github.com/ubclaunchpad/cumulus/conn"
	"github.com/ubclaunchpad/cumulus/msg"
	"github.com/ubclaunchpad/cumulus/peer"
)

var (
	config *conf.Config
	// TODO peer store once it's merged in
	chain *blockchain.BlockChain
)

// Run sets up and starts a new Cumulus node with the
// given configuration.
func Run(cfg conf.Config) {
	log.Info("Starting Cumulus node")
	config = &cfg

	// Set Peer default Push and Request handlers. These functions will handle
	// request and push messages from all peers we connect to unless overridden
	// for specific peers by calls like p.SetRequestHandler(someHandler)
	peer.SetDefaultPushHandler(PushHandler)
	peer.SetDefaultRequestHandler(RequestHandler)

	// Start listening on the given interface and port so we can receive
	// conenctions from other peers
	log.Infof("Starting listener on %s:%d", cfg.Interface, cfg.Port)
	peer.LocalAddr = fmt.Sprintf("%s:%d", cfg.Interface, cfg.Port)
	go func() {
		err := conn.Listen(fmt.Sprintf("%s:%d", cfg.Interface, cfg.Port), peer.ConnectionHandler)
		if err != nil {
			log.WithError(err).Fatalf("Failed to listen on %s:%d", cfg.Interface, cfg.Port)
		}
	}()

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
	select {}
}

// RequestHandler is called every time a peer sends us a request message expect
// on peers whos PushHandlers have been overridden.
func RequestHandler(req *msg.Request) msg.Response {
	res := msg.Response{ID: req.ID}

	switch req.ResourceType {
	case msg.ResourcePeerInfo:
		res.Resource = peer.PStore.Addrs()
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

	// Ask target for its peers
	// Connect to these peers until we have enough peers
	// Download the blockchain
}
