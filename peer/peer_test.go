package peer

import (
	"fmt"
	"io"
	"net"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/ubclaunchpad/cumulus/blockchain"
	"github.com/ubclaunchpad/cumulus/conn"
	"github.com/ubclaunchpad/cumulus/msg"
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
)

// fakeConn implements net.Conn
type fakeConn struct {
	Addr net.Addr
}

func (fc fakeConn) Read(b []byte) (n int, err error)   { return 0, nil }
func (fc fakeConn) Write(b []byte) (n int, err error)  { return 0, nil }
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

// This will error if there are concurrent accesses to the PeerStore, or error
// if an atomic operation returns un unexpected result.
func TestConcurrentPeerStore(t *testing.T) {
	ps := NewPeerStore()

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
	ps := NewPeerStore()
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
	ps := NewPeerStore()
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
	p := New(fc, PStore, fa.String())
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

	p2 := New(fc, PStore, fa.String())
	if p2.requestHandler != nil {
		t.FailNow()
	}
}

func TestSetPushHandler(t *testing.T) {
	var fa net.Addr
	var fc net.Conn
	fa = fakeAddr{Addr: "127.0.0.1"}
	fc = fakeConn{Addr: fa}
	p := New(fc, PStore, fa.String())
	if p.pushHandler != nil {
		t.FailNow()
	}

	ph := func(req *msg.Push) {}

	p.SetPushHandler(ph)

	if p.pushHandler == nil {
		t.FailNow()
	}

	p2 := New(fc, PStore, fa.String())
	if p2.pushHandler != nil {
		t.FailNow()
	}
}

func TestSetDefaultRequestHandler(t *testing.T) {
	var fa net.Addr
	var fc net.Conn
	fa = fakeAddr{Addr: "127.0.0.1"}
	fc = fakeConn{Addr: fa}
	p := New(fc, PStore, fa.String())
	if p.requestHandler != nil {
		t.FailNow()
	}

	rh := func(req *msg.Request) msg.Response {
		return msg.Response{
			ID:       "heyyou",
			Resource: "i can see you",
		}
	}

	SetDefaultRequestHandler(rh)

	if p.requestHandler != nil {
		t.FailNow()
	}

	p2 := New(fc, PStore, fa.String())
	if p2.requestHandler == nil {
		t.FailNow()
	}
}

func TestSetDefaultPushHandler(t *testing.T) {
	var fa net.Addr
	var fc net.Conn
	fa = fakeAddr{Addr: "127.0.0.1"}
	fc = fakeConn{Addr: fa}
	p := New(fc, PStore, fa.String())
	if p.pushHandler != nil {
		t.FailNow()
	}

	ph := func(req *msg.Push) {}

	SetDefaultPushHandler(ph)

	if p.pushHandler != nil {
		t.FailNow()
	}

	p2 := New(fc, PStore, fa.String())
	if p2.pushHandler == nil {
		t.FailNow()
	}
}

