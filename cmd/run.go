package cmd

import (
	"fmt"
	"net"

	log "github.com/Sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/ubclaunchpad/cumulus/conn"
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

var ps = peer.NewPeerStore()

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

	err := conn.Listen(fmt.Sprintf("%s:%d", ip, port), handleConnection)
	if err != nil {
		log.WithError(err).Errorf("Failed to listen on %s:%d", ip, port)
	}
}

func handleConnection(c net.Conn) {
	p := peer.New(c, ps)
	ps.Add(p)

	go p.Dispatch()
	go p.PushHandler()
	go p.RequestHandler()
}
