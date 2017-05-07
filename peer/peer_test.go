package peer

import (
	"context"
	"fmt"
	"testing"

	pstore "github.com/libp2p/go-libp2p-peerstore"
	log "github.com/sirupsen/logrus"
)

func TestMain(t *testing.T) {
	// Disable logging for tests
	log.SetLevel(log.FatalLevel)
}

func TestNewPeerDefault(t *testing.T) {
	h, err := NewPeer(DefaultIP, DefaultPort)
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

func TestNewPeerValidPort(t *testing.T) {
	h, err := NewPeer(DefaultIP, 8000)
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

func TestNewPeerValidIP(t *testing.T) {
	_, err := NewPeer("123.211.231.45", DefaultPort)
	if err == nil {
		t.Fail()
	}
}

func TestNewPeerInvalidIP(t *testing.T) {
	_, err := NewPeer("asdfasdf", 123)
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
	receiver, err := NewPeer(DefaultIP, DefaultPort)
	if err != nil {
		t.Fail()
	}

	sender, err := NewPeer(DefaultIP, 8080)
	if err != nil {
		t.Fail()
	}

	receiver.SetStreamHandler(CumulusProtocol, receiver.Receive)
	sender.SetStreamHandler(CumulusProtocol, sender.Receive)

	receiverMA := fmt.Sprintf("%s/ipfs/%s",
		receiver.Addrs()[0].String(), receiver.ID().Pretty())

	receiverID, receiverAddr, err := ExtractPeerInfo(receiverMA)
	if err != nil {
		t.Fail()
	}

	sender.Peerstore().AddAddr(receiverID, receiverAddr, pstore.PermanentAddrTTL)
	stream, err := sender.NewStream(context.Background(), receiverID,
		CumulusProtocol)
	if err != nil {
		t.Fail()
	}

	_, err = stream.Write([]byte("Hello, world!\n"))
	if err != nil {
		t.Fail()
	}

	stream.Close()
}

func TestReceiveInvalidMessage(t *testing.T) {
	receiver, err := NewPeer(DefaultIP, DefaultPort)
	if err != nil {
		t.Fail()
	}

	sender, err := NewPeer(DefaultIP, 8080)
	if err != nil {
		t.Fail()
	}

	receiver.SetStreamHandler(CumulusProtocol, receiver.Receive)

	receiverMA := fmt.Sprintf("%s/ipfs/%s",
		receiver.Addrs()[0].String(), receiver.ID().Pretty())

	receiverID, receiverAddr, err := ExtractPeerInfo(receiverMA)
	if err != nil {
		t.Fail()
	}

	sender.Peerstore().AddAddr(receiverID, receiverAddr, pstore.PermanentAddrTTL)
	stream, err := sender.NewStream(context.Background(), receiverID,
		CumulusProtocol)
	if err != nil {
		t.Fail()
	}

	_, err = stream.Write([]byte("Hello, world!"))
	if err != nil {
		t.Fail()
	}

	stream.Close()
}
