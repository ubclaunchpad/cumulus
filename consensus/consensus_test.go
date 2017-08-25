package consensus

import (
	"math/rand"
	"testing"

	crand "crypto/rand"

	"github.com/stretchr/testify/assert"
	"github.com/ubclaunchpad/cumulus/blockchain"
	c "github.com/ubclaunchpad/cumulus/common/constants"
	"github.com/ubclaunchpad/cumulus/common/util"
)

// VerifyTransaction Tests

func TestVerifyTransactionNilTransaction(t *testing.T) {
	bc, _ := blockchain.NewValidBlockChainFixture()

	valid, code := VerifyTransaction(bc, nil)

	assert.False(t, valid)
	assert.Equal(t, code, NilTransaction)
}

func TestVerifyTransactionNoInputTransaction(t *testing.T) {
	txn := blockchain.NewTestTransaction()
	bc, _ := blockchain.NewValidBlockChainFixture()
	txn.Inputs = []blockchain.TxHashPointer{}
	valid, code := VerifyTransaction(bc, txn)

	assert.False(t, valid)
	assert.Equal(t, code, NoInputTransactions)
}

func TestVerifyTransactionOverspend(t *testing.T) {
	// 2 + 2 = 5 ?
	bc, _ := blockchain.NewValidBlockChainFixture()
	tr := bc.Blocks[1].Transactions[1]
	tr.Outputs[0].Amount = 5

	valid, code := VerifyTransaction(bc, tr)

	assert.False(t, valid)
	assert.Equal(t, code, Overspend)
}

func TestVerifyTransactionSignatureFail(t *testing.T) {
	bc, txn := blockchain.NewValidChainAndTxn()

	// Resign txn with fake wallet.
	fakeSender := blockchain.NewWallet()
	txn, _ = txn.TxBody.Sign(*fakeSender, crand.Reader)

	valid, code := VerifyTransaction(bc, txn)

	assert.False(t, valid)
	assert.Equal(t, code, BadSig)
}

func TestVerifyTransactionPass(t *testing.T) {
	bc, txn := blockchain.NewValidChainAndTxn()

	valid, code := VerifyTransaction(bc, txn)

	assert.True(t, valid)
	assert.Equal(t, code, ValidTransaction)
}

func TestTransactionRespend(t *testing.T) {
	bc, _ := blockchain.NewValidTestChainAndBlock()
	trC := bc.Blocks[2].Transactions[1]

	valid, code := VerifyTransaction(bc, trC)

	assert.False(t, valid)
	assert.Equal(t, code, Respend)
}

// VerifyBlock Tests

func TestVerifyBlockNilBlock(t *testing.T) {
	bc, _ := blockchain.NewValidBlockChainFixture()

	valid, code := VerifyBlock(bc, nil)

	assert.False(t, valid)
	assert.Equal(t, code, NilBlock)
}

func TestVerifyBlockBadNonce(t *testing.T) {
	bc, b := blockchain.NewValidTestChainAndBlock()
	b.Target = blockchain.BigIntToHash(c.Big1)
	CurrentDifficulty = c.MaxTarget
	valid, code := VerifyBlock(bc, b)
	CurrentDifficulty = c.MinTarget

	assert.False(t, valid)
	assert.Equal(t, code, BadNonce)
}

func TestVerifyBlockBadGenesisBlock(t *testing.T) {
	miner := blockchain.NewWallet()
	currentTarget := blockchain.BigIntToHash(util.BigExp(2, 255))
	currentBlockReward := uint64(25)
	gb := blockchain.Genesis(miner.Public(), currentTarget, currentBlockReward, []byte{})
	bc := &blockchain.BlockChain{
		Blocks: []*blockchain.Block{},
		Head:   blockchain.NilHash,
	}

	valid, code := VerifyBlock(bc, gb)

	assert.False(t, valid)
	assert.Equal(t, code, BadGenesisBlock)
}

func TestVerifyBlockBadTransaction(t *testing.T) {
	bc, b := blockchain.NewValidTestChainAndBlock()

	// This would be an overspend on alices part (she only has 3 coins here).
	b.Transactions[1].Outputs[0].Amount = 5

	valid, code := VerifyBlock(bc, b)

	assert.False(t, valid)
	assert.Equal(t, code, BadTransaction)
}

func TestVerifyBlockBadBlockNumber(t *testing.T) {
	bc, _ := blockchain.NewValidBlockChainFixture()
	bc.Blocks[1].BlockNumber = 2

	valid, code := VerifyBlock(bc, bc.Blocks[1])

	assert.False(t, valid)
	assert.Equal(t, code, BadBlockNumber)
}

