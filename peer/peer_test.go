package peer

import (
	"fmt"
	"testing"

	log "github.com/Sirupsen/logrus"
	"github.com/google/uuid"
	"github.com/ubclaunchpad/cumulus/message"
	"github.com/ubclaunchpad/cumulus/subnet"
)

func TestMain(t *testing.T) {
	// Disable logging for tests
	log.SetLevel(log.FatalLevel)
}

func TestNewDefault(t *testing.T) {
	h, err := New(DefaultIP, DefaultPort)
	if err != nil {
		t.Fail()
	}

	if h == nil {
		t.Fail()
	}

	if h.Peerstore() == nil {
		t.Fail()
	}
}

func TestNewValidPort(t *testing.T) {
	h, err := New(DefaultIP, 8000)
	if err != nil {
		t.Fail()
	}

	if h == nil {
		t.Fail()
	}

	if h.Peerstore() == nil {
		t.Fail()
	}
}

func TestNewValidIP(t *testing.T) {
	_, err := New("123.211.231.45", DefaultPort)
	if err == nil {
		t.Fail()
	}
}

func TestNewInvalidIP(t *testing.T) {
	_, err := New("asdfasdf", 123)
	if err == nil {
		t.Fail()
	}
}

func TestExtractPeerInfoValidMultiAddr(t *testing.T) {
	peerma := "/ip4/127.0.0.1/tcp/8765/ipfs/QmQdfp9Ug4MoLRsBToDPN2aQhg2jPtmmA8UidQUTXGjZcy"
	pid, ma, err := extractPeerInfo(peerma)

	if err != nil {
		t.Fail()
	}

	if pid.Pretty() != "QmQdfp9Ug4MoLRsBToDPN2aQhg2jPtmmA8UidQUTXGjZcy" {
		t.Fail()
	}

	if ma.String() != "/ip4/127.0.0.1/tcp/8765" {
		t.Fail()
	}
}

func TestExtractPeerInfoInvalidIP(t *testing.T) {
	peerma := "/ip4/203.532.211.5/tcp/8765/ipfs/Qmb89FuJ8UG3dpgUqEYu9eUqK474uP3mx32WnQ7kePXp8N"
	_, _, err := extractPeerInfo(peerma)

	if err == nil {
		t.Fail()
	}
}

func TestReceiveValidMessage(t *testing.T) {
	sender, err := New(DefaultIP, DefaultPort)
	if err != nil {
		t.FailNow()
	}

	sender.SetStreamHandler(CumulusProtocol, sender.Receive)

	receiver, err := New(DefaultIP, 8080)
	if err != nil {
		t.FailNow()
	}

	receiver.SetStreamHandler(CumulusProtocol, receiver.Receive)

	receiverMultiAddr := fmt.Sprintf("%s/ipfs/%s",
		receiver.Addrs()[0], receiver.ID().Pretty())

	stream, err := sender.Connect(receiverMultiAddr)
	if err != nil {
		t.FailNow()
	}

	_, err = stream.Write([]byte("This is a test\n"))
	if err != nil {
		t.FailNow()
	}
}

func TestReceiveInvalidAddress(t *testing.T) {
	receiver, err := New(DefaultIP, DefaultPort)
	if err != nil {
		t.Fail()
	}

	sender, err := New(DefaultIP, 8080)
	if err != nil {
		t.Fail()
	}

	receiver.SetStreamHandler(CumulusProtocol, receiver.Receive)

	_, err = sender.Connect(receiver.Addrs()[0].String())
	if err == nil {
		t.Fail()
	}
}

func TestConnectWithSubnetFull(t *testing.T) {
	testPeer, err := New(DefaultIP, DefaultPort)
	if err != nil {
		t.Fail()
	}
	peers := make([]*Peer, subnet.DefaultMaxPeers)

	for i := 0; i < subnet.DefaultMaxPeers; i++ {
		peers[i], err = New(DefaultIP, 8080+i)
		if err != nil {
			t.FailNow()
		}
		peers[i].SetStreamHandler(CumulusProtocol, peers[i].Receive)
		ma, maErr := NewMultiaddr(peers[i].Addrs()[0], peers[i].ID())
		if maErr != nil {
			t.FailNow()
		}
		_, err = testPeer.Connect(ma.String())
		if err != nil {
			t.FailNow()
		}
	}

	lastPeer, err := New(DefaultIP, 8081+subnet.DefaultMaxPeers)
	if err != nil {
		t.FailNow()
	}
	_, err = testPeer.Connect(lastPeer.Addrs()[0].String())
	if err == nil {
		t.FailNow()
	}
}

func TestReceiveWithSubnetFull(t *testing.T) {
	targetPeer, err := New(DefaultIP, DefaultPort)
	if err != nil {
		t.FailNow()
	}
	targetPeer.SetStreamHandler(CumulusProtocol, targetPeer.Receive)
	ma, err := NewMultiaddr(targetPeer.Addrs()[0], targetPeer.ID())
	if err != nil {
		t.FailNow()
	}

	for i := 0; i < subnet.DefaultMaxPeers; i++ {
		p, err := New(DefaultIP, 8080+i)
		if err != nil {
			t.FailNow()
		}
		_, err = p.Connect(ma.String())
		if err != nil {
			t.FailNow()
		}
	}

	lastPeer, err := New(DefaultIP, 8080+subnet.DefaultMaxPeers)
	if err != nil {
		t.FailNow()
	}

	_, err = lastPeer.Connect(ma.String())
	if err != nil {
		t.FailNow()
	}
}

func TestRequest(t *testing.T) {
	requester, err := New(DefaultIP, DefaultPort)
	if err != nil {
		t.Fail()
	}

	responder, err := New(DefaultIP, 8080)
	if err != nil {
		t.Fail()
	}

	requester.SetStreamHandler(CumulusProtocol, requester.Receive)
	responder.SetStreamHandler(CumulusProtocol, responder.Receive)
	responderAddr, err := NewMultiaddr(responder.Addrs()[0], responder.ID())
	if err != nil {
		t.Fail()
	}

	stream, err := requester.Connect(responderAddr.String())
	if err != nil {
		fmt.Println("Failed to connect to remote peer")
		t.Fail()
	}

	request := message.Request{
		ID:           uuid.New().String(),
		ResourceType: message.ResourcePeerInfo,
		Params:       nil,
	}
	response, err := requester.Request(request, *stream)
	if err != nil {
		fmt.Printf("Failed to make request: %s", err)
		t.FailNow()
	} else if response.Error != nil {
		fmt.Printf("Remote peer returned response %s", response.Error)
		t.FailNow()
	}
}
