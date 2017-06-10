package blockchain

import (
	"bytes"
	"testing"

	log "github.com/Sirupsen/logrus"
)

func TestMain(t *testing.T) {
	log.SetLevel(log.DebugLevel)
}

func TestEncodeDecodeBlockChain(t *testing.T) {
	b1 := newBlockChain()

	buf := bytes.NewBuffer(make([]byte, 0, b1.Len()))

	b1.Encode(buf)
	b2 := DecodeBlockChain(buf)

	if HashSum(b1) != HashSum(b2) {
		t.Fail()
	}
}