func TestVerifyBlockBadHash(t *testing.T) {
	bc, b := blockchain.NewValidTestChainAndBlock()
	b.BlockHeader.LastBlock = blockchain.NewTestHash()

	valid, code := VerifyBlock(bc, b)

	if valid {
		t.Fail()
	}
	if code != BadHash {
		t.Fail()
	}
}

func TestVerifyBlockBadTime(t *testing.T) {
	bc, b := blockchain.NewValidTestChainAndBlock()
	b.Time = 0
	valid, code := VerifyBlock(bc, b)

	if valid {
		t.Fail()
	}
	if code != BadTime {
		t.Fail()
	}
}

func TestVerifyBlockBadTarget(t *testing.T) {
	bc, b := blockchain.NewValidTestChainAndBlock()
	b.Target = blockchain.BigIntToHash(util.BigAdd(c.MaxTarget, c.Big1))
	valid, code := VerifyBlock(bc, b)

	if valid {
		t.Fail()
	}
	if code != BadTarget {
		t.Fail()
	}

	b.Target = blockchain.BigIntToHash(util.BigSub(c.MinTarget, c.Big1))
	valid, code = VerifyBlock(bc, b)

	if valid {
		t.Fail()
	}
	if code != BadTarget {
		t.Fail()
	}
}

func TestVerifyBlock(t *testing.T) {
	bc, b := blockchain.NewValidTestChainAndBlock()

	valid, code := VerifyBlock(bc, b)

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
	bc, b := blockchain.NewValidTestChainAndBlock()
	b.Transactions = append(b.Transactions, b.Transactions[1])

	valid, code := VerifyBlock(bc, b)

	assert.False(t, valid)
	assert.Equal(t, code, DoubleSpend)
}

func TestVerifyBlockBigNumber(t *testing.T) {
	bc, b := blockchain.NewValidTestChainAndBlock()
	b.BlockNumber = uint32(len(bc.Blocks)) + 1

	valid, code := VerifyBlock(bc, b)

	if valid {
		t.Fail()
	}
	if code != BadBlockNumber {
		t.Fail()
	}
}

func TestVerifyBlockBadCloudBaseTransaction(t *testing.T) {
	bc, b := blockchain.NewValidTestChainAndBlock()
	b.Transactions[0] = blockchain.NewTestTransaction()

	valid, code := VerifyBlock(bc, b)

	if valid {
		t.Fail()
	}
	if code != BadCloudBaseTransaction {
		t.Fail()
	}
}

// VerifyCloudBase Tests

func TestVerifyCloudBaseNilCloudBase(t *testing.T) {
	bc, _ := blockchain.NewValidTestChainAndBlock()
	valid, code := VerifyCloudBase(bc, nil)

	if valid {
		t.Fail()
	}
	if code != NilCloudBaseTransaction {
		t.Fail()
	}
}

func TestVerifyCloudBaseBadCloudBaseBlockReward(t *testing.T) {
	bc, _ := blockchain.NewValidBlockChainFixture()
	b := bc.Blocks[0]
	b.Transactions[0].Outputs[0].Amount = CurrentBlockReward(bc) + 1

	valid, code := VerifyCloudBase(bc, b.GetCloudBaseTransaction())

	if valid {
		t.Fail()
	}

	if code != BadCloudBaseReward {
		t.Fail()
	}

}

func TestVerifyCloudBaseTransaction(t *testing.T) {
	bc, _ := blockchain.NewValidBlockChainFixture()
	b := bc.Blocks[0]
	valid, code := VerifyCloudBase(bc, b.GetCloudBaseTransaction())

	if !valid {
		t.Fail()
	}
	if code != ValidCloudBaseTransaction {
		t.Fail()
	}
}

func TestVerifyCloudBaseBadSender(t *testing.T) {
	w := blockchain.NewWallet()
	bc, _ := blockchain.NewValidBlockChainFixture()
	b := bc.Blocks[0]
	b.GetCloudBaseTransaction().Sender = w.Public()
	valid, code := VerifyCloudBase(bc, b.GetCloudBaseTransaction())

	if valid {
		t.Fail()
	}
	if code != BadCloudBaseSender {
		t.Fail()
	}
}

