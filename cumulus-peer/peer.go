package cumulusPeer

import (
	"context"
	"fmt"
	"log"

	crypto "github.com/libp2p/go-libp2p-crypto"
	host "github.com/libp2p/go-libp2p-host"
	peer "github.com/libp2p/go-libp2p-peer"
	pstore "github.com/libp2p/go-libp2p-peerstore"
	swarm "github.com/libp2p/go-libp2p-swarm"
	ma "github.com/multiformats/go-multiaddr"
	bhost "github.com/libp2p/go-libp2p/p2p/host/basic"
)

// Cumulus peers communicate over this TCP port
const CumulusPort int = 8765

// Create a Cumulus host.
// This may throw an error if we fail to create a key pair, a pid, or a new
// multiaddress.
func MakeHost() (host.Host, error) {
    // Generate a key pair for this host. We will only use the pudlic key to
    // obtain a valid host ID.
    _, pub, err := crypto.GenerateKeyPair(crypto.RSA, 2048)
    if err != nil {
        return nil, err
    }

    // Obtain Peer ID from public key
    pid, err := peer.IDFromPublicKey(pub)
    if err != nil {
        return nil, err
    }

    // Create a multiaddress (IP address and TCP port for this peer)
    addr, err := ma.NewMultiaddr(
        fmt.Sprintf("/ip4/127.0.0.1/tcp/%d",
        CumulusPort))
    if err != nil {
		return nil, err
	}

    // Create a peerstore (this stores information about other peers in the
    // Cumulus network)
	ps := pstore.NewPeerstore()

    // Create swarm (this is the interface to the libP2P Network) using the
    // multiaddress, peerID, and peerStore we just created
	netwrk, err := swarm.NewNetwork(
		context.Background(),
		[]ma.Multiaddr{addr},
		pid,
		ps,
		nil)
    if err != nil {
        return nil, err
    }

	basicHost := bhost.New(netwrk)

	// Build host multiaddress
	hostAddr, _ := ma.NewMultiaddr(fmt.Sprintf("/ipfs/%s",
        basicHost.ID().Pretty()))

	// Now we can build a full multiaddress to reach this host
	// by encapsulating both addresses:
	fullAddr := addr.Encapsulate(hostAddr)
	log.Printf("I am %s\n", fullAddr)

	return basicHost, nil
}
