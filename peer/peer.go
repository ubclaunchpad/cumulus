package peer

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/libp2p/go-libp2p-crypto"
	"github.com/libp2p/go-libp2p-host"
	net "github.com/libp2p/go-libp2p-net"
	lpeer "github.com/libp2p/go-libp2p-peer"
	"github.com/libp2p/go-libp2p-peerstore"
	"github.com/libp2p/go-libp2p-swarm"
	bhost "github.com/libp2p/go-libp2p/p2p/host/basic"
	multiaddr "github.com/multiformats/go-multiaddr"
	"github.com/ubclaunchpad/cumulus/message"
	"github.com/ubclaunchpad/cumulus/stream"
	"github.com/ubclaunchpad/cumulus/subnet"
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

// Peer is a cumulus Peer composed of a host
type Peer struct {
	host.Host
	subnet subnet.Subnet
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
	addr, err := multiaddr.NewMultiaddr(fmt.Sprintf("/ip4/%s/tcp/%d", ip, port))
	if err != nil {
		return nil, err
	}

	// Create Peerstore and add host's keys to it (avoids annoying err)
	ps := peerstore.NewPeerstore()
	ps.AddPubKey(pid, pub)
	ps.AddPrivKey(pid, priv)

	// Create swarm (this is the interface to the libP2P Network) using the
	// multiaddress, peerID, and peerStore we just created.
	network, err := swarm.NewNetwork(
		context.Background(),
		[]multiaddr.Multiaddr{addr},
		pid,
		ps,
		nil)
	if err != nil {
		return nil, err
	}

	sn := *subnet.New(subnet.DefaultMaxPeers)

	// Actually create the host and peer with the network we just set up.
	host := bhost.New(network)
	peer := &Peer{
		Host:   host,
		subnet: sn,
	}

	// Build host multiaddress
	hostAddr, _ := multiaddr.NewMultiaddr(fmt.Sprintf("/ipfs/%s",
		host.ID().Pretty()))

	// Now we can build a full multiaddress to reach this host
	// by encapsulating both addresses:
	fullAddr := addr.Encapsulate(hostAddr)
	log.Info("I am ", fullAddr)

	// Add this host's address to its peerstore (avoid's net/identi error)
	ps.AddAddr(pid, fullAddr, peerstore.PermanentAddrTTL)
	return peer, nil
}

// Receive is the function that gets called when a remote peer
// opens a new stream with this peer.
// This should be passed as the second argument to SetStreamHandler() after this
// peer is initialized.
func (p *Peer) Receive(s net.Stream) {
	// Get remote peer's full multiaddress
	ma, err := NewMultiaddr(s.Conn().RemoteMultiaddr(), s.Conn().RemotePeer())
	if err != nil {
		log.WithError(err).Error("Failed to obtain valid remote peer multiaddress")
		return
	}

	peerstream := *stream.New(s)

	// Add the remote peer to this peer's subnet
	err = p.subnet.AddPeer(ma.String(), peerstream)
	if err != nil {
		// Subnet is full, advertise other available peers and then close
		// the stream
		log.WithError(err).Debug("Peer subnet full. Advertising peers...")
		msg := message.Push{
			ResourceType: message.ResourcePeerInfo,
			Resource:     p.subnet.Multiaddrs(),
		}
		msgErr := msg.Write(s)
		if msgErr != nil {
			log.WithError(err).Error("Failed to send ResourcePeerInfo")
		}
		s.Close()
		return
	}
	go p.Listen(ma.String(), peerstream)
}

// Connect adds the given multiaddress to p's Peerstore and opens a stream
// with the peer at that multiaddress if the multiaddress is valid, otherwise
// returns error. On success the stream and corresponding multiaddress are
// added to this peer's subnet.
func (p *Peer) Connect(sma string) (*stream.Stream, error) {
	pid, targetAddr, err := extractPeerInfo(sma)
	if err != nil {
		return nil, err
	}

	// Store the peer's address in this host's PeerStore
	p.Peerstore().AddAddr(pid, targetAddr, peerstore.PermanentAddrTTL)

	// Open a stream with the peer
	s, err := p.NewStream(context.Background(), pid, CumulusProtocol)
	if err != nil {
		return nil, err
	}
	err = s.SetDeadline(time.Now().Add(Timeout))
	if err != nil {
		log.WithError(err).Error("Failed to set read deadline on stream")
		return nil, err
	}

	// Make a stream.Stream out of a net.Stream (sorry)
	peerstream := *stream.New(s)
	err = p.subnet.AddPeer(sma, peerstream)
	if err != nil {
		s.Close()
		return nil, err
	}

	go p.Listen(sma, peerstream)
	return &peerstream, nil
}

// Broadcast sends message to all peers this peer is currently connected to
func (p *Peer) Broadcast(m message.Message) error {
	var err error
	for _, ma := range p.subnet.Multiaddrs() {
		err = m.Write(p.subnet.Stream(ma))
		if err != nil {
			log.WithError(err).Error("Failed to send broadcast message to peer")
		}
	}
	return err
}