func TestVerifyCloudBaseBadBadInput(t *testing.T) {
	bc, _ := blockchain.NewValidBlockChainFixture()
	b := bc.Blocks[0]
	b.GetCloudBaseTransaction().TxBody.Inputs[0].BlockNumber = 1
	valid, code := VerifyCloudBase(bc, b.GetCloudBaseTransaction())

	if valid {
		t.Fail()
	}
	if code != BadCloudBaseInput {
		t.Fail()
	}

	bc, _ = blockchain.NewValidBlockChainFixture()
	b = bc.Blocks[0]
	b.GetCloudBaseTransaction().TxBody.Inputs[0].Hash = blockchain.NewTestHash()
	valid, code = VerifyCloudBase(bc, b.GetCloudBaseTransaction())

	if valid {
		t.Fail()
	}
	if code != BadCloudBaseInput {
		t.Fail()
	}

	bc, _ = blockchain.NewValidBlockChainFixture()
	b = bc.Blocks[0]
	b.GetCloudBaseTransaction().TxBody.Inputs[0].Index = 1
	valid, code = VerifyCloudBase(bc, b.GetCloudBaseTransaction())

	if valid {
		t.Fail()
	}
	if code != BadCloudBaseInput {
		t.Fail()
	}
}

func TestVerifyCloudBaseBadOutput(t *testing.T) {
	bc, _ := blockchain.NewValidBlockChainFixture()
	b := bc.Blocks[0]
	w := blockchain.NewWallet()
	b.GetCloudBaseTransaction().Outputs =
		append(
			b.GetCloudBaseTransaction().Outputs,
			blockchain.TxOutput{
				Amount:    25,
				Recipient: w.Public().Repr(),
			},
		)
	valid, code := VerifyCloudBase(bc, b.GetCloudBaseTransaction())

	if valid {
		t.Fail()
	}
	if code != BadCloudBaseOutput {
		t.Fail()
	}

	bc, _ = blockchain.NewValidBlockChainFixture()
	b = bc.Blocks[0]
	var emptyOutputs []blockchain.TxOutput
	b.GetCloudBaseTransaction().Outputs = emptyOutputs
	valid, code = VerifyCloudBase(bc, b.GetCloudBaseTransaction())

	if valid {
		t.Fail()
	}
	if code != BadCloudBaseOutput {
		t.Fail()
	}

	bc, _ = blockchain.NewValidBlockChainFixture()
	b = bc.Blocks[0]
	b.GetCloudBaseTransaction().Outputs[0].Recipient = blockchain.NilAddr.Repr()
	valid, code = VerifyCloudBase(bc, b.GetCloudBaseTransaction())

	if valid {
		t.Fail()
	}
	if code != BadCloudBaseOutput {
		t.Fail()
	}
}

func TestVerifyCloudBaseBadSig(t *testing.T) {
	bc, _ := blockchain.NewValidBlockChainFixture()
	b := bc.Blocks[0]
	w := blockchain.NewWallet()
	b.GetCloudBaseTransaction().Sig, _ =
		w.Sign(blockchain.NewTestHash(), crand.Reader)
	valid, code := VerifyCloudBase(bc, b.GetCloudBaseTransaction())

	if valid {
		t.Fail()
	}
	if code != BadCloudBaseSig {
		t.Fail()
	}
}

// VerifyGenesisBlock Tests

func TestVerifyGenesisBlock(t *testing.T) {
	miner := blockchain.NewWallet()
	currentTarget := blockchain.BigIntToHash(c.MaxTarget)
	currentBlockReward := uint64(25)
	gb := blockchain.Genesis(miner.Public(), currentTarget, currentBlockReward, []byte{})
	bc := &blockchain.BlockChain{
		Blocks: []*blockchain.Block{},
		Head:   blockchain.NilHash,
	}

	valid, code := VerifyGenesisBlock(bc, gb)

	if !valid {
		t.Fail()
	}

	if code != ValidGenesisBlock {
		t.Fail()
	}
}

func TestVerifyGenesisBlockNilGenesisBlock(t *testing.T) {

	bc := &blockchain.BlockChain{
		Blocks: []*blockchain.Block{nil},
		Head:   blockchain.NewTestHash(),
	}

	valid, code := VerifyGenesisBlock(bc, nil)

	if valid {
		t.Fail()
	}

	if code != NilGenesisBlock {
		t.Fail()
	}
}

