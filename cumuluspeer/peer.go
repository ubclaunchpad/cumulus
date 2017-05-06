package cumuluspeer

import (
	"bufio"
	"context"
	"fmt"

	log "github.com/sirupsen/logrus"
	crypto "github.com/libp2p/go-libp2p-crypto"
	host "github.com/libp2p/go-libp2p-host"
	peer "github.com/libp2p/go-libp2p-peer"
	net "github.com/libp2p/go-libp2p-net"
	pstore "github.com/libp2p/go-libp2p-peerstore"
	swarm "github.com/libp2p/go-libp2p-swarm"
	ma "github.com/multiformats/go-multiaddr"
	bhost "github.com/libp2p/go-libp2p/p2p/host/basic"
)

const (
	CumulusPort 	= 8765 // Cumulus peers communicate over this TCP port
	CumulusProtocol	= "/cumulate/0.0.1" // Cumulus communication protocol
)

// A basic StreamHandler for a very basic host.
// This should be passed as the second argument to SetStreamHandler().
// This is called when a remote peer opens a new stream with the host that
// SetStreamHandler() is called on.
// We may want to implement another type of StreamHandler in the future.
func BasicStreamHandler(s net.Stream) {
	log.Info("Setting basic stream handler.")
	defer s.Close()
	doCumulate(s)
}

// Communicate with peers.
// TODO: Update this to do something useful. For now it just reads from the
// stream and writes back what it read.
func doCumulate(s net.Stream) {
	buf := bufio.NewReader(s)
	str, err := buf.ReadString('\n')
	if err != nil {
		log.Error(err)
		return
	}

	log.Info("Read: %s", str)
	_, err = s.Write([]byte(str))
	if err != nil {
		log.Error(err)
		return
	}

	log.Info("Done now. Bye!")
}

// Create a Cumulus host.
// This may throw an error if we fail to create a key pair, a pid, or a new
// multiaddress.
func MakeBasicHost(port int) (host.Host, pstore.Peerstore, error) {
	// Make sure we received a valid port number

    // Generate a key pair for this host. We will only use the pudlic key to
    // obtain a valid host ID.
    _, pub, err := crypto.GenerateKeyPair(crypto.RSA, 2048)
    if err != nil {
        return nil, nil, err
    }

    // Obtain Peer ID from public key.
    pid, err := peer.IDFromPublicKey(pub)
    if err != nil {
        return nil, nil, err
    }

    // Create a multiaddress (IP address and TCP port for this peer).
    addr, err := ma.NewMultiaddr(fmt.Sprintf("/ip4/127.0.0.1/tcp/%d", port))
    if err != nil {
		return nil, nil, err
	}

    // Create a peerstore (this stores information about other peers in the
    // Cumulus network).
	ps := pstore.NewPeerstore()

    // Create swarm (this is the interface to the libP2P Network) using the
    // multiaddress, peerID, and peerStore we just created.
	netwrk, err := swarm.NewNetwork(
		context.Background(),
		[]ma.Multiaddr{addr},
		pid,
		ps,
		nil)
    if err != nil {
        return nil, nil, err
    }

	// Actually create the host with the network we just set up.
	basicHost := bhost.New(netwrk)

	// Build host multiaddress
	hostAddr, _ := ma.NewMultiaddr(fmt.Sprintf("/ipfs/%s",
        basicHost.ID().Pretty()))

	// Now we can build a full multiaddress to reach this host
	// by encapsulating both addresses:
	fullAddr := addr.Encapsulate(hostAddr)
	log.Info("I am ", fullAddr)

	// Add this host's address to its peerstore (avoid's net/identi error)
	ps.AddAddr(pid, fullAddr, pstore.PermanentAddrTTL)

	return basicHost, ps, nil
}

// Extracts target's the peer ID and multiaddress from the given multiaddress.
// Returns peer ID (esentially 46 character hash created by the peer)
// and the peer's multiaddress in the form /ip4/<peer IP>/tcp/<CumulusPort>.
func ExtractPeerInfo(peerma string) (peer.ID, ma.Multiaddr, error) {
	ipfsaddr, err := ma.NewMultiaddr(peerma)
	if err != nil {
		log.Error(err)
		return "-", nil, err
	}

	pid, err := ipfsaddr.ValueForProtocol(ma.P_IPFS)
	if err != nil {
		log.Error(err)
		return "-", nil, err
	}

	peerid, err := peer.IDB58Decode(pid)
	if err != nil {
		log.Error(err)
		return "-", nil, err
	}

	// Decapsulate the /ipfs/<peerID> part from the target
	// /ip4/<a.b.c.d>/ipfs/<peer> becomes /ip4/<a.b.c.d>
	targetPeerAddr, _ := ma.NewMultiaddr(
		fmt.Sprintf("/ipfs/%s", peer.IDB58Encode(peerid)))
	trgtAddr := ipfsaddr.Decapsulate(targetPeerAddr)

	return peerid, trgtAddr, nil
}
