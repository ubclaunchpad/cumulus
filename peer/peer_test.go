package peer

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"testing"

	log "github.com/Sirupsen/logrus"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/ubclaunchpad/cumulus/blockchain"
	"github.com/ubclaunchpad/cumulus/conn"
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
)

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

func TestSetRequestHandler(t *testing.T) {
	fa := conn.TestAddr{Addr: "127.0.0.1"}
	bc := conn.NewBufConn(false, false)
	ps := NewPeerStore("")

	p := New(bc, ps, fa.String())
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

	p2 := New(bc, ps, fa.String())
	if p2.requestHandler != nil {
		t.FailNow()
	}
}

func TestSetPushHandler(t *testing.T) {
	fa := conn.TestAddr{Addr: "127.0.0.1"}
	bc := conn.NewBufConn(false, false)
	ps := NewPeerStore("")

	p := New(bc, ps, fa.String())
	if p.pushHandler != nil {
		t.FailNow()
	}

	ph := func(req *msg.Push) {}

	p.SetPushHandler(ph)

	if p.pushHandler == nil {
		t.FailNow()
	}

	p2 := New(bc, ps, fa.String())
	if p2.pushHandler != nil {
		t.FailNow()
	}
}

func TestRequestTimeout(t *testing.T) {
	bc := conn.NewBufConn(false, false)

	ps := NewPeerStore("")
	p := New(bc, ps, "")
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
	bc := conn.NewBufConn(true, false)

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
		Type:    msg.ResponseMessage,
		Payload: responsePayloadBytes,
	}
	responseBytes, err := json.Marshal(responseMsg)
	if err != nil {
		t.FailNow()
	}
	bc.Buf = bytes.NewBuffer(responseBytes)

	ps := NewPeerStore("")
	p := New(bc, ps, "")

	responseHandler := func(res *msg.Response) {
		responseChan <- res
	}

	go p.Dispatch()

	err = p.Request(req, responseHandler)
	if err != nil {
		t.FailNow()
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
	bc := conn.NewBufConn(false, false)

	push := msg.Push{
		ResourceType: msg.ResourceTransaction,
		Resource:     "HEY!",
	}

	pushPayloadBytes, _ := json.Marshal(push)

	pushMsg := msg.Message{
		Type:    msg.PushMessage,
		Payload: pushPayloadBytes,
	}

	pushBytes, _ := json.Marshal(pushMsg)
	bc.Buf = bytes.NewBuffer(pushBytes)

	ps := NewPeerStore("")
	p := New(bc, ps, "")

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
	bc := conn.NewBufConn(false, false)

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
	bc.Buf = bytes.NewBuffer(requestBytes)

	ps := NewPeerStore("")
	p := New(bc, ps, "")

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

func TestSendBlock(t *testing.T) {
	fa := conn.TestAddr{Addr: "127.0.0.1"}

	var fc net.Conn
	fc = conn.NewBufConn(false, false)
	ps := NewPeerStore("")
	p := New(fc, ps, fa.String())

	block := blockchain.NewTestBlock()

	push := msg.Push{
		ResourceType: msg.ResourceBlock,
		Resource:     block,
	}

	err := p.Push(push)
	assert.Nil(t, err)
	payload, err := msg.Read(p.Connection)
	assert.Nil(t, err)

	switch payload.(type) {
	case *msg.Push:
		receivedPush := payload.(*msg.Push)
		assert.Equal(t, receivedPush.ResourceType, msg.ResourceBlock)
		blockBytes, err := json.Marshal(receivedPush.Resource)
		assert.Nil(t, err)
		receivedBlock, err := blockchain.DecodeBlockJSON(blockBytes)
		assert.Nil(t, err)
		assert.Equal(t, blockchain.HashSum(receivedBlock), blockchain.HashSum(block))
	default:
		t.FailNow()
	}
}
