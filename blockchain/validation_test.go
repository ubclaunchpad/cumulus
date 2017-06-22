package blockchain

import (
	"crypto/rand"
	"fmt"
	"math"
	"math/big"
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
	bc.AppendBlock(b)

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

func TestValidBlockBadBlockNumber(t *testing.T) {
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

func TestValidBlockBadTime(t *testing.T) {
	bc, b := NewValidChainAndBlock()
	b.Time = 0
	valid, code := bc.ValidBlock(b)

	if valid {
		t.Fail()
	}
	if code != BadTime {
		t.Fail()
	}

	b.Time = math.MaxUint32
	valid, code = bc.ValidBlock(b)

	if valid {
		t.Fail()
	}
	if code != BadTime {
		t.Fail()
	}
}

func TestValidBlockBadTarget(t *testing.T) {
	bc, b := NewValidChainAndBlock()
	b.Target = BigIntToHash(new(big.Int).Add(MinDifficulty, big.NewInt(1)))
	valid, code := bc.ValidBlock(b)

	if valid {
		t.Fail()
	}
	if code != BadTarget {
		t.Fail()
	}

	b.Target = BigIntToHash(new(big.Int).Sub(MaxDifficulty, big.NewInt(1)))
	valid, code = bc.ValidBlock(b)

	if valid {
		t.Fail()
	}
	if code != BadTarget {
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

func TestValidBlockBigNumber(t *testing.T) {
	bc, b := NewValidChainAndBlock()
	b.BlockNumber = uint32(len(bc.Blocks)) + 1

	valid, code := bc.ValidBlock(b)

	if valid {
		t.Fail()
	}
	if code != BadBlockNumber {
		t.Fail()
	}
}

func TestBigExp(t *testing.T) {
	a := big.NewInt(1)
	b := BigExp(0, 0)

	if a.Cmp(b) != 0 {
		t.Fail()
	}

	a = big.NewInt(1)
	b = BigExp(10, -2)

	if a.Cmp(b) != 0 {
		t.Fail()
	}

	a = new(big.Int).Exp(
		big.NewInt(int64(2)),
		big.NewInt(int64(256)),
		big.NewInt(0),
	)
	b = BigExp(2, 256)

	if a.Cmp(b) != 0 {
		t.Fail()
	}
}
