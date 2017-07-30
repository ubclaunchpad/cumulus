package peer

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
	"sync"
	"testing"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/google/uuid"
	"github.com/ubclaunchpad/cumulus/msg"
)

const (
	FailConnect = "Failed to create connection and perform handshake"
	FailGetPeer = "Failed to get peer from peer store after connection was established"
)

var (
	addrs1 = []string{
		"50.74.89.63", "80.74.170.128", "252.206.163.100", "106.237.154.62",
		"23.37.91.218", "141.205.198.174", "243.116.40.121", "219.202.178.157",
		"208.213.114.64", "99.130.197.2", "215.99.177.252", "253.191.123.93",
		"118.98.13.210", "217.41.101.17", "94.39.137.0", "8.26.57.127",
		"2.121.24.48", "45.166.60.59", "69.14.4.201", "73.11.112.209",
		"119.235.160.135", "158.60.36.47", "173.52.51.91", "160.76.117.247",
		"99.3.196.77", "26.37.188.143", "252.52.197.42", "189.10.80.173",
		"32.15.5.182", "73.178.78.95", "166.109.113.195", "39.137.84.170",
		"82.249.125.238", "154.66.246.230", "53.195.164.196", "79.7.11.105",
		"98.209.56.64", "74.25.239.123", "211.0.166.153", "54.32.155.245",
		"49.37.93.82", "141.54.42.188", "202.15.222.2", "39.251.90.236",
		"227.14.254.34", "163.89.37.232", "232.46.225.13", "125.66.77.152",
		"134.182.19.76", "4.220.173.35",
	}

	addrs2 = []string{
		"7.226.78.25", "90.9.211.160", "201.35.42.214", "71.203.173.18",
		"80.33.238.63", "115.122.81.134", "213.234.92.168", "139.190.9.146",
		"161.86.170.251", "95.126.157.42", "160.198.2.231", "146.174.248.226",
		"206.232.35.67", "99.116.200.57", "95.37.225.234", "227.234.125.133",
		"66.40.130.160", "166.32.202.71", "229.203.48.121", "122.41.93.73",
		"19.127.139.118", "242.85.174.83", "121.145.63.93", "125.187.226.190",
		"227.102.96.138", "133.43.209.108", "245.2.228.113", "43.67.186.61",
		"194.70.178.182", "155.98.10.21", "157.150.51.175", "222.1.20.83",
		"19.253.228.59", "195.118.45.237", "159.78.10.205", "206.31.54.66",
		"31.191.153.165", "130.235.208.32", "130.5.207.98", "5.226.180.24",
	}

	peerStore PeerStore
)

// fakeConn implements net.Conn
type fakeConn struct {
	Addr         net.Addr
	Message      **[]byte
	ReadOnce     bool
	BytesWritten chan []byte
	saveOnWrite  bool
	lock         *sync.RWMutex
}

func newFakeConn(addr net.Addr, message **[]byte,
	readOnce bool, saveOnWrite bool) *fakeConn {
	return &fakeConn{
		Addr:         addr,
		Message:      message,
		ReadOnce:     readOnce,
		BytesWritten: make(chan []byte),
		saveOnWrite:  saveOnWrite,
		lock:         &sync.RWMutex{},
	}
}

func (fc fakeConn) Read(b []byte) (n int, err error) {
	fc.lock.RLock()
	defer fc.lock.RUnlock()

	message := **fc.Message

	var i int
	for i = 0; i < len(message) && i < 512; i++ {
		b[i] = (message)[i]
	}
	if fc.ReadOnce {
		**fc.Message = make([]byte, 0)
	}
	return i, nil
}

func (fc fakeConn) Write(b []byte) (n int, err error) {
	if fc.saveOnWrite {
		fc.BytesWritten <- b
	}
	return len(b), nil
}

