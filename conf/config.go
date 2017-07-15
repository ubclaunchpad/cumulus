package conf

// Config contains all configuration options for a node.
type Config struct {
	// The interface to listen on for new connections.
	Interface string
	// The port to listen on for new connections.
	Port uint16
	// The address of the ingress node we should use to connect
	// to the network.
	Target string
	// Whether or not to enable verbose logging.
	Verbose bool
	// Whether or not to participate in mining new blocks.
	Mine bool
	// Whether or not to start the Cumulus console
	Console bool
}
