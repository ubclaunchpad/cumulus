package peer

import (
	"errors"
	"fmt"
	"io"
	"net"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/google/uuid"
	"github.com/ubclaunchpad/cumulus/conn"
	"github.com/ubclaunchpad/cumulus/msg"
)

const (
	// DefaultPort is the TCP port hosts will communicate over if none is
	// provided
	DefaultPort = 8000
	// DefaultIP is the IP address new hosts will use if none if provided
	DefaultIP = "127.0.0.1"
	// DefaultRequestTimeout is the default amount of time we will wait for a
	// peer to return a response to a request
	DefaultRequestTimeout = time.Second * 10
	// MessageWaitTime is the amount of time the dispatcher should wait before
	// attempting to read from the connection again when no data was received
	MessageWaitTime = time.Second * 2
	// MaxPeers is the maximum number of peers we can be connected to at a time
	MaxPeers = 50
	// PeerSearchWaitTime is the amount of time the maintainConnections goroutine
	// will wait before checking if we can connect to more peers when is sees that
	// our PeerStore is full.
	PeerSearchWaitTime = time.Second * 10
)

// ResponseHandler is any function that handles a response to a request.
type ResponseHandler func(*msg.Response)

// PushHandler is any function that handles a push message.
type PushHandler func(*msg.Push)

// RequestHandler is any function that returns a response to a request.
type RequestHandler func(*msg.Request) msg.Response

// Peer represents a remote peer we are connected to.
type Peer struct {
	Connection       net.Conn
	Store            *PeerStore
	ListenAddr       string
	requestHandler   RequestHandler
	pushHandler      PushHandler
	responseHandlers map[string]ResponseHandler
	lock             sync.RWMutex
}

// New returns a new Peer
func New(c net.Conn, ps *PeerStore, listenAddr string) *Peer {
	return &Peer{
		Connection:       c,
		Store:            ps,
		ListenAddr:       listenAddr,
		requestHandler:   ps.defaultRequestHandler,
		pushHandler:      ps.defaultPushHandler,
		responseHandlers: make(map[string]ResponseHandler),
	}
}

// Connect attempts to establish a connection with a peer given its listen
// address (in the form <IP address>:<TCP port>). If successful returns the
// peer, otherwise returns error. Once the connection is established the peer
// will be added to the given PeerStore and returned.
func Connect(address string, ps *PeerStore) (*Peer, error) {
	c, err := conn.Dial(address)
	if err != nil {
		return nil, err
	}
	ps.ConnectionHandler(c)
	p := ps.Get(c.RemoteAddr().String())
	if p == nil {
		// This will only be the case if we exchangeListedAddrs fails
		return nil, errors.New("Failed to exchange listen addresses with peer")
	}
	return p, nil
}

// SetRequestHandler will add the given request handler to this peer. The
// request handler must be set for this peer to handle received request messages
// and is NOT set by default.
func (p *Peer) SetRequestHandler(rh RequestHandler) {
	p.requestHandler = rh
}

// SetPushHandler will add the given push handler to this peer. The push handler
// must be set for this peer to handle received push messages and is NOT set by
// default.
func (p *Peer) SetPushHandler(ph PushHandler) {
	p.pushHandler = ph
}

// Dispatch listens on this peer's Connection and passes received messages
// to the appropriate message handlers.
func (p *Peer) Dispatch() {
	// After 3 errors we kill this connection and its associated
	// handlers
	errCount := 0

	for {
		message, err := msg.Read(p.Connection)
		if err != nil {
			if err == io.EOF {
				// This just means the peer hasn't sent anything
				time.Sleep(MessageWaitTime)
			} else {
				if strings.Contains(err.Error(), syscall.ECONNRESET.Error()) || errCount == 3 {
					log.WithError(err).Infof("Disconnecting from peer %s",
						p.ListenAddr)
					p.Store.Remove(p.ListenAddr)
					p.Connection.Close()
					return
				}
				log.WithError(err).Error("Dispatcher failed to read message")
				errCount++
			}
			continue
		}

		switch message.(type) {
		case *msg.Request:
			req := message.(*msg.Request)
			if p.requestHandler == nil {
				log.Errorf("Request received but no request handler set for peer %s",
					p.ListenAddr)
			} else {
				response := p.requestHandler(req)
				response.Write(p.Connection)
			}
		case *msg.Response:
			res := message.(*msg.Response)
			rh := p.getResponseHandler(res.ID)
			if rh == nil {
				log.Error("Dispatcher could not find response handler for response on peer",
					p.ListenAddr)
			} else {
				rh(res)
			}
		case *msg.Push:
			push := message.(*msg.Push)
			if p.pushHandler == nil {
				log.Error("Dispatcher could not find push handler for push message on peer",
					p.ListenAddr)
			} else {
				p.pushHandler(push)
			}
		}
	}
}