// Request sends a request to a remote peer over the given stream.
// Returns response if a response was received, otherwise returns error.
// You should typically run this function in a goroutine
func (p *Peer) Request(req message.Request, s stream.Stream) (*message.Response, error) {
	// Set up a listen channel to listen for the response (remove when done)
	lchan := s.NewListener(req.ID)
	defer s.RemoveListener(req.ID)

	// Send request
	err := req.Write(s)
	if err != nil {
		return nil, err
	}

	// Receive response or timeout
	select {
	case res := <-lchan:
		return res, nil
	case <-time.After(Timeout):
		return nil, errors.New("Timed out waiting for response")
	}
}

// Respond responds to a request from another peer
func (p *Peer) Respond(req *message.Request, s stream.Stream) {
	response := message.Response{ID: req.ID}

	switch req.ResourceType {
	case message.ResourcePeerInfo:
		response.Resource = p.subnet.Multiaddrs()
		break
	case message.ResourceBlock:
	case message.ResourceTransaction:
		response.Error = message.NewProtocolError(message.NotImplemented,
			"Functionality not implemented on this peer")
		break
	default:
		response.Error = message.NewProtocolError(message.InvalidResourceType,
			"Invalid resource type")
	}

	err := response.Write(s)
	if err != nil {
		log.WithError(err).Error("Failed to send reponse")
	} else {
		msgJSON, _ := json.Marshal(response)
		log.Infof("Sending response: \n%s", string(msgJSON))
	}
}

// NewMultiaddr creates a Multiaddress from the given Multiaddress (of the form
// /ip4/<ip address>/tcp/<TCP port>) and the peer id (a hash) and turn them
// into one Multiaddress. Will return error if Multiaddress is invalid.
func NewMultiaddr(iAddr multiaddr.Multiaddr, pid lpeer.ID) (multiaddr.Multiaddr, error) {
	strAddr := iAddr.String()
	strID := pid.Pretty()
	strMA := fmt.Sprintf("%s/ipfs/%s", strAddr, strID)
	mAddr, err := multiaddr.NewMultiaddr(strMA)
	return mAddr, err
}

// HandleMessage responds to a received message
func (p *Peer) handleMessage(m message.Message, s stream.Stream) {
	msgJSON, err := json.Marshal(m)
	if err == nil {
		log.Infof("Received message: \n%s", string(msgJSON))
	}

	switch m.Type() {
	case message.MessageRequest:
		// Respond to the request by sending request resource
		p.Respond(m.(*message.Request), s)
		break
	case message.MessageResponse:
		// Pass the response to the goroutine that requested it
		res := m.(*message.Response)
		lchan := s.Listener(res.ID)
		if lchan != nil {
			log.Debug("Found listener channel for response")
			lchan <- res
		}
		break
	case message.MessagePush:
		// Handle data from push message
		log.Error("Message push handling not yet implemented")
		break
	default:
		// Invalid message type, ignore
		log.Errorln("Received message with invalid type")
	}
}

// Listen listens for messages over the stream and responds to them, closing
// the given stream and removing the remote peer from this peer's subnet when
// done. This should be run as a goroutine.
func (p *Peer) Listen(sma string, s stream.Stream) {
	defer s.Close()
	defer p.subnet.RemovePeer(sma)
	for {
		err := s.SetDeadline(time.Now().Add(Timeout))
		if err != nil {
			log.WithError(err).Error("Failed to set read deadline on stream")
			return
		}
		msg, err := message.Read(s)
		if err != nil {
			log.WithError(err).Error("Error reading from the stream")
			return
		}
		log.Debug("Listener received message")
		go p.handleMessage(msg, s)
	}
}

// ExtractPeerInfo extracts the peer ID and multiaddress from the
// given multiaddress.
// Returns peer ID (esentially 46 character hash created by the peer)
// and the peer's multiaddress in the form /ip4/<peer IP>/tcp/<CumulusPort>.
func extractPeerInfo(sma string) (lpeer.ID, multiaddr.Multiaddr, error) {
	log.Debug("Extracting peer info from ", sma)

	ipfsaddr, err := multiaddr.NewMultiaddr(sma)
	if err != nil {
		return "-", nil, err
	}

	// Cannot throw error when passed P_IPFS
	pid, err := ipfsaddr.ValueForProtocol(multiaddr.P_IPFS)
	if err != nil {
		return "-", nil, err
	}

	// Cannot return error if no error was returned in ValueForProtocol
	peerid, _ := lpeer.IDB58Decode(pid)

	// Decapsulate the /ipfs/<peerID> part from the target
	// /ip4/<a.b.c.d>/ipfs/<peer> becomes /ip4/<a.b.c.d>
	targetPeerAddr, err := multiaddr.NewMultiaddr(
		fmt.Sprintf("/ipfs/%s", lpeer.IDB58Encode(peerid)))
	if err != nil {
		return "-", nil, err
	}

	trgtAddr := ipfsaddr.Decapsulate(targetPeerAddr)
	return peerid, trgtAddr, nil
}
