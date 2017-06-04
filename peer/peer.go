package peer

import (
	"net"
	"sync"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/google/uuid"
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

// Peerstore is a thread-safe container for all the peers we are currently
// connected to.
type Peerstore struct {
	Peers map[string]*Peer
	lock  sync.RWMutex
}

// AddPeer synchronously adds the given peer to the peerstore
func (ps *Peerstore) AddPeer(p *Peer) {
	ps.lock.Lock()
	ps.Peers[p.ID.String()] = p
	ps.lock.Unlock()
}

// RemovePeer synchronously removes the given peer from the peerstore
func (ps *Peerstore) RemovePeer(p *Peer) {
	ps.lock.Lock()
	delete(ps.Peers, p.ID.String())
	ps.lock.Unlock()
}

// Peer synchronously retreives the peer with the given id from the peerstore
func (ps *Peerstore) Peer(id string) *Peer {
	ps.lock.RLock()
	p := ps.Peers[id]
	ps.lock.RUnlock()
	return p
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
				log.Error("Dispatcher could not find channel for response %s",
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
		case message.ResourceBlock:
		case message.ResourceTransaction:
			res.Error = message.NewProtocolError(message.NotImplemented,
				"PeerInfo, Block, and Transaction requests are not yet implemented on this peer")
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
