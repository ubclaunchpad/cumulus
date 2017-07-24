package blockchain

import (
	"bytes"
	"reflect"
	"testing"

	log "github.com/Sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestMain(t *testing.T) {
	log.SetLevel(log.DebugLevel)
}

func TestEncodeDecodeBlockChain(t *testing.T) {
	b1 := NewTestBlockChain()

	buf := bytes.NewBuffer(make([]byte, 0, b1.Len()))

	b1.Encode(buf)
	b2 := DecodeBlockChain(buf)

	if HashSum(b1) != HashSum(b2) {
		t.Fail()
	}
}

func TestCopyBlock(t *testing.T) {
	bc, _ := NewValidTestChainAndBlock()
	b, _ := bc.CopyBlockByIndex(0)

	if !reflect.DeepEqual(b, bc.Blocks[0]) {
		t.FailNow()
	}

	// Enforce copy.
	if b == bc.Blocks[0] {
		t.FailNow()
	}
}

func TestWalletRepr(t *testing.T) {
	w := NewWallet()
	assert.Equal(t, len(w.Public().Repr()), 40)
}
