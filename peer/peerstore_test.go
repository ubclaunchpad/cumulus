package peer

import (
	"encoding/json"
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
		for _, addr := range addrs2 {
			bc := conn.NewBufConn(false, false)
			ps.Add(New(bc, ps, addr))
			p := ps.Get(addr)
			if p.ListenAddr != addr {
				resChan1 <- false
			}
		}
		resChan1 <- true
	}()

	// Asynchronously add and remove peers
	go func() {
		for _, addr := range addrs1 {
			bc := conn.NewBufConn(false, false)
			ps.Add(New(bc, ps, addr))
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
	ps := NewPeerStore("")
	for _, addr := range addrs1 {
		bc := conn.NewBufConn(false, false)
		ps.Add(New(bc, ps, addr))
	}

	for i := ps.Size(); i > 0; i-- {
		ps.RemoveRandom()
		if ps.Size() != i-1 {
			t.FailNow()
		}
	}
}

func TestAddrs(t *testing.T) {
	ps := NewPeerStore("")
	for _, addr := range addrs1 {
		bc := conn.NewBufConn(false, false)
		ps.Add(New(bc, ps, addr))
	}

	addrs := ps.Addrs()
	for _, addr := range addrs {
		if !inList(addr, addrs1) {
			t.FailNow()
		}
	}
}

func TestSetDefaultRequestHandler(t *testing.T) {
	bc := conn.NewBufConn(false, false)
	ps := NewPeerStore("")

	p := New(bc, ps, "127.0.0.1:8000")
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

	p2 := New(bc, ps, "127.0.0.1:8000")
	if p2.requestHandler == nil {
		t.FailNow()
	}
}

func TestSetDefaultPushHandler(t *testing.T) {
	bc := conn.NewBufConn(false, false)
	ps := NewPeerStore("")

	p := New(bc, ps, "127.0.0.1:8000")
	if p.pushHandler != nil {
		t.FailNow()
	}

	ph := func(req *msg.Push) {}

	ps.SetDefaultPushHandler(ph)

	if p.pushHandler != nil {
		t.FailNow()
	}

	p2 := New(bc, ps, "127.0.0.1:8000")
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
		Type:    msg.RequestMessage,
		Payload: requestPayloadBytes,
	}
	requestBytes, _ := json.Marshal(requestMsg)
	readChan := make(chan []byte)
	writeChan := make(chan []byte)

	bc := conn.NewBufConn(false, false)
	bc.OnRead = func() []byte {
		return <-readChan
	}
	bc.OnWrite = func(writeBytes []byte) {
		writeChan <- writeBytes
	}

	ps := NewPeerStore("")
	connectionHandlerDone := make(chan bool)

	go func() {
		ps.ConnectionHandler(bc)
		connectionHandlerDone <- true
	}()

	receivedRequest := false
	receivedResponse := false

	for !receivedRequest || !receivedResponse {
		select {
		case receivedMsg := <-writeChan:
			var message msg.Message
			var request msg.Request
			var response msg.Response

			json.Unmarshal(receivedMsg, &message)

			switch message.Type {
			case msg.RequestMessage:
				err := json.Unmarshal([]byte(message.Payload), &request)
				if err != nil {
					panic(err)
				}
				receivedRequest = true
				readChan <- requestBytes
			case msg.ResponseMessage:
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
					Type:    msg.ResponseMessage,
					Payload: resBytes,
				}
				responseBytes, _ := json.Marshal(responseMsg)
				readChan <- responseBytes
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
