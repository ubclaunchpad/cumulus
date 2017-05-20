package peer

import (
	"bufio"
	"context"
	"errors"
	"fmt"

	crypto "github.com/libp2p/go-libp2p-crypto"
	host "github.com/libp2p/go-libp2p-host"
	net "github.com/libp2p/go-libp2p-net"
	lpeer "github.com/libp2p/go-libp2p-peer"
	pstore "github.com/libp2p/go-libp2p-peerstore"
	swarm "github.com/libp2p/go-libp2p-swarm"
	bhost "github.com/libp2p/go-libp2p/p2p/host/basic"
	ma "github.com/multiformats/go-multiaddr"
	log "github.com/sirupsen/logrus"
	msg "github.com/ubclaunchpad/cumulus/message"
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
)

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
	defer s.Close()

	// Get remote peer's full multiaddress
	remoteMA, err := makeMultiaddr(
		s.Conn().RemoteMultiaddr(), s.Conn().RemotePeer())
	if err != nil {
		log.Fatal("Failed to obtain valid remote peer multiaddress")
	}

	// Add the remote peer to this peer's subnet
	err = p.subnet.AddPeer(remoteMA, s)
	if err != nil {
		// Subnet is full, advertise other available peers and then close
		// the stream
		log.Debug("Peer subnet full. Advertising peers...")
		p.advertisePeers(s)
		return
	}
	defer p.subnet.RemovePeer(remoteMA)

	buf := bufio.NewReader(s)
	strMsg, err := buf.ReadString('\n') // TODO: set timeout here
	if err != nil {
		log.Error(err)
		return
	}

	// Turn the string into a message we can deal with
	message, err := msg.FromString(strMsg)
	if err != nil {
		log.Error(err)
		return
	}

	log.Debugf("Peer %s message:\n%s", p.ID(), strMsg)

	// Respond to message
	p.handleMessage(*message, s)
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

	log.Debug("Connected to Cumulus Peer:")
	log.Debugf("Peer ID: %s", peerid.Pretty())
	log.Debug("Peer Address:", targetAddr)

	// Open a stream with the peer
	stream, err := p.NewStream(context.Background(), peerid,
		CumulusProtocol)
	if err != nil {
		return nil, err
	}

	mAddr, err := ma.NewMultiaddr(peerma)
	if err != nil {
		stream.Close()
		return nil, err
	}

	err = p.subnet.AddPeer(mAddr, stream)
	if err != nil {
		stream.Close()
		return nil, err
	}

	return stream, err
}

// Broadcast sends message to all peers this peer is currently connected to
func (p *Peer) Broadcast(m msg.Message) error {
	return errors.New("Function not implemented")
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

// advertisePeers writes messages into the given stream advertising the
// multiaddress of each peer in this peer's subnet.
func (p *Peer) advertisePeers(s net.Stream) {
	mAddrs := p.subnet.Multiaddrs()
	log.Debug("Peers on this subnet: ")
	for mAddr := range mAddrs {
		mAddrString := string(mAddr)
		log.Debug("\t", mAddrString)
		message, err := msg.New([]byte(mAddrString), msg.PeerInfo)
		if err != nil {
			log.Error("Failed to create message")
			return
		}
		msgBytes, err := message.Bytes()
		if err != nil {
			log.Error("Failed to marshal message")
			return
		}
		_, err = s.Write(msgBytes)
		if err != nil {
			log.Errorf("Failed to send message to %s", string(mAddr))
		}
	}
}

// makeMultiaddr creates a Multiaddress from the given Multiaddress (of the form
// /ip4/<ip address>/tcp/<TCP port>) and the peer id (a hash) and turn them
// into one Multiaddress. Will return error if Multiaddress is invalid.
func makeMultiaddr(iAddr ma.Multiaddr, pid lpeer.ID) (ma.Multiaddr, error) {
	strAddr := iAddr.String()
	strID := pid.Pretty()
	strMA := fmt.Sprintf("%s/ipfs/%s", strAddr, strID)
	mAddr, err := ma.NewMultiaddr(strMA)
	return mAddr, err
}

func (p *Peer) handleMessage(m msg.Message, s net.Stream) {
	switch m.Type() {
	case msg.RequestPeerInfo:
		p.advertisePeers(s)
		break
	default:
		// Do nothing. WHEOOO!
	}
}
