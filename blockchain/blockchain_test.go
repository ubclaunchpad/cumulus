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
	sender := newWallet()
	tr, _ := newTransactionValue(sender, newWallet(), 1)

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
	original := newWallet()
	sender := newWallet()
	recipient := newWallet()

	trA, _ := newTransactionValue(original, sender, 2)
	trA.Outputs = append(trA.Outputs, TxOutput{
		Amount:    2,
		Recipient: sender.Public(),
	})

	trB, _ := newTransactionValue(sender, recipient, 5)
	trB.Input.Hash = HashSum(trA)

	inputTransactions := []*Transaction{trA}
	outputTransactions := []*Transaction{trB}

	inputBlock := newInputBlock(inputTransactions)
	outputBlock := newOutputBlock(outputTransactions, inputBlock)

	bc := BlockChain{
		Blocks: []*Block{inputBlock, outputBlock},
		Head:   newHash(),
	}

	if bc.ValidTransaction(trB) {
		t.Fail()
	}
}

func TestValidTransactionSignatureFail(t *testing.T) {
	original := newWallet()
	sender := newWallet()
	recipient := newWallet()

	trA, _ := newTransactionValue(original, sender, 2)
	trA.Outputs = append(trA.Outputs, TxOutput{
		Amount:    2,
		Recipient: sender.Public(),
	})

	trB, _ := newTransactionValue(sender, recipient, 4)
	trB.Input.Hash = HashSum(trA)

	inputTransactions := []*Transaction{trA}
	outputTransactions := []*Transaction{trB}

	inputBlock := newInputBlock(inputTransactions)
	outputBlock := newOutputBlock(outputTransactions, inputBlock)

	bc := BlockChain{
		Blocks: []*Block{inputBlock, outputBlock},
		Head:   newHash(),
	}

	fakeSender := newWallet()
	trB, _ = trB.TxBody.Sign(fakeSender, crand.Reader)

	if bc.ValidTransaction(trB) {
		t.Fail()
	}
}

func TestValidTransactionPass(t *testing.T) {
	original := newWallet()
	sender := newWallet()
	recipient := newWallet()

	trA, _ := newTransactionValue(original, sender, 2)
	trA.Outputs = append(trA.Outputs, TxOutput{
		Amount:    2,
		Recipient: sender.Public(),
	})

	trB, _ := newTransactionValue(sender, recipient, 4)
	// trB, _ = trB.TxBody.Sign(sender, crand.Reader)
	trB.Input.Hash = HashSum(trA)

	trB, _ = trB.TxBody.Sign(sender, crand.Reader)

	inputTransactions := []*Transaction{trA}
	outputTransactions := []*Transaction{trB}

	inputBlock := newInputBlock(inputTransactions)
	outputBlock := newOutputBlock(outputTransactions, inputBlock)

	bc := BlockChain{
		Blocks: []*Block{inputBlock, outputBlock},
		Head:   newHash(),
	}

	if !bc.ValidTransaction(trB) {
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