func TestSendRequestAndReceiveResponse(t *testing.T) {
	ListenAddr = "127.0.0.1:8080"
	sentResponse := make(chan bool)
	receivedValidResponse := make(chan bool)
	var c net.Conn

	go conn.Listen(ListenAddr, ConnectionHandler)
	c, err := conn.Dial(ListenAddr)
	if err != nil {
		t.FailNow()
	}

	_, err = exchangeListenAddrs(c, time.Second*5)
	if err != nil {
		t.FailNow()
	}

	// Wait a little while to give the peer time to get set up before we check
	// the PeerStore.
	select {
	case <-time.After(time.Second * 1):
	}

	p := PStore.Get(c.RemoteAddr().String())
	if p == nil {
		t.FailNow()
	}

	req := msg.Request{
		ID:           uuid.New().String(),
		ResourceType: msg.ResourcePeerInfo,
	}
	responseHandler := func(res *msg.Response) {
		addrs := res.Resource.([]string)
		if len(addrs) != 1 || addrs[0] != "127.0.0.1:8080" {
			receivedValidResponse <- false
		}
		receivedValidResponse <- true
	}
	err = p.Request(req, responseHandler)
	if err != nil {
		t.FailNow()
	}

	// Receive request and send response
	go func() {
		for {
			reqMsg, err := msg.Read(c)
			if err == io.EOF {
				continue
			} else if err != nil {
				sentResponse <- false
				return
			}

			switch reqMsg.(type) {
			case *msg.Request:
				receivedReq := reqMsg.(*msg.Request)
				if receivedReq.ResourceType != msg.ResourcePeerInfo {
					sentResponse <- false
					return
				}
				res := msg.Response{
					ID:       receivedReq.ID,
					Resource: PStore.Addrs(),
				}
				err := res.Write(c)
				if err != nil {
					sentResponse <- false
					return
				}
				sentResponse <- true
			default:
				sentResponse <- false
			}
		}
	}()

	select {
	case sent := <-sentResponse:
		if !sent {
			t.FailNow()
		}
	case <-time.After(time.Second * 5):
		t.FailNow()
	}

	select {
	case passed := <-receivedValidResponse:
		if !passed {
			t.FailNow()
		}
	case <-time.After(time.Second * 5):
		t.FailNow()
	}

	c.Close()
}

func TestReceiveRequestAndSendResponse(t *testing.T) {
	ListenAddr = "127.0.0.1:8080"

	SetDefaultRequestHandler(func(req *msg.Request) msg.Response {
		res := msg.Response{ID: req.ID}

		switch req.ResourceType {
		case msg.ResourcePeerInfo:
			res.Resource = PStore.Addrs()
		default:
			res.Error = msg.NewProtocolError(msg.InvalidResourceType,
				"Invalid resource type")
		}

		return res
	})

	go conn.Listen(ListenAddr, ConnectionHandler)
	c, err := conn.Dial(ListenAddr)
	if err != nil {
		t.FailNow()
	}

	_, err = exchangeListenAddrs(c, time.Second*5)
	if err != nil {
		t.FailNow()
	}

	// Wait a little while to give the peer time to get set up before we check
	// the PeerStore.
	select {
	case <-time.After(time.Second * 1):
	}

	p := PStore.Get(c.RemoteAddr().String())
	if p == nil {
		t.FailNow()
	}

	req := msg.Request{
		ID:           uuid.New().String(),
		ResourceType: msg.ResourcePeerInfo,
	}
	err = req.Write(c)
	if err != nil {
		t.FailNow()
	}

	resChan := make(chan *msg.Response)
	failChan := make(chan bool)

	go func() {
		for {
			resMsg, err := msg.Read(c)
			if err == io.EOF {
				continue
			} else if err != nil {
				failChan <- true
				return
			}
			resChan <- resMsg.(*msg.Response)
			return
		}
	}()

	select {
	case res := <-resChan:
		addrs := res.Resource.([]string)
		if len(addrs) != 1 || addrs[0] != ListenAddr {
			t.FailNow()
		}
	case <-failChan:
		t.FailNow()
	case <-time.After(time.Second * 5):
		t.FailNow()
	}

	c.Close()
}

func TestReceivePush(t *testing.T) {
	ListenAddr = "127.0.0.1:8080"
	receivedValidPush := make(chan bool)
	var c net.Conn

	SetDefaultPushHandler(func(req *msg.Push) {
		if req.ResourceType != msg.ResourceTransaction {
			receivedValidPush <- false
			return
		}
		tr := req.Resource.(*blockchain.Transaction)
		if tr.Len() <= blockchain.SigLen {
			receivedValidPush <- false
			return
		}
		receivedValidPush <- true
	})

	go conn.Listen(ListenAddr, ConnectionHandler)
	c, err := conn.Dial(ListenAddr)
	if err != nil {
		t.FailNow()
	}

	_, err = exchangeListenAddrs(c, time.Second*5)
	if err != nil {
		t.FailNow()
	}

	// Wait a little while to give the peer time to get set up before we check
	// the PeerStore.
	select {
	case <-time.After(time.Second * 1):
	}

	p := PStore.Get(c.RemoteAddr().String())
	if p == nil {
		t.FailNow()
	}

	req := msg.Push{
		ResourceType: msg.ResourceTransaction,
		Resource:     blockchain.NewTestTransaction(),
	}
	err = req.Write(c)
	if err != nil {
		fmt.Println("Failed to write push message ", err)
		t.FailNow()
	}

	select {
	case passed := <-receivedValidPush:
		if !passed {
			t.FailNow()
		}
	case <-time.After(time.Second * 5):
		t.FailNow()
	}

	c.Close()
}

