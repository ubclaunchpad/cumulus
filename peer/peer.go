package peer

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/google/uuid"
	"github.com/libp2p/go-libp2p-crypto"
	"github.com/libp2p/go-libp2p-host"
	"github.com/libp2p/go-libp2p-net"
	lpeer "github.com/libp2p/go-libp2p-peer"
	pstore "github.com/libp2p/go-libp2p-peerstore"
	"github.com/libp2p/go-libp2p-swarm"
	bhost "github.com/libp2p/go-libp2p/p2p/host/basic"
	ma "github.com/multiformats/go-multiaddr"
	protoerr "github.com/ubclaunchpad/cumulus/errors"
	"github.com/ubclaunchpad/cumulus/message"
	sn "github.com/ubclaunchpad/cumulus/subnet"
)

const (
	// DefaultPort is the TCP port hosts will communicate over if none is
	// provided
	DefaultPort = 8765
	// CumulusProtocol is the name of the protocol peers communicate over
	CumulusProtocol = "/cumulus/0.0.1"
	// DefaultIP is the IP address new hosts will use if none if provided
	DefaultIP = "127.0.0.1"
	// Timeout is the time after which reads from a stream will timeout
	Timeout = time.Second * 30
)

// MessageHandler is any package that implements HandleMessage
type MessageHandler interface {
	HandleMessage(m message.Message, s net.Stream)
}

// Peer is a cumulus Peer composed of a host
type Peer struct {
	host.Host
	subnet sn.Subnet
}

// New creates a Cumulus host with the given IP addr and TCP port.
// This may throw an error if we fail to create a key pair, a pid, or a new
// multiaddress.
func New(ip string, port int) (*Peer, error) {
	// Generate a key pair for this host. We will only use the pudlic key to
	// obtain a valid host ID.
	// Cannot throw error with given arguments
	priv, pub, _ := crypto.GenerateKeyPair(crypto.RSA, 2048)

	// Obtain Peer ID from public key.
	// Cannot throw error with given argument
	pid, _ := lpeer.IDFromPublicKey(pub)

	// Create a multiaddress (IP address and TCP port for this peer).
	addr, err := ma.NewMultiaddr(fmt.Sprintf("/ip4/%s/tcp/%d", ip, port))
	if err != nil {
		return nil, err
	}

	// Create Peerstore and add host's keys to it (avoids annoying err)
	ps := pstore.NewPeerstore()
	ps.AddPubKey(pid, pub)
	ps.AddPrivKey(pid, priv)

	// Create swarm (this is the interface to the libP2P Network) using the
	// multiaddress, peerID, and peerStore we just created.
	netwrk, err := swarm.NewNetwork(
		context.Background(),
		[]ma.Multiaddr{addr},
		pid,
		ps,
		nil)
	if err != nil {
		return nil, err
	}

	subnet := *sn.New(sn.DefaultMaxPeers)

	// Actually create the host and peer with the network we just set up.
	host := bhost.New(netwrk)
	peer := &Peer{Host: host, subnet: subnet}

	// Build host multiaddress
	hostAddr, _ := ma.NewMultiaddr(fmt.Sprintf("/ipfs/%s",
		host.ID().Pretty()))

	// Now we can build a full multiaddress to reach this host
	// by encapsulating both addresses:
	fullAddr := addr.Encapsulate(hostAddr)
	log.Info("I am ", fullAddr)

	// Add this host's address to its peerstore (avoid's net/identi error)
	ps.AddAddr(pid, fullAddr, pstore.PermanentAddrTTL)

	return peer, nil
}

// Receive is the function that gets called when a remote peer
// opens a new stream with this peer.
// This should be passed as the second argument to SetStreamHandler() after this
// peer is initialized.
func (p *Peer) Receive(s net.Stream) {
	// Get remote peer's full multiaddress
	remoteMA, err := NewMultiaddr(
		s.Conn().RemoteMultiaddr(), s.Conn().RemotePeer())
	if err != nil {
		log.WithError(err).Error(
			"Failed to obtain valid remote peer multiaddress")
		return
	}

	// Add the remote peer to this peer's subnet
	err = p.subnet.AddPeer(remoteMA, s)
	if err != nil {
		// Subnet is full, advertise other available peers and then close
		// the stream
		log.WithError(err).Debug("Peer subnet full. Advertising peers...")
		msg := message.NewResponseMessage(uuid.New().String(),
			nil, p.subnet.StringMultiaddrs())
		msgErr := msg.Write(s)
		if msgErr != nil {
			log.WithError(err).Error("Failed to send ResourcePeerInfo")
		}
		return
	}
	go p.Listen(remoteMA, s)
}

// Connect adds the given multiaddress to p's Peerstore and opens a stream
// with the peer at that multiaddress if the multiaddress is valid, otherwise
// returns error. On success the stream and corresponding multiaddress are
// added to this peer's subnet.
func (p *Peer) Connect(peerma string) (net.Stream, error) {
	peerid, targetAddr, err := extractPeerInfo(peerma)
	if err != nil {
		return nil, err
	}

	// Store the peer's address in this host's PeerStore
	p.Peerstore().AddAddr(peerid, targetAddr, pstore.PermanentAddrTTL)

	// Open a stream with the peer
	stream, err := p.NewStream(context.Background(), peerid,
		CumulusProtocol)
	if err != nil {
		return nil, err
	}
	err = stream.SetDeadline(time.Now().Add(Timeout))
	if err != nil {
		log.WithError(err).Error("Failed to set read deadline on stream")
	}

	mAddr, err := ma.NewMultiaddr(peerma)
	if err != nil {
		stream.Close()
		return nil, err
	}

	err = p.subnet.AddPeer(mAddr, stream)
	if err != nil {
		stream.Close()
	}

	return stream, err
}

