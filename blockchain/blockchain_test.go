package blockchain

import (
	"bytes"
	crand "crypto/rand"
	"testing"

	log "github.com/Sirupsen/logrus"
)

func TestMain(t *testing.T) {
	log.SetLevel(log.DebugLevel)
}

func TestValidTransactionNotInBlock(t *testing.T) {
	tr, _ := newTransactionValue(newWallet(), newWallet(), 1)
	bc := newValidBlockChainFixture()

	if bc.ValidTransaction(tr) {
		t.Fail()
	}
}

func TestValidTransactionInputsFail(t *testing.T) {
	// 2 + 2 = 5 ?
	bc := newValidBlockChainFixture()
	tr := bc.Blocks[1].Transactions[0]
	tr.Outputs[0].Amount = 5

	if bc.ValidTransaction(tr) {
		t.Fail()
	}
}

func TestValidTransactionSignatureFail(t *testing.T) {
	bc := newValidBlockChainFixture()
	tr := bc.Blocks[1].Transactions[0]

	fakeSender := newWallet()
	tr, _ = tr.TxBody.Sign(fakeSender, crand.Reader)
	bc.Blocks[1].Transactions[0] = tr

	if bc.ValidTransaction(tr) {
		t.Fail()
	}
}

func TestValidTransactionPass(t *testing.T) {
	bc := newValidBlockChainFixture()
	tr := bc.Blocks[1].Transactions[0]

	if !bc.ValidTransaction(tr) {
		t.Fail()
	}
}

func TestValidBlockTransactionInvalid(t *testing.T) {
	bc := newValidBlockChainFixture()
	tr := bc.Blocks[1].Transactions[0]
	tr.Outputs[0].Amount = 5

	if bc.ValidBlock(bc.Blocks[1]) {
		t.Fail()
	}
}

func TestValidBlockNumberWrong(t *testing.T) {
	bc := newValidBlockChainFixture()
	bc.Blocks[1].BlockNumber = 2

	if bc.ValidBlock(bc.Blocks[1]) {
		t.Fail()
	}
}

func TestValidBlockHashWrong(t *testing.T) {
	bc := newValidBlockChainFixture()
	bc.Blocks[0].BlockHeader.LastBlock = newHash()

	if bc.ValidBlock(bc.Blocks[1]) {
		t.Fail()
	}
}

func TestValidBlock(t *testing.T) {
	bc := newValidBlockChainFixture()
	if !bc.ValidBlock(bc.Blocks[1]) {
		t.Fail()
	}
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