func (fc fakeConn) Close() error                       { return nil }
func (fc fakeConn) LocalAddr() net.Addr                { return fc.Addr }
func (fc fakeConn) RemoteAddr() net.Addr               { return fc.Addr }
func (fc fakeConn) SetDeadline(t time.Time) error      { return nil }
func (fc fakeConn) SetReadDeadline(t time.Time) error  { return nil }
func (fc fakeConn) SetWriteDeadline(t time.Time) error { return nil }

// fakeAddr implementes net.Addr
type fakeAddr struct {
	Addr string
}

func (fa fakeAddr) Network() string { return fa.Addr }
func (fa fakeAddr) String() string  { return fa.Addr }

func inList(item string, list []string) bool {
	for _, listItem := range list {
		if item == listItem {
			return true
		}
	}
	return false
}

func TestMain(m *testing.M) {
	log.SetLevel(log.DebugLevel)
	fmt.Println("NOTE: Some errors will be logged during tests. Note that these",
		"errors do NOT necessarily mean the tests are failing.")
	os.Exit(m.Run())
}

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
			fa = fakeAddr{Addr: addr}
			fc = fakeConn{Addr: fa}
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
			fa = fakeAddr{Addr: addr}
			fc = fakeConn{Addr: fa}
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
	var fa fakeAddr
	var fc fakeConn
	ps := NewPeerStore("")
	for _, addr := range addrs1 {
		fa = fakeAddr{Addr: addr}
		fc = fakeConn{Addr: fa}
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
	var fa fakeAddr
	var fc fakeConn
	ps := NewPeerStore("")
	for _, addr := range addrs1 {
		fa = fakeAddr{Addr: addr}
		fc = fakeConn{Addr: fa}
		ps.Add(New(fc, ps, addr))
	}

	addrs := ps.Addrs()
	for _, addr := range addrs {
		if !inList(addr, addrs1) {
			t.FailNow()
		}
	}
}

func TestSetRequestHandler(t *testing.T) {
	var fa net.Addr
	var fc net.Conn
	fa = fakeAddr{Addr: "127.0.0.1"}
	fc = fakeConn{Addr: fa}
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

	p.SetRequestHandler(rh)

	if p.requestHandler == nil {
		t.FailNow()
	}

	p2 := New(fc, ps, fa.String())
	if p2.requestHandler != nil {
		t.FailNow()
	}
}

func TestSetPushHandler(t *testing.T) {
	var fa net.Addr
	var fc net.Conn
	fa = fakeAddr{Addr: "127.0.0.1"}
	fc = fakeConn{Addr: fa}
	ps := NewPeerStore("")

	p := New(fc, ps, fa.String())
	if p.pushHandler != nil {
		t.FailNow()
	}

	ph := func(req *msg.Push) {}

	p.SetPushHandler(ph)

	if p.pushHandler == nil {
		t.FailNow()
	}

	p2 := New(fc, ps, fa.String())
	if p2.pushHandler != nil {
		t.FailNow()
	}
}

