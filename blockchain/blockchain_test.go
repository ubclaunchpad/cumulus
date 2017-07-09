package blockchain

import (
	"bytes"
	"reflect"
	"testing"

	log "github.com/Sirupsen/logrus"
)

func TestMain(t *testing.T) {
	log.SetLevel(log.DebugLevel)
}

func TestEncodeDecodeBlockChain(t *testing.T) {
	b1 := NewBlockChain()

	buf := bytes.NewBuffer(make([]byte, 0, b1.Len()))

	b1.Encode(buf)
	b2 := DecodeBlockChain(buf)

	if HashSum(b1) != HashSum(b2) {
		t.Fail()
	}
}

func TestCopyBlock(t *testing.T) {
	bc, _ := NewValidChainAndBlock()
	b := bc.CopyBlockByIndex(0)

	if !reflect.DeepEqual(b, bc.Blocks[0]) {
		t.FailNow()
	}

	// Enforce copy.
	if b == bc.Blocks[0] {
		t.FailNow()
	}
}
