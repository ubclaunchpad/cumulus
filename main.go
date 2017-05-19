package main

import (
	"flag"
	"io/ioutil"

	log "github.com/sirupsen/logrus"
	"github.com/ubclaunchpad/cumulus/peer"
)

func main() {
	log.Info("Starting Cumulus Peer")

	// Get and parse command line arguments
	// targetPeer is a Multiaddr representing the target peer to connect to
	// when joining the Cumulus Network.
	// port is the port to communicate over (defaults to peer.DefaultPort).
	// ip is the public IP address of the this host.
	targetPeer := flag.String("t", "", "target peer to connect to")
	port := flag.Int("p", peer.DefaultPort, "TCP port to use for this host")
	ip := flag.String("i", peer.DefaultIP, "IP address to use for this host")
	debug := flag.Bool("d", false, "Enable debug logging")
	flag.Parse()

	if *debug {
		log.SetLevel(log.DebugLevel)
	}

	// Set up a new host on the Cumulus network
	host, err := peer.New(*ip, *port)
	if err != nil {
		log.Fatal(err)
	}

	// Set the host StreamHandler for the Cumulus Protocol and use
	// BasicStreamHandler as its StreamHandler.
	host.SetStreamHandler(peer.CumulusProtocol, host.Receive)
	if *targetPeer == "" {
		// No target was specified, wait for incoming connections
		log.Info("No target provided. Listening for incoming connections...")
		select {} // Hang until someone connects to us
	}

	stream, err := host.Connect(*targetPeer)
	if err != nil {
		log.Fatal(err)
	}

	// Send a message to the peer
	_, err = stream.Write([]byte("Hello, world!\n"))
	if err != nil {
		log.Fatal(err)
	}

	// Read the reply from the peer
	reply, err := ioutil.ReadAll(stream)
	if err != nil {
		log.Fatal(err)
	}

	log.Debugf("Peer %s read reply: %s", host.ID(), string(reply))

	log.Debug("Hanging...")
	select {}

	host.Close()
}
