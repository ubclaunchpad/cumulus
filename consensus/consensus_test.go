package consensus

import (
	"fmt"
	"math/rand"
	"testing"

	crand "crypto/rand"

	"github.com/ubclaunchpad/cumulus/blockchain"
	c "github.com/ubclaunchpad/cumulus/common/constants"
	"github.com/ubclaunchpad/cumulus/common/util"
)

// VerifyTransaction Tests

func TestVerifyTransactionNilTransaction(t *testing.T) {
	bc, _ := blockchain.NewValidBlockChainFixture()

	valid, code := VerifyTransaction(bc, nil)

	if valid {
		t.Fail()
	}
	if code != NilTransaction {
		t.Fail()
	}
}

func TestVerifyTransactionNoInputTransaction(t *testing.T) {
	tr, _ := blockchain.NewTestTransactionValue(
		blockchain.NewWallet(),
		blockchain.NewWallet(),
		1,
		0,
	)
	bc, _ := blockchain.NewValidBlockChainFixture()

	valid, code := VerifyTransaction(bc, tr)

	if valid {
		t.Fail()
	}
	if code != NoInputTransaction {
		t.Fail()
	}
}

func TestVerifyTransactionOverspend(t *testing.T) {
	// 2 + 2 = 5 ?
	bc, _ := blockchain.NewValidBlockChainFixture()
	tr := bc.Blocks[1].Transactions[1]
	tr.Outputs[0].Amount = 5

	valid, code := VerifyTransaction(bc, tr)

	if valid {
		t.Fail()
	}
	if code != Overspend {
		fmt.Println(code)
		t.Fail()
	}
}

func TestVerifyTransactionSignatureFail(t *testing.T) {
	bc, _ := blockchain.NewValidBlockChainFixture()
	tr := bc.Blocks[1].Transactions[1]

	fakeSender := blockchain.NewWallet()
	tr, _ = tr.TxBody.Sign(*fakeSender, crand.Reader)
	bc.Blocks[1].Transactions[1] = tr

	valid, code := VerifyTransaction(bc, tr)
	if valid {
		t.Fail()
	}
	if code != BadSig {
		t.Fail()
	}
}

func TestVerifyTransactionPass(t *testing.T) {
	bc, b := blockchain.NewValidTestChainAndBlock()
	tr := b.Transactions[1]

	valid, code := VerifyTransaction(bc, tr)

	if !valid {
		t.Fail()
	}
	if code != ValidTransaction {
		t.Fail()
	}
}

func TestTransactionRespend(t *testing.T) {
	bc, _ := blockchain.NewValidBlockChainFixture()
	trC := bc.Blocks[1].Transactions[1]
	b := blockchain.NewTestOutputBlock([]*blockchain.Transaction{trC}, bc.Blocks[1])
	bc.AppendBlock(b)

	valid, code := VerifyTransaction(bc, trC)

	if valid {
		t.Fail()
	}
	if code != Respend {
		t.Fail()
	}
}

// VerifyBlock Tests

func TestVerifyBlockNilBlock(t *testing.T) {
	bc, _ := blockchain.NewValidBlockChainFixture()

	valid, code := VerifyBlock(bc, nil)

	if valid {
		t.Fail()
	}
	if code != NilBlock {
		t.Fail()
	}
}

func TestVerifyBlockBadNonce(t *testing.T) {
	bc, b := blockchain.NewValidTestChainAndBlock()
	b.Target = blockchain.BigIntToHash(c.Big1)
	CurrentDifficulty = c.MaxTarget
	valid, code := VerifyBlock(bc, b)
	CurrentDifficulty = c.MinTarget

	if valid {
		t.Fail()
	}

	if code != BadNonce {
		t.Fail()
	}
}

func TestVerifyBlockBadGenesisBlock(t *testing.T) {
	miner := blockchain.NewWallet()
	currentTarget := blockchain.BigIntToHash(c.MaxTarget)
	currentBlockReward := uint64(25)
	gb := blockchain.Genesis(miner.Public(), currentTarget, currentBlockReward, []byte{})
	gb.Target = blockchain.BigIntToHash(util.BigExp(2, 255))
	bc := &blockchain.BlockChain{
		Blocks: []*blockchain.Block{gb},
		Head:   blockchain.HashSum(gb),
	}

	valid, code := VerifyBlock(bc, gb)

	if valid {
		t.Fail()
	}

	if code != BadGenesisBlock {
		t.Fail()
	}
}

func TestVerifyBlockBadTransaction(t *testing.T) {
	bc, _ := blockchain.NewValidBlockChainFixture()
	tr := bc.Blocks[1].Transactions[1]
	tr.Outputs[0].Amount = 5

	valid, code := VerifyBlock(bc, bc.Blocks[1])

	if valid {
		t.Fail()
	}
	if code != BadTransaction {
		t.Fail()
	}
}

func TestVerifyBlockBadBlockNumber(t *testing.T) {
	bc, _ := blockchain.NewValidBlockChainFixture()
	bc.Blocks[1].BlockNumber = 2

	valid, code := VerifyBlock(bc, bc.Blocks[1])

	if valid {
		t.Fail()
	}
	if code != BadBlockNumber {
		t.Fail()
	}
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

	if valid {
		t.Fail()
	}
	if code != DoubleSpend {
		t.Fail()
	}
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
	b.GetCloudBaseTransaction().TxBody.Input.BlockNumber = 1
	valid, code := VerifyCloudBase(bc, b.GetCloudBaseTransaction())

	if valid {
		t.Fail()
	}
	if code != BadCloudBaseInput {
		t.Fail()
	}

	bc, _ = blockchain.NewValidBlockChainFixture()
	b = bc.Blocks[0]
	b.GetCloudBaseTransaction().TxBody.Input.Hash = blockchain.NewTestHash()
	valid, code = VerifyCloudBase(bc, b.GetCloudBaseTransaction())

	if valid {
		t.Fail()
	}
	if code != BadCloudBaseInput {
		t.Fail()
	}

	bc, _ = blockchain.NewValidBlockChainFixture()
	b = bc.Blocks[0]
	b.GetCloudBaseTransaction().TxBody.Input.Index = 1
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
		Blocks: []*blockchain.Block{gb},
		Head:   blockchain.HashSum(gb),
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
