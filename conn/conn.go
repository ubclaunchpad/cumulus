package conn

import "net"

// Dial opens a connection to a remote host. `host` should be a string
// in the format <IP>|<hostname>:<port>
func Dial(host string) (net.Conn, error) {
	return net.Dial("tcp", host)
}

// Listen binds to a TCP port and waits for incoming connections.
// When a connection is accepted, dispatches to the handler.
func Listen(iface string, handler func(net.Conn)) error {
	listener, err := net.Listen("tcp", iface)
	if err != nil {
		return err
	}
	for {
		c, err := listener.Accept()
		if err != nil {
			return err
		}
		go handler(c)
	}
}
