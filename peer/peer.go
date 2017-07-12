package peer

import (
	"encoding/json"
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
	// MessageWaitTime is the amount of time the dispatcher should wait before
	// attempting to read from the connection again when no data was received
	MessageWaitTime = time.Second * 5
	// MaxPeers is the maximum number of peers we can be connected to at a time
	MaxPeers = 50
	// PeerSearchWaitTime is the amount of time the maintainConnections goroutine
	// will wait before checking if we can connect to more peers when is sees that
	// our PeerStore is full.
	PeerSearchWaitTime = time.Second * 10
)

// PeerStore is a thread-safe container for all the peers we are currently
// connected to. It maps remote peer listen addresses to Peer objects.
type PeerStore struct {
	peers                 map[string]*Peer
	ListenAddr            string
	defaultRequestHandler RequestHandler
	defaultPushHandler    PushHandler
	lock                  *sync.RWMutex
}

// NewPeerStore returns an initialized peerstore.
func NewPeerStore(la string) *PeerStore {
	return &PeerStore{
		peers:      make(map[string]*Peer, 0),
		ListenAddr: la,
		lock:       &sync.RWMutex{},
	}
}

// Connect attempts to establish a connection with a peer given its listen
// address (in the form <IP address>:<TCP port>). If successful returns the
// peer, otherwise returns error.
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

// ConnectionHandler is called when a new connection is opened with us by a
// remote peer. It will create a dispatcher and message handlers to handle
// retrieving messages over the new connection and sending them to App.
func (ps *PeerStore) ConnectionHandler(c net.Conn) {
	// Before we can continue we must exchange listen addresses
	addr, err := exchangeListenAddrs(c, PeerSearchWaitTime, ps.ListenAddr)
	if err != nil {
		log.WithError(err).Error("Failed to retrieve peer listen address")
		return
	}
	p := New(c, ps, addr)

	// If we are already at MaxPeers, disconnect from a peer to connect to a new
	// one. This way nobody gets choked out of the network because everybody
	// happens to be fully connected.
	if p.Store.Size() >= MaxPeers {
		p.Store.RemoveRandom()
	}
	p.Store.Add(p)

	go p.Dispatch()
	log.Infof("Connected to %s", p.ListenAddr)
}

// Add synchronously adds the given peer to the peerstore
func (ps *PeerStore) Add(p *Peer) {
	ps.lock.Lock()
	defer ps.lock.Unlock()
	ps.peers[p.ListenAddr] = p
}

// Remove synchronously removes the given peer from the peerstore
func (ps *PeerStore) Remove(addr string) {
	ps.lock.Lock()
	defer ps.lock.Unlock()
	delete(ps.peers, addr)
}

// RemoveRandom removes a random peer from the given PeerStore
func (ps *PeerStore) RemoveRandom() {
	ps.lock.Lock()
	defer ps.lock.Unlock()
	for _, p := range ps.peers {
		delete(ps.peers, p.ListenAddr)
		break
	}
}

// Get synchronously retreives the peer with the given id from the peerstore
func (ps *PeerStore) Get(addr string) *Peer {
	ps.lock.RLock()
	defer ps.lock.RUnlock()
	return ps.peers[addr]
}

// SetDefaultRequestHandler will ensure that all new peers created who's
// RequestHandlers have not been set will use the given request handler by default
// until it is overridden by the call to SetRequestHandler().
func (ps *PeerStore) SetDefaultRequestHandler(rh RequestHandler) {
	ps.defaultRequestHandler = rh
}

// SetDefaultPushHandler will ensure that all new peers created who's
// PushHandlers have not been set will use the given request handler by default
// until it is overridden by the call to SetPushHandler().
func (ps *PeerStore) SetDefaultPushHandler(ph PushHandler) {
	ps.defaultPushHandler = ph
}

// Broadcast sends the given push message to all peers in the PeerStore at the
// time this function is called. Note that if we fail to write the push message
// to a peer the failure is ignored. Generally this is okay, because push
// messages sent via Broadcast() should be propagated by other peers.
func (ps *PeerStore) Broadcast(push msg.Push) {
	ps.lock.RLock()
	defer ps.lock.RUnlock()
	for _, p := range ps.peers {
		p.Push(push)
	}
}

// PeerInfoHandler will handle the response to a PeerInfo request by attempting
// to establish connections with all new peers in the given response Resource.
func (ps *PeerStore) PeerInfoHandler(res *msg.Response) {
	peers := res.Resource.([]string)
	strPeers, _ := json.Marshal(peers)
	log.Debugf("Found peers %s", string(strPeers))
	for i := 0; i < len(peers) && ps.Size() < MaxPeers; i++ {
		if ps.Get(peers[i]) != nil || peers[i] == ps.ListenAddr {
			// We are already connected to this peer. Skip it.
			continue
		}

		p, err := Connect(peers[i], ps)
		if err != nil {
			log.WithError(err).Errorf("Failed to dial peer %s", peers[i])
		}
		log.Infof("Connected to %s", p.ListenAddr)
	}
}

