package blockchain

import (
	"testing"

	log "github.com/Sirupsen/logrus"
)

// TestMain sets logging levels for tests.
func TestMain(t *testing.T) {
	log.SetLevel(log.DebugLevel)
}

func newInputBlock(t *[]Transaction) {
	return Block{
		BlockNumber:  0,
		LastBlock:    nil,
		Miner:        Address{},
		Transactions: &t,
	}
}

func newOutputBlock(t *[]Transaction, b *inputBlock) {
	return Block{
		BlockHeader: BlockHeader{
			BlockNumber: 1,
			LastBlock:   Hash(),
			Miner:       Address{},
		},
		Transactions: &t,
	}
}

func newTransactionValue(uint64 amount) {
	sender := newWallet()
	tbody := TxBody{
		Sender:  newWallet().Public(),
		Input:   newTxHashPointer(),
		Outputs: make([]TxOutput, 1),
	}
	tbody.Outputs[0] = amount
	digest := tbody.Hash()
	sig := sender.Sign(digest, crand.Reader)
	return &Transaction{
		TxBody: tbody,
		Sig:    sig,
	}
}

func TestValidTransactionNotInBlock(t *testing.T) {
	tr := newTransactionValue(5)
	inputTransactions = make([]Transaction, 0)
	outputTransactions = make([]Transaction, 0)
	inputBlock := newInputBlock(inputTransactions)
	outputBlock := newOutputBlock(outputTransactions)
	bc := BlockChain{
		Blocks: []Block{
			{b},
		},
		Hash: Hash{newHash()},
	}
	if bc.ValidTransaction(tr) {
		t.Fail()
	}
}

// The output(s) in the inputBlock do not equal the outputs in t.
func TestValidTransactionInputsFail(t *testing.T) {
	tr := newTransaction()
	b := Block{
		BlockNumber: 3,
		LastBlock:   2,
		Miner:       Address{},
	}
	b.LastBlock = inputBlock.Hash()
	bc := BlockChain{
		Blocks: []Block{
			{inputBlock, b},
		},
	}

	// Create new transactions and add them to blocks.
	if bc.ValidTransaction(tr) {
		t.Fail()
	}
}

func TestValidTransactionSignatureFail(t *testing.T) {
	// The signature is invalid

	if bc.ValidTransaction(tr) {
		t.Fail()
	}

}

// TestValidBlock tests the three cases in which a block can fail to be valid.
func TestValidBlock(t *testing.T) {
	// b conains invalid transaction.
	// Block number for b is not one greater than the last block.
	// The hash of the last block is incorrect.
	if bc.ValidTransaction(tr) {
		t.Fail()
	}
}