// Broadcast sends message to all peers this peer is currently connected to
func (p *Peer) Broadcast(m message.Message) error {
	return errors.New("Function not implemented")
}

// Request sends a request to a remote peer over the given stream.
// Returns response if a response was received, otherwise returns error.
func (p *Peer) Request(req message.Request, s net.Stream) (*message.Response, error) {
	reqMsg := message.New(message.MessageRequest, req)
	err := reqMsg.Write(s)
	if err != nil {
		return nil, err
	}
	resMsg, err := message.Read(s)
	if err != nil {
		return nil, err
	}
	res := resMsg.Payload.(message.Response)
	log.Debugf("Sending request with ResourceType: %d", req.ResourceType)
	return &res, nil
}

// Respond responds to a request from another peer
func (p *Peer) Respond(req message.Request, s net.Stream) {
	var response message.Response

	switch req.ResourceType {
	case message.ResourcePeerInfo:
		response = message.Response{
			ID:       req.ID,
			Error:    nil,
			Resource: p.subnet.StringMultiaddrs(),
		}
		break
	case message.ResourceBlock:
		response = message.Response{
			ID:       req.ID,
			Error:    protoerr.New(protoerr.NotImplemented),
			Resource: nil,
		}
		break
	case message.ResourceTransaction:
		response = message.Response{
			ID:       req.ID,
			Error:    protoerr.New(protoerr.NotImplemented),
			Resource: nil,
		}
		break
	default:
		response = message.Response{
			ID:       req.ID,
			Error:    protoerr.New(protoerr.InvalidResourceType),
			Resource: nil,
		}
	}

	msg := message.New(message.MessageResponse, response)
	err := msg.Write(s)
	if err != nil {
		log.WithError(err).Error("Failed to send reponse")
	} else {
		msgJSON, _ := json.Marshal(msg)
		log.Info("Sending response: \n%s", string(msgJSON))
	}
}

// NewMultiaddr creates a Multiaddress from the given Multiaddress (of the form
// /ip4/<ip address>/tcp/<TCP port>) and the peer id (a hash) and turn them
// into one Multiaddress. Will return error if Multiaddress is invalid.
func NewMultiaddr(iAddr ma.Multiaddr, pid lpeer.ID) (ma.Multiaddr, error) {
	strAddr := iAddr.String()
	strID := pid.Pretty()
	strMA := fmt.Sprintf("%s/ipfs/%s", strAddr, strID)
	mAddr, err := ma.NewMultiaddr(strMA)
	return mAddr, err
}

// HandleMessage responds to a received message
func (p *Peer) HandleMessage(m message.Message, s net.Stream) {
	msgJSON, _ := json.Marshal(m)
	log.Info("Received message: \n%s", string(msgJSON))

	switch m.Type {
	case message.MessageRequest:
		p.Respond(m.Payload.(message.Request), s)
		break
	case message.MessageResponse:
		log.Error("Message response handling not yet implemented")
		break
	case message.MessagePush:
		log.Error("Message push handling not yet implemented")
		break
	default:
		errRes := protoerr.New(protoerr.InvalidMessageType)
		res := message.NewResponseMessage(uuid.New().String(), errRes, nil)
		err := res.Write(s)
		if err != nil {
			log.WithError(err).Error("Failed to handle message")
		}
	}
}

// Listen listens for messages over the stream and responds to them, closing
// the given stream and removing the remote peer from this peer's subnet when
// done. This should be run as a goroutine.
func (p *Peer) Listen(remoteMA ma.Multiaddr, s net.Stream) {
	defer s.Close()
	defer p.subnet.RemovePeer(remoteMA)
	for s != nil {
		err := s.SetDeadline(time.Now().Add(Timeout))
		if err != nil {
			log.WithError(err).Error("Failed to set read deadline on stream")
		}
		msg, err := message.Read(s)
		if err != nil {
			log.WithError(err).Error("Error reading from the stream")
			return
		}
		p.HandleMessage(*msg, s)
	}
}

// ExtractPeerInfo extracts the peer ID and multiaddress from the
// given multiaddress.
// Returns peer ID (esentially 46 character hash created by the peer)
// and the peer's multiaddress in the form /ip4/<peer IP>/tcp/<CumulusPort>.
func extractPeerInfo(peerma string) (lpeer.ID, ma.Multiaddr, error) {
	log.Debug("Extracting peer info from ", peerma)

	ipfsaddr, err := ma.NewMultiaddr(peerma)
	if err != nil {
		return "-", nil, err
	}

	// Cannot throw error when passed P_IPFS
	pid, err := ipfsaddr.ValueForProtocol(ma.P_IPFS)
	if err != nil {
		return "-", nil, err
	}

	// Cannot return error if no error was returned in ValueForProtocol
	peerid, _ := lpeer.IDB58Decode(pid)

	// Decapsulate the /ipfs/<peerID> part from the target
	// /ip4/<a.b.c.d>/ipfs/<peer> becomes /ip4/<a.b.c.d>
	targetPeerAddr, err := ma.NewMultiaddr(
		fmt.Sprintf("/ipfs/%s", lpeer.IDB58Encode(peerid)))
	if err != nil {
		return "-", nil, err
	}

	trgtAddr := ipfsaddr.Decapsulate(targetPeerAddr)

	return peerid, trgtAddr, nil
}