// Request sends the given request over this peer's Connection and registers the
// given response handler to be called when the response arrives at the dispatcher.
// implementations. Returns error if request could not be written.
func (p *Peer) Request(req msg.Request, rh ResponseHandler) error {
	responseReceived := make(chan bool)

	// Wrap response handler so we know to stop waiting for a timeout event when
	// the response is received
	wrapper := func(response *msg.Response) {
		responseReceived <- true
		rh(response)
	}

	p.addResponseHandler(req.ID, wrapper)
	err := req.Write(p.Connection)
	if err != nil {
		return err
	}

	// Wait for timeout or response
	go func() {
		select {
		case <-responseReceived:
		case <-time.After(DefaultRequestTimeout):
			// Timeout waiting for response
			message := fmt.Sprintf("Timed out waiting for response from peer %s", p.ListenAddr)
			log.Debug(message)

			err := &msg.ProtocolError{
				Code:    msg.RequestTimeout,
				Message: message,
			}
			timeoutResponse := &msg.Response{
				ID:    req.ID,
				Error: err,
			}

			// Pass the ProtocolError to the actual response handler
			rh(timeoutResponse)
		}

		// Remove response handler when we are done waiting
		p.removeResponseHandler(req.ID)
	}()

	return nil
}

// Push sends a push message to this peer. Returns an error if the push message
// could not be written.
func (p *Peer) Push(push msg.Push) error {
	return push.Write(p.Connection)
}

// MaintainConnections will infinitely attempt to maintain as close to MaxPeers
// connections as possible by requesting PeerInfo from peers in the PeerStore
// and establishing connections with newly discovered peers.
// NOTE: this should be called only once and should be run as a goroutine.
func (ps *PeerStore) MaintainConnections(wg *sync.WaitGroup) {
	// Signal that we MaintainConnections is running
	wg.Done()
	for {
		peerAddrs := ps.Addrs()
		for i := 0; i < len(peerAddrs); i++ {
			if ps.Size() >= MaxPeers {
				// Already connected to enough peers. Don't try connecting to
				// any more for a while.
				break
			}
			peerInfoRequest := msg.Request{
				ID:           uuid.New().String(),
				ResourceType: msg.ResourcePeerInfo,
			}
			p := ps.Get(peerAddrs[i])
			ps.lock.RLock()
			if p != nil {
				// Need to do this check in case the peer got removed
				p.Request(peerInfoRequest, ps.PeerInfoHandler)
			}
			ps.lock.RUnlock()
		}
		// Looks like we hit peer.MaxPeers. Wait for a while before checking how many
		// peers we are connected to. We don't want to spin.
		time.Sleep(PeerSearchWaitTime)
	}
}

func (p *Peer) addResponseHandler(id string, rh ResponseHandler) {
	p.lock.Lock()
	defer p.lock.Unlock()
	p.responseHandlers[id] = rh
}

func (p *Peer) removeResponseHandler(id string) {
	p.lock.Lock()
	defer p.lock.Unlock()
	delete(p.responseHandlers, id)
}

func (p *Peer) getResponseHandler(id string) ResponseHandler {
	p.lock.Lock()
	defer p.lock.Unlock()
	return p.responseHandlers[id]
}

// validAddress checks if the given TCP/IP address is valid
func validAddress(addr string) bool {
	parts := strings.Split(addr, ":")
	if len(parts) != 2 || net.ParseIP(parts[0]) == nil {
		return false
	}
	port, err := strconv.Atoi(parts[1])
	if err != nil {
		return false
	}
	return port > 1024 && port < 65536
}
