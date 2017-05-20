package blockchain

import (
	"bytes"
	crand "crypto/rand"
	"testing"

	log "github.com/Sirupsen/logrus"
)

// TestMain sets logging levels for tests.
func TestMain(t *testing.T) {
	log.SetLevel(log.DebugLevel)
}

func TestValidTransactionNotInBlock(t *testing.T) {
	tr := newTransactionValue(5, Address{})

	inputTransactions := make([]*Transaction, 0)
	outputTransactions := make([]*Transaction, 0)

	inputBlock := newInputBlock(inputTransactions)
	outputBlock := newOutputBlock(outputTransactions, inputBlock)

	bc := BlockChain{
		Blocks: []*Block{inputBlock, outputBlock},
		Head:   newHash(),
	}

	if bc.ValidTransaction(tr) {
		t.Fail()
	}
}

func TestValidTransactionInputsFail(t *testing.T) {
	// 2 + 2 = 5 ?
	trA := newTransactionValue(2, newWallet().Public())
	trB := newTransactionValue(2, newWallet().Public())
	trC := newTransactionValue(5, newWallet().Public())

	inputTransactions := []*Transaction{trA, trB}
	outputTransactions := []*Transaction{trC}

	inputBlock := newInputBlock(inputTransactions)
	outputBlock := newOutputBlock(outputTransactions, inputBlock)

	bc := BlockChain{
		Blocks: []*Block{inputBlock, outputBlock},
		Head:   newHash(),
	}

	if bc.ValidTransaction(trC) {
		t.Fail()
	}
}

func TestValidTransactionSignatureFail(t *testing.T) {
	tr := newTransactionValue(2, newWallet().Public())

	// Fake sender puts in a silly signature.
	fakeSender := newWallet()
	fakeTransaction := newTransaction()
	digest := HashSum(fakeTransaction.TxBody)
	tr.Sig, _ = fakeSender.Sign(digest, crand.Reader)

	transactions := []*Transaction{tr}
	b := newInputBlock(transactions)

	bc := BlockChain{
		Blocks: []*Block{b},
		Head:   newHash(),
	}

	if bc.ValidTransaction(tr) {
		t.Fail()
	}

}

// TestValidBlock tests the three cases in which a block can fail to be valid.
func TestValidBlock(t *testing.T) {
	// b conains invalid transaction.
	// Block number for b is not one greater than the last block.
	// The hash of the last block is incorrect.
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
