package peer

import (
	"encoding/json"
	"net"
	"testing"

	"github.com/google/uuid"
	"github.com/ubclaunchpad/cumulus/conn"
	"github.com/ubclaunchpad/cumulus/msg"
)

// This will error if there are concurrent accesses to the PeerStore, or error
// if an atomic operation returns un unexpected result.
func TestConcurrentPeerStore(t *testing.T) {
	ps := NewPeerStore("")

	resChan1 := make(chan bool)
	resChan2 := make(chan bool)

	// Asynchronously add and find peers
	go func() {
		var fa net.Addr
		var fc net.Conn
		for _, addr := range addrs2 {
			fa = conn.TestAddr{Addr: addr}
			fc = conn.TestConn{Addr: fa}
			ps.Add(New(fc, ps, addr))
			p := ps.Get(addr)
			if p.ListenAddr != addr {
				resChan1 <- false
			}
		}
		resChan1 <- true
	}()

	// Asynchronously add and remove peers
	go func() {
		var fa net.Addr
		var fc net.Conn
		for _, addr := range addrs1 {
			fa = conn.TestAddr{Addr: addr}
			fc = conn.TestConn{Addr: fa}
			ps.Add(New(fc, ps, addr))
			ps.Remove(addr)
		}
		resChan2 <- true
	}()

	returnCount := 0
	for returnCount != 2 {
		select {
		case res1 := <-resChan1:
			if !res1 {
				t.FailNow()
			}
			returnCount++
		case res2 := <-resChan2:
			if !res2 {
				t.FailNow()
			}
			returnCount++
		}
	}

	if ps.Size() != len(addrs2) {
		t.FailNow()
	}

	for i := 0; i < len(addrs2); i++ {
		p := ps.Get(addrs2[i])
		if p == nil {
			t.FailNow()
		}
		ps.Remove(addrs2[i])
	}

	if ps.Size() != 0 {
		t.FailNow()
	}
}

func TestRemoveRandom(t *testing.T) {
	var fa conn.TestAddr
	var fc conn.TestConn
	ps := NewPeerStore("")
	for _, addr := range addrs1 {
		fa = conn.TestAddr{Addr: addr}
		fc = conn.TestConn{Addr: fa}
		ps.Add(New(fc, ps, addr))
	}

	for i := ps.Size(); i > 0; i-- {
		ps.RemoveRandom()
		if ps.Size() != i-1 {
			t.FailNow()
		}
	}
}

func TestAddrs(t *testing.T) {
	var fa conn.TestAddr
	var fc conn.TestConn
	ps := NewPeerStore("")
	for _, addr := range addrs1 {
		fa = conn.TestAddr{Addr: addr}
		fc = conn.TestConn{Addr: fa}
		ps.Add(New(fc, ps, addr))
	}

	addrs := ps.Addrs()
	for _, addr := range addrs {
		if !inList(addr, addrs1) {
			t.FailNow()
		}
	}
}

func TestSetDefaultRequestHandler(t *testing.T) {
	var fa net.Addr
	var fc net.Conn
	fa = conn.TestAddr{Addr: "127.0.0.1"}
	fc = conn.TestConn{Addr: fa}
	ps := NewPeerStore("")

	p := New(fc, ps, fa.String())
	if p.requestHandler != nil {
		t.FailNow()
	}

	rh := func(req *msg.Request) msg.Response {
		return msg.Response{
			ID:       "heyyou",
			Resource: "i can see you",
		}
	}

	ps.SetDefaultRequestHandler(rh)

	if p.requestHandler != nil {
		t.FailNow()
	}

	p2 := New(fc, ps, fa.String())
	if p2.requestHandler == nil {
		t.FailNow()
	}
}

func TestSetDefaultPushHandler(t *testing.T) {
	var fa net.Addr
	var fc net.Conn
	fa = conn.TestAddr{Addr: "127.0.0.1"}
	fc = conn.TestConn{Addr: fa}
	ps := NewPeerStore("")

	p := New(fc, ps, fa.String())
	if p.pushHandler != nil {
		t.FailNow()
	}

	ph := func(req *msg.Push) {}

	ps.SetDefaultPushHandler(ph)

	if p.pushHandler != nil {
		t.FailNow()
	}

	p2 := New(fc, ps, fa.String())
	if p2.pushHandler == nil {
		t.FailNow()
	}
}

func TestConnectionHandler(t *testing.T) {
	req := msg.Request{
		ID:           uuid.New().String(),
		ResourceType: msg.ResourcePeerInfo,
	}
	requestPayloadBytes, _ := json.Marshal(req)
	requestMsg := msg.Message{
		Type:    "Request",
		Payload: requestPayloadBytes,
	}
	requestBytes, _ := json.Marshal(requestMsg)
	rbp := &requestBytes

	var fa net.Addr

	fa = conn.TestAddr{Addr: "127.0.0.1"}
	fc := conn.NewTestConn(fa, &rbp, true, true)

	ps := NewPeerStore("")
	connectionHandlerDone := make(chan bool)

	go func() {
		ps.ConnectionHandler(fc)
		connectionHandlerDone <- true
	}()

	receivedRequest := false
	receivedResponse := false

	for !receivedRequest || !receivedResponse {
		select {
		case receivedMsg := <-fc.BytesWritten:
			var message msg.Message
			var request msg.Request
			var response msg.Response

			json.Unmarshal(receivedMsg, &message)

			switch message.Type {
			case "Request":
				err := json.Unmarshal([]byte(message.Payload), &request)
				if err != nil {
					panic(err)
				}
				receivedRequest = true
				fc.Lock.Lock()
				**fc.Message = requestBytes
				fc.Lock.Unlock()
			case "Response":
				err := json.Unmarshal([]byte(message.Payload), &response)
				if err != nil {
					panic(err)
				}
				receivedResponse = true

				res := msg.Response{
					ID:       request.ID,
					Resource: "127.0.0.1:8000",
				}
				resBytes, _ := json.Marshal(res)
				responseMsg := msg.Message{
					Type:    "Response",
					Payload: resBytes,
				}
				responseMsgBytes, _ := json.Marshal(responseMsg)
				fc.Lock.Lock()
				**fc.Message = responseMsgBytes
				fc.Lock.Unlock()
			}
		}
	}

	select {
	case <-connectionHandlerDone:
		if ps.Get("127.0.0.1:8000") == nil {
			t.Fail()
		}
	}
}