// Addrs returns the list of addresses of the peers in the peerstore in the form
// <IP addr>:<port>
func (ps *PeerStore) Addrs() []string {
	ps.lock.RLock()
	defer ps.lock.RUnlock()
	addrs := make([]string, 0)
	for addr := range ps.peers {
		addrs = append(addrs, addr)
	}
	return addrs
}

// Size synchornously returns the number of peers in the PeerStore
func (ps *PeerStore) Size() int {
	ps.lock.RLock()
	defer ps.lock.RUnlock()
	return len(ps.peers)
}

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

// ResponseHandler is any function that handles a response to a request.
type ResponseHandler func(*msg.Response)

// PushHandler is any function that handles a push message.
type PushHandler func(*msg.Push)

// RequestHandler is any function that returns a response to a request.
type RequestHandler func(*msg.Request) msg.Response

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
	// After 3 consecutive errors we kill this connection and its associated
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
						p.Connection.RemoteAddr().String())
					p.Store.Remove(p.ListenAddr)
					p.Connection.Close()
					return
				}
				log.WithError(err).Error("Dispatcher failed to read message")
				errCount++
			}
			continue
		}
		errCount = 0

		switch message.(type) {
		case *msg.Request:
			if p.requestHandler == nil {
				log.Errorf("Request received but no request handler set for peer %s",
					p.Connection.RemoteAddr().String())
			} else {
				response := p.requestHandler(message.(*msg.Request))
				response.Write(p.Connection)
			}
		case *msg.Response:
			res := message.(*msg.Response)
			rh := p.getResponseHandler(res.ID)
			if rh == nil {
				log.Error("Dispatcher could not find response handler for response")
			} else {
				rh(res)
				p.removeResponseHandler(res.ID)
			}
		case *msg.Push:
			if p.pushHandler == nil {
				log.Errorf("Push message received but no push handler set for peer %s",
					p.Connection.RemoteAddr().String())
			} else {
				p.pushHandler(message.(*msg.Push))
			}
		default:
			// Invalid messgae type. Ignore
			log.Debug("Dispatcher received message with invalid type")
		}
	}
}

// Request sends the given request over this peer's Connection and registers the
// given response hadnler to be called when the response arrives at the dispatcher.
// Returns error if request could not be written.
func (p *Peer) Request(req msg.Request, rh ResponseHandler) error {
	p.addResponseHandler(req.ID, rh)
	err := req.Write(p.Connection)
	return err
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
func (ps *PeerStore) MaintainConnections() {
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
			if p != nil && peerAddrs[i] != ps.ListenAddr {
				// Need to do this check in case the peer got removed
				p.Request(peerInfoRequest, ps.PeerInfoHandler)
			}
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

func exchangeListenAddrs(c net.Conn, d time.Duration, listenAddr string) (string, error) {
	addrChan := make(chan string)
	errChan := make(chan error)

	req := msg.Request{
		ID:           uuid.New().String(),
		ResourceType: msg.ResourcePeerInfo,
	}
	err := req.Write(c)
	if err != nil {
		return "", err
	}

	// Wait for peer to request our listen address and send us its listen address.
	go func() {
		receivedAddr := false
		sentAddr := false
		var addr string

		for !receivedAddr || !sentAddr {
			message, err := msg.Read(c)
			if err == io.EOF {
				continue
			} else if err != nil {
				errChan <- err
			}

			switch message.(type) {
			case *msg.Response:
				// We got the listen address back
				addr = message.(*msg.Response).Resource.(string)
				if validAddress(addr) || addr != listenAddr {
					receivedAddr = true
				}
			case *msg.Request:
				if message.(*msg.Request).ResourceType != msg.ResourcePeerInfo {
					continue
				}
				// We got a listen address request.
				// Send the remote peer our listen address
				res := msg.Response{
					ID:       uuid.New().String(),
					Resource: listenAddr,
				}
				err = res.Write(c)
				if err != nil {
					errChan <- err
				}
				sentAddr = true
			default:
			}
		}

		addrChan <- addr
	}()

	select {
	case addr := <-addrChan:
		return addr, nil
	case err := <-errChan:
		return "", err
	case <-time.After(d):
		return "", fmt.Errorf("Failed to exchange listen addresses with %s", c.RemoteAddr().String())
	}
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
