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

// PStore stores information about every peer we are connected to.
var PStore = &Peerstore{Peers: make([]*Peer, 0)}

// Peerstore is a thread-safe container for all the peers we are currently
// connected to.
type Peerstore struct {
	Peers []*Peer
	lock  sync.RWMutex
}

// AddPeer synchronously adds the given peer to the peerstore
func (ps *Peerstore) AddPeer(p *Peer) {
	ps.lock.Lock()
	ps.Peers = append(ps.Peers, p)
	ps.lock.Unlock()
}

// RemovePeer synchronously removes the given peer from the peerstore
func (ps *Peerstore) RemovePeer(id string) {
	ps.lock.Lock()
	for i := 0; i < len(ps.Peers); i++ {
		if ps.Peers[i].ID.String() == id {
			ps.Peers = append(ps.Peers[:i], ps.Peers[i+1:]...)
		}
	}
	ps.lock.Unlock()
}

// Peer synchronously retreives the peer with the given id from the peerstore
func (ps *Peerstore) Peer(id string) *Peer {
	var peer *Peer
	ps.lock.RLock()
	for i := 0; i < len(ps.Peers); i++ {
		if ps.Peers[i].ID.String() == id {
			peer = ps.Peers[i]
			break
		}
	}
	ps.lock.RUnlock()
	return peer
}

// Addrs returns the list of addresses of the peers in the peerstore in the form
// <IP addr>:<port>
func (ps *Peerstore) Addrs() []string {
	ps.lock.RLock()
	addrs := make([]string, len(ps.Peers), len(ps.Peers))
	for i := 0; i < len(ps.Peers); i++ {
		addrs[i] = ps.Peers[i].Connection.RemoteAddr().String()
	}
	ps.lock.RUnlock()
	return addrs
}

// Peer represents a remote peer we are connected to
type Peer struct {
	ID         uuid.UUID
	Connection net.Conn
	Peerstore  *Peerstore
	resChans   map[string]chan *message.Response
	reqChan    chan *message.Request
	pushChan   chan *message.Push
	lock       sync.RWMutex
}

// New returns a new Peer
func New(c net.Conn, ps *Peerstore) *Peer {
	return &Peer{
		ID:         uuid.New(),
		Connection: c,
		Peerstore:  ps,
		resChans:   make(map[string]chan *message.Response),
		reqChan:    make(chan *message.Request),
		pushChan:   make(chan *message.Push),
	}
}

// HandleConnection is called when a new connection is opened with us by a
// remote peer. It will create a dispatcher and message handlers to handle
// sending and reveiving messages over the new connection.
func HandleConnection(c net.Conn) {
	p := New(c, PStore)
	PStore.AddPeer(p)

	go p.Dispatch()
	go p.HandlePushes()
	go p.HandleRequests()
}

// Dispatch listens on this peer's Connection and passes received messages
// to the appropriate message handlers.
func (p *Peer) Dispatch() {
	p.Connection.SetDeadline(time.Now().Add(Timeout))

	for {
		msg, err := message.Read(p.Connection)
		if err != nil {
			log.WithError(err).Error("Dispatcher fialed to read message")
			continue
		}

		switch msg.Type() {
		case message.MessageRequest:
			p.reqChan <- msg.(*message.Request)
			break
		case message.MessageResponse:
			id := msg.(*message.Response).ID
			resChan := p.addResponseChan(id)
			if resChan != nil {
				resChan <- msg.(*message.Response)
			} else {
				log.Errorf("Dispatcher could not find channel for response %s",
					msg.(*message.Response).ID)
			}
			p.removeResponseChan(id)
			break
		case message.MessagePush:
			p.pushChan <- msg.(*message.Push)
			break
		default:
			// Invalid messgae type. Ignore
			log.Debug("Dispatcher received message with invalid type")
		}
	}
}

// HandleRequests waits on this peer's request channel for incoming requests
// from the Dispatcher, responding to each request appropriately.
func (p *Peer) HandleRequests() {
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
			res.Resource = p.Peerstore.Addrs()
			break
		case message.ResourceBlock:
		case message.ResourceTransaction:
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

// HandlePushes waits on this peer's request channel for incoming requests
// from the Dispatcher, responding to each request appropriately.
func (p *Peer) HandlePushes() {
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
					PStore.AddPeer(p)

					go p.Dispatch()
					go p.HandlePushes()
					go p.HandleRequests()
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
	select {
	case res := <-c:
		// TODO: do something with the response
		log.Debugf("Received response %s", res.ID)
		break
	case <-time.After(Timeout):
		break
	}

	p.removeResponseChan(req.ID)
}

// Request sends the given request over this peer's Connection and spawns a
// response listener with AwaitResponse. Returns error if request could not be
// written.
func (p *Peer) Request(req message.Request) error {
	resChan := p.addResponseChan(req.ID)
	err := req.Write(p.Connection)
	if err != nil {
		p.removeResponseChan(req.ID)
		return err
	}

	go p.AwaitResponse(req, resChan)
	return nil
}

func (p *Peer) addResponseChan(id string) chan *message.Response {
	resChan := make(chan *message.Response)
	p.lock.Lock()
	p.resChans[id] = resChan
	p.lock.Unlock()
	return resChan
}

func (p *Peer) removeResponseChan(id string) {
	p.lock.Lock()
	delete(p.resChans, id)
	p.lock.Unlock()
}

func (p *Peer) responseChan(id string) chan *message.Response {
	var resChan chan *message.Response
	p.lock.RLock()
	resChan = p.resChans[id]
	p.lock.RUnlock()
	return resChan
}
