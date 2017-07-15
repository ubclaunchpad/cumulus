package cmd

import (
	"github.com/spf13/cobra"
	"github.com/ubclaunchpad/cumulus/app"
	"github.com/ubclaunchpad/cumulus/conf"
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
		iface, _ := cmd.Flags().GetString("interface")
		target, _ := cmd.Flags().GetString("target")
		verbose, _ := cmd.Flags().GetBool("verbose")
		console, _ := cmd.Flags().GetBool("console")
		config := conf.Config{
			Interface: iface,
			Port:      uint16(port),
			Target:    target,
			Verbose:   verbose,
			Console:   console,
		}
		app.Run(config)

		// Hang main thread. Everything happens in goroutines from here
		select {}
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
	runCmd.Flags().StringP("target", "t", "", "Address of peer to connect to")
	runCmd.Flags().BoolP("verbose", "v", false, "Enable verbose logging")
	runCmd.Flags().BoolP("console", "c", false, "Start Cumulus console")
}
