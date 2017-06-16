package app

import (
	"fmt"

	log "github.com/Sirupsen/logrus"

	"github.com/ubclaunchpad/cumulus/conf"
	"github.com/ubclaunchpad/cumulus/conn"
	"github.com/ubclaunchpad/cumulus/peer"
)

var (
	config *conf.Config
	// TODO peer store once it's merged in
)

// Run sets up and starts a new Cumulus node with the
// given configuration.
func Run(cfg conf.Config) {
	log.Info("Starting Cumulus node")
	config = &cfg
	// Create peer store

	log.Infof("Starting listener on %s:%d", cfg.Interface, cfg.Port)
	go func() {
		err := conn.Listen(fmt.Sprintf("%s:%d", cfg.Interface, cfg.Port), peer.ConnectionHandler)
		if err != nil {
			log.WithError(err).Fatalf("Failed to listen on %s:%d", cfg.Interface, cfg.Port)
		}
	}()

	if len(cfg.Target) > 0 {
		log.Infof("Dialing target %s", cfg.Target)
		c, err := conn.Dial(cfg.Target)
		if err != nil {
			log.WithError(err).Fatalf("Failed to connect to target")
		}
		peer.ConnectionHandler(c)
	}

	select {}
	// Ask target for its peers
	// Connect to these peers until we have enough peers
	// Download the blockchain
}
