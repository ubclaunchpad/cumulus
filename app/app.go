package app

import "github.com/ubclaunchpad/cumulus/conf"

var (
	config *conf.Config
	// TODO peer store once it's merged in
)

// Run sets up and starts a new Cumulus node with the
// given configuration.
func Run(c conf.Config) {
	config = &c
	// Create peer store
	// Open listening port
	// If target specified, try to open connection

	// Ask target for its peers
	// Connect to these peers until we have enough peers
}
