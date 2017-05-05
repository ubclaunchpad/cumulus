package main

import (
    "context"
    "flag"
    "fmt"
    "io/ioutil"

    logger "github.com/ubclaunchpad/cumulus/logging"
    cumuluspeer "github.com/ubclaunchpad/cumulus/cumulus-peer"
    pstore "github.com/libp2p/go-libp2p-peerstore"
)

func main() {
    fmt.Println("Starting Cumulus Peer")

    // Initialize Cumulus logger
    logger.Init()

    // Get and parse command line arguments
    // targetPeer is a Multiaddr representing the target peer to connect to
    // when joining the Cumulus Network.
    targetPeer := flag.String("p", "", "target peer to connect to")
    flag.Parse()

    // Set up a new host on the Cumulus network
    host, err := cumuluspeer.MakeHost()
    if err != nil {
        logger.Log.Error(err)
    }

    // Set the host StreamHandler for the Cumulus Protocol and use
    // BasicStreamHandler as its StreamHandler.
    host.SetStreamHandler(cumuluspeer.CumulusProtocol,
        cumuluspeer.BasicStreamHandler)

    if *targetPeer == "" {
        // No target was specified, wait for incoming connections
        logger.Log.Info("No target provided. Listening for incoming connections...")
        select {} // Hang until someone connects to us
    }

    // Target is specified so connect to it and remember its address
    peerid, targetAddr := cumuluspeer.ExtractPeerInfo(*targetPeer)

    // Store the peer's address in this host's PeerStore
    host.Peerstore().AddAddr(peerid, targetAddr, pstore.PermanentAddrTTL)

    logger.Log.Notice("Connected to Cumulus Peer:")
    logger.Log.Notice("\tPeer ID:", peerid.Pretty())
    logger.Log.Notice("\tPeer Address:", targetAddr)

    // Open a stream with the peer
    stream, err := host.NewStream(context.Background(), peerid,
        cumuluspeer.CumulusProtocol)
	if err != nil {
		logger.Log.Error(err)
	}

    // Send a message to the peer
	_, err = stream.Write([]byte("Hello, world!\n"))
	if err != nil {
		logger.Log.Error(err)
	}

    // Read the reply from the peer
	reply, err := ioutil.ReadAll(stream)
	if err != nil {
		logger.Log.Error(err)
	}

	logger.Log.Info("read reply: %q\n", reply)
}
