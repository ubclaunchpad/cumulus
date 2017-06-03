package stream

import (
	"sync"

	"github.com/libp2p/go-libp2p-net"
	"github.com/ubclaunchpad/cumulus/message"
)

// Stream is a synchronized container for net.Stream and a mapping of listener
// ids to response channels
type Stream struct {
	net.Stream
	listeners map[string]chan *message.Response
	lock      *sync.RWMutex
}

// New returns a new Stream containing the given net.Stream
func New(s net.Stream) *Stream {
	return &Stream{
		s,
		make(map[string]chan *message.Response),
		&sync.RWMutex{},
	}
}

// NewListener synchronously adds a listener to this peer's listeners map
func (s *Stream) NewListener(id string) chan *message.Response {
	s.lock.Lock()
	s.listeners[id] = make(chan *message.Response)
	lchan := s.listeners[id]
	s.lock.Unlock()
	return lchan
}

// RemoveListener synchronously removes a listener from this peer's listeners map
func (s *Stream) RemoveListener(id string) {
	s.lock.Lock()
	delete(s.listeners, id)
	s.lock.Unlock()
}

// Listener synchronously retrieves the channel the listener with the given
// request/response id is waiting on
func (s *Stream) Listener(id string) chan *message.Response {
	s.lock.RLock()
	lchan := s.listeners[id]
	s.lock.RUnlock()
	return lchan
}
