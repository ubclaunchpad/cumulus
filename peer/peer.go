package peer

import (
	"net"
	"sync"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/google/uuid"
	"github.com/ubclaunchpad/cumulus/conn"
	"github.com/ubclaunchpad/cumulus/message"
)

const (
	// DefaultPort is the TCP port hosts will communicate over if none is
	// provided
	DefaultPort = 8000
	// DefaultIP is the IP address new hosts will use if none if provided
	DefaultIP = "127.0.0.1"
	// Timeout is the time after which reads from a stream will timeout
	Timeout = time.Second * 30
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

// ChanStore is a threadsafe container for response channels.
type ChanStore struct {
	chans map[string]chan *message.Response
	lock  sync.RWMutex
}

// Add synchronously adds a channel with the given id to the store.
func (cs *ChanStore) Add(id string, channel chan *message.Response) {
	cs.lock.Lock()
	defer cs.lock.Lock()
	cs.chans[id] = channel
}

// Remove synchronously removes the channel with the given ID.
func (cs *ChanStore) Remove(id string) {
	cs.lock.Lock()
	defer cs.lock.Unlock()
	delete(cs.chans, id)
}

// Get retrieves the channel with the given ID.
func (cs *ChanStore) Get(id string) chan *message.Response {
	cs.lock.RLock()
	defer cs.lock.RUnlock()
	return cs.chans[id]
}

// Peer represents a remote peer we are connected to.
type Peer struct {
	ID         uuid.UUID
	Connection net.Conn
	Store      *PeerStore
	resChans   *ChanStore
	reqChan    chan *message.Request
	pushChan   chan *message.Push
	lock       sync.RWMutex
}

// New returns a new Peer
func New(c net.Conn, ps *PeerStore) *Peer {
	cs := &ChanStore{
		chans: make(map[string]chan *message.Response),
		lock:  sync.RWMutex{},
	}
	return &Peer{
		ID:         uuid.New(),
		Connection: c,
		Store:      ps,
		resChans:   cs,
		reqChan:    make(chan *message.Request),
		pushChan:   make(chan *message.Push),
	}
}

// ConnectionHandler is called when a new connection is opened with us by a
// remote peer. It will create a dispatcher and message handlers to handle
// sending and reveiving messages over the new connection.
func ConnectionHandler(c net.Conn) {
	p := New(c, PStore)
	PStore.Add(p)

	go p.Dispatch()
	go p.PushHandler()
	go p.RequestHandler()
}

// Dispatch listens on this peer's Connection and passes received messages
// to the appropriate message handlers.
func (p *Peer) Dispatch() {
	p.Connection.SetDeadline(time.Now().Add(Timeout))

	for {
		msg, err := message.Read(p.Connection)
		if err != nil {
			log.WithError(err).Error("Dispatcher failed to read message")
			continue
		}

		switch msg.Type() {
		case message.MessageRequest:
			p.reqChan <- msg.(*message.Request)
		case message.MessageResponse:
			res := msg.(*message.Response)
			resChan := p.resChans.Get(res.ID)
			if resChan != nil {
				resChan <- msg.(*message.Response)
			} else {
				log.Errorf("Dispatcher could not find channel for response %s", res.ID)
			}
			p.resChans.Remove(res.ID)
		case message.MessagePush:
			p.pushChan <- msg.(*message.Push)
		default:
			// Invalid messgae type. Ignore
			log.Debug("Dispatcher received message with invalid type")
		}
	}
}

// RequestHandler waits on this peer's request channel for incoming requests
// from the Dispatcher, responding to each request appropriately.
func (p *Peer) RequestHandler() {
	var req *message.Request
	for {
		select {
		case req = <-p.reqChan:
			break
		case <-time.After(Timeout):
			continue
		}

		res := message.Response{ID: req.ID}

		switch req.ResourceType {
		case message.ResourcePeerInfo:
			res.Resource = p.Store.Addrs()
			break
		case message.ResourceBlock, message.ResourceTransaction:
			res.Error = message.NewProtocolError(message.NotImplemented,
				"Block and Transaction requests are not yet implemented on this peer")
		default:
			res.Error = message.NewProtocolError(message.InvalidResourceType,
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
	var push *message.Push
	for {
		select {
		case push = <-p.pushChan:
			break
		case <-time.After(Timeout):
			continue
		}

		switch push.ResourceType {
		case message.ResourcePeerInfo:
			for _, addr := range push.Resource.([]string) {
				conn.Listen(addr, func(c net.Conn) {
					p := New(c, PStore)
					p.Store.Add(p)

					go p.Dispatch()
					go p.PushHandler()
					go p.RequestHandler()
				})
			}
			break
		case message.ResourceBlock:
		case message.ResourceTransaction:
		default:
			// Invalid resource type. Ignore
		}
	}
}

// AwaitResponse waits on a response channel for a response message sent by the
// Dispatcher. When a response arrives it is handled appropriately.
func (p *Peer) AwaitResponse(req message.Request, c chan *message.Response) {
	defer p.resChans.Remove(req.ID)
	select {
	case res := <-c:
		// TODO: do something with the response
		log.Debugf("Received response %s", res.ID)
	case <-time.After(Timeout):
		break
	}
}

// Request sends the given request over this peer's Connection and spawns a
// response listener with AwaitResponse. Returns error if request could not be
// written.
func (p *Peer) Request(req message.Request) error {
	resChan := make(chan *message.Response)
	p.resChans.Add(req.ID, resChan)
	err := req.Write(p.Connection)
	if err != nil {
		p.resChans.Remove(req.ID)
		return err
	}

	go p.AwaitResponse(req, resChan)
	return nil
}
