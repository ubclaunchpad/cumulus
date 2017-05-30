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
	bc, _ := newValidBlockChainFixture()

	if bc.ValidTransaction(tr) {
		t.Fail()
	}
}

func TestValidTransactionInputsFail(t *testing.T) {
	// 2 + 2 = 5 ?
	bc, _ := newValidBlockChainFixture()
	tr := bc.Blocks[1].Transactions[0]
	tr.Outputs[0].Amount = 5

	if bc.ValidTransaction(tr) {
		t.Fail()
	}
}

func TestValidTransactionSignatureFail(t *testing.T) {
	bc, _ := newValidBlockChainFixture()
	tr := bc.Blocks[1].Transactions[0]

	fakeSender := newWallet()
	tr, _ = tr.TxBody.Sign(fakeSender, crand.Reader)
	bc.Blocks[1].Transactions[0] = tr

	if bc.ValidTransaction(tr) {
		t.Fail()
	}
}

func TestValidTransactionPass(t *testing.T) {
	bc, b := newValidChainAndBlock()
	tr := b.Transactions[0]

	if !bc.ValidTransaction(tr) {
		t.Fail()
	}
}

func TestTransactionRespend(t *testing.T) {
	bc, _ := newValidBlockChainFixture()
	trC := bc.Blocks[1].Transactions[0]
	b := newOutputBlock([]*Transaction{trC}, bc.Blocks[1])
	bc.AppendBlock(b, newWallet().Public())

	if bc.ValidTransaction(trC) {
		t.Fail()
	}
}

func TestValidBlockTransactionInvalid(t *testing.T) {
	bc, _ := newValidBlockChainFixture()
	tr := bc.Blocks[1].Transactions[0]
	tr.Outputs[0].Amount = 5

	if bc.ValidBlock(bc.Blocks[1]) {
		t.Fail()
	}
}

func TestValidBlockNumberWrong(t *testing.T) {
	bc, _ := newValidBlockChainFixture()
	bc.Blocks[1].BlockNumber = 2

	if bc.ValidBlock(bc.Blocks[1]) {
		t.Fail()
	}
}

func TestValidBlockHashWrong(t *testing.T) {
	bc, b := newValidChainAndBlock()
	b.BlockHeader.LastBlock = newHash()

	if bc.ValidBlock(b) {
		t.Fail()
	}
}

func TestValidBlock(t *testing.T) {
	bc, b := newValidChainAndBlock()

	if !bc.ValidBlock(b) {
		t.Fail()
	}
}

func TestBlockTwoInputs(t *testing.T) {
	// block should fail to be valid if there exists two transactions
	// referencing the same input, but output > input (double spend attack)
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
