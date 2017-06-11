package peer

import (
	"io"
	"net"
	"sync"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/google/uuid"
	"github.com/ubclaunchpad/cumulus/msg"
)

const (
	// DefaultPort is the TCP port hosts will communicate over if none is
	// provided
	DefaultPort = 8000
	// DefaultIP is the IP address new hosts will use if none if provided
	DefaultIP = "127.0.0.1"
	// Timeout is the time after which reads from a stream will timeout
	Timeout = time.Second * 30
	// messageWaitTime is the amount of time the dispatcher should wait before
	// attempting to read from the connection again when no data was received
	messageWaitTime = time.Second * 5
)

// PStore stores information about every peer we are connected to. All peers we
// connect to should have a reference to this peerstore so they can populate it.
var PStore = &PeerStore{peers: make(map[string]*Peer, 0)}

// PeerStore is a thread-safe container for all the peers we are currently
// connected to.
type PeerStore struct {
	peers map[string]*Peer
	lock  sync.RWMutex
}

// NewPeerStore returns an initialized peerstore.
func NewPeerStore() *PeerStore {
	return &PeerStore{
		peers: make(map[string]*Peer, 0),
		lock:  sync.RWMutex{},
	}
}

// Add synchronously adds the given peer to the peerstore
func (ps *PeerStore) Add(p *Peer) {
	ps.lock.Lock()
	defer ps.lock.Unlock()
	ps.peers[p.ID.String()] = p
}

// Remove synchronously removes the given peer from the peerstore
func (ps *PeerStore) Remove(id string) {
	ps.lock.Lock()
	defer ps.lock.Unlock()
	delete(ps.peers, id)
}

// Get synchronously retreives the peer with the given id from the peerstore
func (ps *PeerStore) Get(id string) *Peer {
	ps.lock.RLock()
	defer ps.lock.RUnlock()
	return ps.peers[id]
}

// Addrs returns the list of addresses of the peers in the peerstore in the form
// <IP addr>:<port>
func (ps *PeerStore) Addrs() []string {
	ps.lock.RLock()
	defer ps.lock.RUnlock()
	addrs := make([]string, len(ps.peers), len(ps.peers))
	for _, p := range ps.peers {
		addrs = append(addrs, p.Connection.RemoteAddr().String())
	}
	return addrs
}

// Peer represents a remote peer we are connected to.
type Peer struct {
	ID               uuid.UUID
	Connection       net.Conn
	Store            *PeerStore
	responseHandlers map[string]ResponseHandler
	reqChan          chan *msg.Request
	pushChan         chan *msg.Push
	killChan         chan bool
	lock             sync.RWMutex
}

// ResponseHandler is any function that handles a response to a request.
type ResponseHandler func(*msg.Response)

// New returns a new Peer
func New(c net.Conn, ps *PeerStore) *Peer {
	return &Peer{
		ID:               uuid.New(),
		Connection:       c,
		Store:            ps,
		responseHandlers: make(map[string]ResponseHandler),
		reqChan:          make(chan *msg.Request),
		pushChan:         make(chan *msg.Push),
		killChan:         make(chan bool),
	}
}

// ConnectionHandler is called when a new connection is opened with us by a
// remote peer. It will create a dispatcher and message handlers to handle
// sending and retrieving messages over the new connection.
func ConnectionHandler(c net.Conn) {
	p := New(c, PStore)
	PStore.Add(p)

	go p.Dispatch()
	go p.PushHandler()
	go p.RequestHandler()

	log.Infof("Connected to %s", p.Connection.RemoteAddr().String())
}

// Dispatch listens on this peer's Connection and passes received messages
// to the appropriate message handlers.
func (p *Peer) Dispatch() {
	// After 3 consecutive errors we kill this connection and its associated
	// handlers using the killChan
	errCount := 0

	for {
		message, err := msg.Read(p.Connection)
		if err != nil {
			if err == io.EOF {
				// This just means the peer hasn't sent anything
				select {
				case <-time.After(messageWaitTime):
				}
			} else {
				log.WithError(err).Error("Dispatcher failed to read message")
				if errCount == 3 {
					p.killChan <- true
					p.Connection.Close()
					return
				}
				errCount++
			}
			continue
		}
		errCount = 0

		switch message.(type) {
		case *msg.Request:
			p.reqChan <- message.(*msg.Request)
		case *msg.Response:
			res := message.(*msg.Response)
			rh := p.getResponseHandler(res.ID)
			if rh == nil {
				log.Error("Dispatcher could not find response handler for response")
			}
			go rh(res)
			p.removeResponseHandler(res.ID)
		case *msg.Push:
			p.pushChan <- message.(*msg.Push)
		default:
			// Invalid messgae type. Ignore
			log.Debug("Dispatcher received message with invalid type")
		}
	}
}

// RequestHandler waits on this peer's request channel for incoming requests
// from the Dispatcher, responding to each request appropriately.
func (p *Peer) RequestHandler() {
	var req *msg.Request
	for {
		select {
		case req = <-p.reqChan:
		case <-p.killChan:
			return
		}

		res := msg.Response{ID: req.ID}

		switch req.ResourceType {
		case msg.ResourcePeerInfo:
			res.Resource = p.Store.Addrs()
		case msg.ResourceBlock, msg.ResourceTransaction:
			res.Error = msg.NewProtocolError(msg.NotImplemented,
				"Block and Transaction requests are not yet implemented on this peer")
		default:
			res.Error = msg.NewProtocolError(msg.InvalidResourceType,
				"Invalid resource type")
		}

		err := res.Write(p.Connection)
		if err != nil {
			log.WithError(err).Error("RequestHandler failed to send response")
		}
	}
}

// PushHandler waits on this peer's request channel for incoming requests
// from the Dispatcher, responding to each request appropriately.
func (p *Peer) PushHandler() {
	var push *msg.Push
	for {
		select {
		case push = <-p.pushChan:
		case <-p.killChan:
			return
		}

		switch push.ResourceType {
		case msg.ResourceBlock:
		case msg.ResourceTransaction:
		default:
			// Invalid resource type. Ignore
		}
	}
}

// Request sends the given request over this peer's Connection and registers the
// given response hadnler to be called when the response arrives at the dispatcher.
// Returns error if request could not be written.
func (p *Peer) Request(req msg.Request, rh ResponseHandler) error {
	p.addResponseHandler(req.ID, rh)
	err := req.Write(p.Connection)
	if err != nil {
		return err
	}
	return nil
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