func TestSetDefaultRequestHandler(t *testing.T) {
	var fa net.Addr
	var fc net.Conn
	fa = fakeAddr{Addr: "127.0.0.1"}
	fc = fakeConn{Addr: fa}
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
	fa = fakeAddr{Addr: "127.0.0.1"}
	fc = fakeConn{Addr: fa}
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

func TestRequestTimeout(t *testing.T) {
	message := make([]byte, 0)
	messagePtr := &message

	fa := fakeAddr{Addr: "127.0.0.1"}
	fc := newFakeConn(fa, &messagePtr, false, false)

	ps := NewPeerStore("")
	p := New(fc, ps, "")
	responseChan := make(chan *msg.Response)

	responseHandler := func(res *msg.Response) {
		responseChan <- res
	}

	req := msg.Request{
		ID:           uuid.New().String(),
		ResourceType: msg.ResourcePeerInfo,
	}

	p.Request(req, responseHandler)
	select {
	case res := <-responseChan:
		if res.Error == nil || res.Error.Code != msg.RequestTimeout {
			t.Fail()
		}
	}

	if p.getResponseHandler(req.ID) != nil {
		t.Fail()
	}
}

func TestValidAddress(t *testing.T) {
	if !validAddress("124.53.12.53:8080") {
		t.Fail()
	} else if validAddress("132.76.211.0:333") {
		t.Fail()
	} else if validAddress("400.12.43.1:8080") {
		t.Fail()
	} else if validAddress("222.12.43.1:12") {
		t.Fail()
	} else if validAddress("222.12.43.1") {
		t.Fail()
	} else if validAddress("lksdjfliosrut8r") {
		t.Fail()
	} else if validAddress("") {
		t.Fail()
	}
}

func TestResponse(t *testing.T) {
	responseChan := make(chan *msg.Response)

	req := msg.Request{
		ID:           uuid.New().String(),
		ResourceType: msg.ResourcePeerInfo,
	}

	response := msg.Response{
		ID:       req.ID,
		Resource: "HI!",
	}

	responsePayloadBytes, _ := json.Marshal(response)
	responseMsg := msg.Message{
		Type:    "Response",
		Payload: responsePayloadBytes,
	}
	responseBytes, err := json.Marshal(responseMsg)
	if err != nil {
		panic(err)
	}
	rbp := &responseBytes

	var fa net.Addr
	var fc net.Conn

	fa = fakeAddr{Addr: "127.0.0.1"}
	fc = newFakeConn(fa, &rbp, false, false)

	ps := NewPeerStore("")
	p := New(fc, ps, "")

	responseHandler := func(res *msg.Response) {
		responseChan <- res
	}

	go p.Dispatch()

	err = p.Request(req, responseHandler)
	if err != nil {
		panic(err)
	}

	select {
	case res := <-responseChan:
		if res.ID != req.ID {
			t.FailNow()
		} else if res.Resource.(string) != "HI!" {
			t.FailNow()
		}
	}
}

func TestPush(t *testing.T) {
	pushChan := make(chan *msg.Push)

	push := msg.Push{
		ResourceType: msg.ResourceTransaction,
		Resource:     "HEY!",
	}

	pushPayloadBytes, _ := json.Marshal(push)

	pushMsg := msg.Message{
		Type:    "Push",
		Payload: pushPayloadBytes,
	}

	pushBytes, _ := json.Marshal(pushMsg)
	pbp := &pushBytes

	var fa net.Addr
	var fc net.Conn

	fa = fakeAddr{Addr: "127.0.0.1"}
	fc = newFakeConn(fa, &pbp, false, false)

	ps := NewPeerStore("")
	p := New(fc, ps, "")

	p.SetPushHandler(func(push *msg.Push) {
		pushChan <- push
	})

	go p.Dispatch()

	select {
	case push := <-pushChan:
		if push.ResourceType != msg.ResourceTransaction ||
			push.Resource.(string) != "HEY!" {
			t.FailNow()
		}
	}
}

func TestRequest(t *testing.T) {
	requestChan := make(chan *msg.Request)

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
	var fc net.Conn

	fa = fakeAddr{Addr: "127.0.0.1"}
	fc = newFakeConn(fa, &rbp, false, false)

	ps := NewPeerStore("")
	p := New(fc, ps, "")

	p.SetRequestHandler(func(req *msg.Request) msg.Response {
		requestChan <- req
		return msg.Response{}
	})

	go p.Dispatch()

	select {
	case request := <-requestChan:
		if request.ResourceType != msg.ResourcePeerInfo ||
			request.ID != req.ID {
			t.FailNow()
		}
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

	fa = fakeAddr{Addr: "127.0.0.1"}
	fc := newFakeConn(fa, &rbp, true, true)

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
				fc.lock.Lock()
				**fc.Message = requestBytes
				fc.lock.Unlock()
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
				fc.lock.Lock()
				**fc.Message = responseMsgBytes
				fc.lock.Unlock()
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
