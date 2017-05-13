package peer

import (
	"bufio"
	"context"
	"fmt"

	log "github.com/Sirupsen/logrus"
	crypto "github.com/libp2p/go-libp2p-crypto"
	host "github.com/libp2p/go-libp2p-host"
	net "github.com/libp2p/go-libp2p-net"
	lpeer "github.com/libp2p/go-libp2p-peer"
	pstore "github.com/libp2p/go-libp2p-peerstore"
	swarm "github.com/libp2p/go-libp2p-swarm"
	bhost "github.com/libp2p/go-libp2p/p2p/host/basic"
	ma "github.com/multiformats/go-multiaddr"
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

	// Actually create the host and peer with the network we just set up.
	host := bhost.New(netwrk)
	peer := &Peer{Host: host}

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
// opens a new stream with the host that SetStreamHandler() is called on.
// This should be passed as the second argument to SetStreamHandler().
// We may want to implement another type of StreamHandler in the future.
func (p *Peer) Receive(s net.Stream) {
	log.Debug("Setting basic stream handler.")
	defer s.Close()
	p.doCumulus(s)
}

// Communicate with peers.
// TODO: Update this to do something useful. For now it just reads from the
// stream and writes back what it read.
func (p *Peer) doCumulus(s net.Stream) {
	buf := bufio.NewReader(s)
	str, err := buf.ReadString('\n')
	if err != nil {
		log.Error(err)
		return
	}

	log.Debugf("Peer %s read: %s", p.ID(), str)

	_, err = s.Write([]byte(str))
	if err != nil {
		log.Error(err)
	}
}

// ExtractPeerInfo extracts the peer ID and multiaddress from the
// given multiaddress.
// Returns peer ID (esentially 46 character hash created by the peer)
// and the peer's multiaddress in the form /ip4/<peer IP>/tcp/<CumulusPort>.
func ExtractPeerInfo(peerma string) (lpeer.ID, ma.Multiaddr, error) {
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

// Connect adds the given multiaddress to p's Peerstore and opens a stream
// with the peer at that multiaddress if the multiaddress is valid, otherwise
// returns error.
func (p *Peer) Connect(peerma string) (net.Stream, error) {
	peerid, targetAddr, err := ExtractPeerInfo(peerma)
	if err != nil {
		return nil, err
	}

	// Store the peer's address in this host's PeerStore
	p.Peerstore().AddAddr(peerid, targetAddr, pstore.PermanentAddrTTL)

	log.Debug("Connected to Cumulus Peer:")
	log.Debug("Peer ID:", peerid.Pretty())
	log.Debug("Peer Address:", targetAddr)

	// Open a stream with the peer
	stream, err := p.NewStream(context.Background(), peerid,
		CumulusProtocol)

	return stream, err
}
