package blockchain

import (
	"crypto/rand"
	"fmt"
	"testing"
)

func TestValidTransactionNoInputTransaction(t *testing.T) {
	tr, _ := NewTransactionValue(NewWallet(), NewWallet(), 1, 0)
	bc, _ := NewValidBlockChainFixture()

	valid, code := bc.ValidTransaction(tr)

	if valid {
		t.Fail()
	}
	if code != NoInputTransaction {
		t.Fail()
	}
}

func TestValidTransactionOverspend(t *testing.T) {
	// 2 + 2 = 5 ?
	bc, _ := NewValidBlockChainFixture()
	tr := bc.Blocks[1].Transactions[0]
	tr.Outputs[0].Amount = 5

	valid, code := bc.ValidTransaction(tr)

	if valid {
		t.Fail()
	}
	if code != Overspend {
		fmt.Println(code)
		t.Fail()
	}
}

func TestValidTransactionSignatureFail(t *testing.T) {
	bc, _ := NewValidBlockChainFixture()
	tr := bc.Blocks[1].Transactions[0]

	fakeSender := NewWallet()
	tr, _ = tr.TxBody.Sign(fakeSender, rand.Reader)
	bc.Blocks[1].Transactions[0] = tr

	valid, code := bc.ValidTransaction(tr)
	if valid {
		t.Fail()
	}
	if code != BadSig {
		t.Fail()
	}
}

func TestValidTransactionPass(t *testing.T) {
	bc, b := NewValidChainAndBlock()
	tr := b.Transactions[0]

	valid, code := bc.ValidTransaction(tr)

	if !valid {
		t.Fail()
	}
	if code != ValidTransaction {
		t.Fail()
	}
}

func TestTransactionRespend(t *testing.T) {
	bc, _ := NewValidBlockChainFixture()
	trC := bc.Blocks[1].Transactions[0]
	b := NewOutputBlock([]*Transaction{trC}, bc.Blocks[1])
	bc.AppendBlock(b, NewWallet().Public())

	valid, code := bc.ValidTransaction(trC)

	if valid {
		t.Fail()
	}
	if code != Respend {
		t.Fail()
	}
}

func TestValidBlockBadTransaction(t *testing.T) {
	bc, _ := NewValidBlockChainFixture()
	tr := bc.Blocks[1].Transactions[0]
	tr.Outputs[0].Amount = 5

	valid, code := bc.ValidBlock(bc.Blocks[1])

	if valid {
		t.Fail()
	}
	if code != BadTransaction {
		t.Fail()
	}
}

func TestValidBlocBadBlockNumber(t *testing.T) {
	bc, _ := NewValidBlockChainFixture()
	bc.Blocks[1].BlockNumber = 2

	valid, code := bc.ValidBlock(bc.Blocks[1])

	if valid {
		t.Fail()
	}
	if code != BadBlockNumber {
		t.Fail()
	}
}

func TestValidBlockBadHash(t *testing.T) {
	bc, b := NewValidChainAndBlock()
	b.BlockHeader.LastBlock = NewHash()

	valid, code := bc.ValidBlock(b)

	if valid {
		t.Fail()
	}
	if code != BadHash {
		t.Fail()
	}
}

func TestValidBlock(t *testing.T) {
	bc, b := NewValidChainAndBlock()

	valid, code := bc.ValidBlock(b)

	if !valid {
		t.Fail()
	}
	if code != ValidBlock {
		t.Fail()
	}
}

func TestBlockDoubleSpend(t *testing.T) {
	// block should fail to be valid if there exists two transactions
	// referencing the same input, but output > input (double spend attack)
	bc, b := NewValidChainAndBlock()
	b.Transactions = append(b.Transactions, b.Transactions[0])

	valid, code := bc.ValidBlock(b)

	if valid {
		t.Fail()
	}
	if code != DoubleSpend {
		t.Fail()
	}
}

func TestValidBlockNegNumber(t *testing.T) {
	bc, b := NewValidChainAndBlock()
	b.BlockNumber = uint32(len(bc.Blocks))

	valid, code := bc.ValidBlock(b)

	if valid {
		t.Fail()
	}
	if code != BadBlockNumber {
		t.Fail()
	}
}
