package cmd

import (
	"io/ioutil"

	log "github.com/Sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/ubclaunchpad/cumulus/peer"
)

// runCmd represents the run command
var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Run creates and runs a node on the Cumulus network",
	Long: `Run creates a new Cumulus node and connects to the specified target node.
	If a target is not provided, listen for incoming connections.`,
	Run: func(cmd *cobra.Command, args []string) {
		port, _ := cmd.Flags().GetInt("port")
		ip, _ := cmd.Flags().GetString("interface")
		target, _ := cmd.Flags().GetString("target")
		verbose, _ := cmd.Flags().GetBool("verbose")
		run(port, ip, target, verbose)
	},
}

func init() {
	RootCmd.AddCommand(runCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// runCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	runCmd.Flags().IntP("port", "p", peer.DefaultPort, "Port to bind to")
	runCmd.Flags().StringP("interface", "i", peer.DefaultIP, "IP address to listen on")
	runCmd.Flags().StringP("target", "t", "", "Multiaddress of peer to connect to")
	runCmd.Flags().BoolP("verbose", "v", false, "Enable verbose logging")
}

func run(port int, ip, target string, verbose bool) {
	log.Info("Starting Cumulus Peer")

	if verbose {
		log.SetLevel(log.DebugLevel)
	}

	// Set up a new host on the Cumulus network
	host, err := peer.New(ip, port)
	if err != nil {
		log.Fatal(err)
	}

	// Set the host StreamHandler for the Cumulus Protocol and use
	// BasicStreamHandler as its StreamHandler.
	host.SetStreamHandler(peer.CumulusProtocol, host.Receive)
	if target == "" {
		// No target was specified, wait for incoming connections
		log.Info("No target provided. Listening for incoming connections...")
		select {} // Hang until someone connects to us
	}

	stream, err := host.Connect(target)
	if err != nil {
		log.WithError(err).Fatal("Error connecting to target: ", target)
	}

	// Send a message to the peer
	_, err = stream.Write([]byte("Hello, world!"))
	if err != nil {
		log.WithError(err).Fatal("Error sending a message to the peer")
	}

	// Read the reply from the peer
	reply, err := ioutil.ReadAll(stream)
	if err != nil {
		log.WithError(err).Fatal("Error reading a message from the peer")
	}

	log.Debugf("Peer %s read reply: %s", host.ID(), string(reply))

	host.Close()
}