func TestVerifyGenesisBlockBadGenesisBlockNumber(t *testing.T) {
	miner := blockchain.NewWallet()
	currentTarget := blockchain.BigIntToHash(c.MaxTarget)
	currentBlockReward := uint64(25)
	gb := blockchain.Genesis(miner.Public(), currentTarget, currentBlockReward, []byte{})
	gb.BlockHeader.BlockNumber = 1
	bc := &blockchain.BlockChain{
		Blocks: []*blockchain.Block{gb},
		Head:   blockchain.HashSum(gb),
	}

	valid, code := VerifyGenesisBlock(bc, gb)

	if valid {
		t.Fail()
	}

	if code != BadGenesisBlockNumber {
		t.Fail()
	}

	gb.BlockHeader.BlockNumber = 0
	bc = &blockchain.BlockChain{
		Blocks: []*blockchain.Block{blockchain.NewTestBlock(), gb},
		Head:   blockchain.HashSum(gb),
	}

	valid, code = VerifyGenesisBlock(bc, gb)

	if valid {
		t.Fail()
	}

	if code != BadGenesisBlockNumber {
		t.Fail()
	}
}

func TestVerifyGenesisBlockBadGenesisLastBlock(t *testing.T) {
	miner := blockchain.NewWallet()
	currentTarget := blockchain.BigIntToHash(c.MaxTarget)
	currentBlockReward := uint64(25)
	gb := blockchain.Genesis(miner.Public(), currentTarget, currentBlockReward, []byte{})
	gb.LastBlock = blockchain.NewTestHash()

	bc := &blockchain.BlockChain{
		Blocks: []*blockchain.Block{gb},
		Head:   blockchain.HashSum(gb),
	}

	valid, code := VerifyGenesisBlock(bc, gb)

	if valid {
		t.Fail()
	}

	if code != BadGenesisLastBlock {
		t.Fail()
	}
}

func TestVerifyGenesisBlockBadGenesisTransactions(t *testing.T) {
	miner := blockchain.NewWallet()
	currentTarget := blockchain.BigIntToHash(c.MaxTarget)
	currentBlockReward := uint64(25)
	gb := blockchain.Genesis(miner.Public(), currentTarget, currentBlockReward, []byte{})
	gb.Transactions = append(gb.Transactions, blockchain.NewTestTransaction())
	bc := &blockchain.BlockChain{
		Blocks: []*blockchain.Block{gb},
		Head:   blockchain.HashSum(gb),
	}

	valid, code := VerifyGenesisBlock(bc, gb)

	if valid {
		t.Fail()
	}

	if code != BadGenesisTransactions {
		t.Fail()
	}
}

func TestVerifyGenesisBlockBadGenesisCloudBaseTransaction(t *testing.T) {
	miner := blockchain.NewWallet()
	currentTarget := blockchain.BigIntToHash(c.MaxTarget)
	currentBlockReward := uint64(25)
	gb := blockchain.Genesis(miner.Public(), currentTarget, currentBlockReward, []byte{})
	gb.Transactions[0] = blockchain.NewTestTransaction()
	bc := &blockchain.BlockChain{
		Blocks: []*blockchain.Block{gb},
		Head:   blockchain.HashSum(gb),
	}

	valid, code := VerifyGenesisBlock(bc, gb)

	if valid {
		t.Fail()
	}

	if code != BadGenesisCloudBaseTransaction {
		t.Fail()
	}
}

func TestVerifyGenesisBlockBadGenesisTarget(t *testing.T) {
	miner := blockchain.NewWallet()
	currentTarget := blockchain.BigIntToHash(c.MaxTarget)
	currentBlockReward := uint64(25)
	gb := blockchain.Genesis(miner.Public(), currentTarget, currentBlockReward, []byte{})
	gb.Target = blockchain.BigIntToHash(util.BigExp(2, 255))
	bc := &blockchain.BlockChain{
		Blocks: []*blockchain.Block{gb},
		Head:   blockchain.HashSum(gb),
	}

	valid, code := VerifyGenesisBlock(bc, gb)

	if valid {
		t.Fail()
	}

	if code != BadGenesisTarget {
		t.Fail()
	}
}

func TestVerifyGenesisBlockBadGenesisTime(t *testing.T) {
	miner := blockchain.NewWallet()
	currentTarget := blockchain.BigIntToHash(c.MaxTarget)
	currentBlockReward := uint64(25)
	gb := blockchain.Genesis(miner.Public(), currentTarget, currentBlockReward, []byte{})
	gb.Time = 0
	bc := &blockchain.BlockChain{
		Blocks: []*blockchain.Block{gb},
		Head:   blockchain.HashSum(gb),
	}

	valid, code := VerifyGenesisBlock(bc, gb)

	if valid {
		t.Fail()
	}

	if code != BadGenesisTime {
		t.Fail()
	}
}

// Creates a random uint64 value
func RandomUint64() uint64 {
	return uint64(rand.Uint32())<<32 + uint64(rand.Uint32())
}