func TestSendPush(t *testing.T) {
	ListenAddr = "127.0.0.1:8080"
	receivedValidPush := make(chan bool)
	var c net.Conn

	go conn.Listen(ListenAddr, ConnectionHandler)
	c, err := conn.Dial(ListenAddr)
	if err != nil {
		t.FailNow()
	}

	_, err = exchangeListenAddrs(c, time.Second*5)
	if err != nil {
		t.FailNow()
	}

	// Wait a little while to give the peer time to get set up before we check
	// the PeerStore.
	select {
	case <-time.After(time.Second * 1):
	}

	p := PStore.Get(c.RemoteAddr().String())
	if p == nil {
		t.FailNow()
	}

	push := msg.Push{
		ResourceType: msg.ResourceBlock,
		Resource:     blockchain.NewTestBlock(),
	}
	err = p.Push(push)
	if err != nil {
		t.FailNow()
	}

	go func() {
		for {
			pushMsg, err := msg.Read(c)
			if err == io.EOF {
				continue
			} else if err != nil {
				receivedValidPush <- false
				return
			}

			push := pushMsg.(*msg.Push)
			block := push.Resource.(*blockchain.Block)
			if push.ResourceType != msg.ResourceBlock || block.Len() <= blockchain.BlockHeaderLen {
				receivedValidPush <- false
				return
			}
			receivedValidPush <- true
			return
		}
	}()

	select {
	case passed := <-receivedValidPush:
		if !passed {
			t.FailNow()
		}
	case <-time.After(time.Second * 5):
		t.FailNow()
	}

	c.Close()
}

func TestMaintinConnections(t *testing.T) {
	ListenAddr = "127.0.0.1:8080"
	listenAddr2 := "127.0.0.1:8081"
	var c1 net.Conn
	receivedConnection := make(chan bool)
	responded := make(chan bool)

	go conn.Listen(ListenAddr, ConnectionHandler)
	c1, err := conn.Dial(ListenAddr)
	if err != nil {
		t.FailNow()
	}

	_, err = exchangeListenAddrs(c1, time.Second*5)
	if err != nil {
		t.FailNow()
	}

	testConnectionHandler := func(c net.Conn) {
		ConnectionHandler(c)
		c.Close()
		receivedConnection <- true
	}

	go conn.Listen(listenAddr2, testConnectionHandler)
	ListenAddr = ""
	go MaintainConnections()
	go func() {
		for {
			reqMsg, err := msg.Read(c1)
			if err == io.EOF {
				continue
			} else if err != nil {
				responded <- false
				return
			}

			req := reqMsg.(*msg.Request)
			if req.ResourceType != msg.ResourcePeerInfo {
				responded <- false
				return
			}
			res := msg.Response{
				ID:       req.ID,
				Resource: append(make([]string, 0), listenAddr2),
			}
			err = res.Write(c1)
			responded <- err == nil
			return
		}
	}()

	select {
	case sentInfo := <-responded:
		if !sentInfo {
			t.FailNow()
		}
	case <-time.After(time.Second * 5):
		t.FailNow()
	}
	select {
	case <-receivedConnection:
	case <-time.After(time.Second * 5):
		t.FailNow()
	}

	c1.Close()
}
