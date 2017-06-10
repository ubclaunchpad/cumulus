package cmd

import (
	"fmt"

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

	// Set up listener. This is run as a goroutine because Listen blocks forever
	log.Infof("Starting listener on %s:%d", ip, port)
	go func() {
		err := conn.Listen(fmt.Sprintf("%s:%d", ip, port), peer.ConnectionHandler)
		if err != nil {
			log.WithError(err).Fatalf("Failed to listen on %s:%d", ip, port)
		}
	}()

	// Connect to remote peer if target provided
	if target != "" {
		log.Infof("Dialing target %s", target)
		c, err := conn.Dial(target)
		if err != nil {
			log.WithError(err).Errorf("Failed to dial target %s", target)
			return
		}
		peer.ConnectionHandler(c)
	}

	// Hang forever. All the work from here on is handled in goroutines. We need
	// this to hang to keep them alive.
	select {}
}
