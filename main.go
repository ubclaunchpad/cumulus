package main

import (
    "context"
    "flag"
    "io/ioutil"

    log "github.com/sirupsen/logrus"
    cumuluspeer "github.com/ubclaunchpad/cumulus/cumuluspeer"
    pstore "github.com/libp2p/go-libp2p-peerstore"
)

func main() {
    log.Info("Starting Cumulus Peer")

    // Get and parse command line arguments
    // targetPeer is a Multiaddr representing the target peer to connect to
    // when joining the Cumulus Network.
    // port is the port to communicate over (defaults to peer.CumulusPort)
    targetPeer := flag.String("t", "", "target peer to connect to")
    port := flag.Int("p", cumuluspeer.CumulusPort, "TCP port to use")
    flag.Parse()

    // Set up a new host on the Cumulus network
    host, ps, err := cumuluspeer.MakeBasicHost(*port)
    if err != nil {
        log.Fatal(err)
    }

    // Set the host StreamHandler for the Cumulus Protocol and use
    // BasicStreamHandler as its StreamHandler.
    host.SetStreamHandler(cumuluspeer.CumulusProtocol,
        cumuluspeer.BasicStreamHandler)

    if *targetPeer == "" {
        // No target was specified, wait for incoming connections
        log.Info("No target provided. Listening for incoming connections...")
        select {} // Hang until someone connects to us
    }

    // Target is specified so connect to it and remember its address
    peerid, targetAddr, err := cumuluspeer.ExtractPeerInfo(*targetPeer)
    if err != nil {
        log.Fatal(err)
    }

    // Store the peer's address in this host's PeerStore
    ps.AddAddr(peerid, targetAddr, pstore.PermanentAddrTTL)

    log.Info("Connected to Cumulus Peer:")
    log.Info("\tPeer ID:", peerid.Pretty())
    log.Info("\tPeer Address:", targetAddr)

    // Open a stream with the peer
    stream, err := host.NewStream(context.Background(), peerid,
        cumuluspeer.CumulusProtocol)
	if err != nil {
		log.Fatal(err)
	}

    // Send a message to the peer
	_, err = stream.Write([]byte("Hello, world!\n"))
	if err != nil {
		log.Error(err)
	}

    // Read the reply from the peer
	reply, err := ioutil.ReadAll(stream)
	if err != nil {
		log.Error(err)
	}

	log.Infof("Read reply: %s", string(reply))
}
