package peer

import (
	"fmt"
	"testing"

	log "github.com/sirupsen/logrus"
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
	pid, ma, err := ExtractPeerInfo(peerma)

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
	_, _, err := ExtractPeerInfo(peerma)

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
