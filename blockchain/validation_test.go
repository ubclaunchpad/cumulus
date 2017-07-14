package blockchain

import (
	"crypto/rand"
	"fmt"
	"testing"

	c "github.com/ubclaunchpad/cumulus/common/constants"
	"github.com/ubclaunchpad/cumulus/common/util"
)

func TestValidTransactionNilTransaction(t *testing.T) {
	bc, _ := NewValidBlockChainFixture()

	valid, code := bc.ValidTransaction(nil)

	if valid {
		t.Fail()
	}
	if code != NilTransaction {
		t.Fail()
	}
}

func TestValidTransactionNoInputTransaction(t *testing.T) {
	tr, _ := NewTestTransactionValue(NewWallet(), NewWallet(), 1, 0)
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
	tr := bc.Blocks[1].Transactions[1]
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
	tr := bc.Blocks[1].Transactions[1]

	fakeSender := NewWallet()
	tr, _ = tr.TxBody.Sign(fakeSender, rand.Reader)
	bc.Blocks[1].Transactions[1] = tr

	valid, code := bc.ValidTransaction(tr)
	if valid {
		t.Fail()
	}
	if code != BadSig {
		t.Fail()
	}
}

func TestValidTransactionPass(t *testing.T) {
	bc, b := NewValidTestChainAndBlock()
	tr := b.Transactions[1]

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
	trC := bc.Blocks[1].Transactions[1]
	b := NewTestOutputBlock([]*Transaction{trC}, bc.Blocks[1])
	bc.AppendBlock(b)

	valid, code := bc.ValidTransaction(trC)

	if valid {
		t.Fail()
	}
	if code != Respend {
		t.Fail()
	}
}

func TestValidBlockNilBlock(t *testing.T) {
	bc, _ := NewValidBlockChainFixture()

	valid, code := bc.ValidBlock(nil)

	if valid {
		t.Fail()
	}
	if code != NilBlock {
		t.Fail()
	}
}

func TestValidBlockBadGenesisBlock(t *testing.T) {
	miner := NewWallet()
	currentTarget := BigIntToHash(MaxTarget)
	currentBlockReward := uint64(25)
	gb := Genesis(miner.Public(), currentTarget, currentBlockReward, []byte{})
	gb.Target = BigIntToHash(util.BigExp(2, 255))
	bc := &BlockChain{
		Blocks: []*Block{gb},
		Head:   HashSum(gb),
	}

	valid, code := bc.ValidBlock(gb)

	if valid {
		t.Fail()
	}

	if code != BadGenesisBlock {
		t.Fail()
	}
}

func TestValidBlockBadTransaction(t *testing.T) {
	bc, _ := NewValidBlockChainFixture()
	tr := bc.Blocks[1].Transactions[1]
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
	bc, b := NewValidTestChainAndBlock()
	b.BlockHeader.LastBlock = NewTestHash()

	valid, code := bc.ValidBlock(b)

	if valid {
		t.Fail()
	}
	if code != BadHash {
		t.Fail()
	}
}

func TestValidBlockBadTime(t *testing.T) {
	bc, b := NewValidTestChainAndBlock()
	b.Time = 0
	valid, code := bc.ValidBlock(b)

	if valid {
		t.Fail()
	}
	if code != BadTime {
		t.Fail()
	}
}

func TestValidBlockBadTarget(t *testing.T) {
	bc, b := NewValidTestChainAndBlock()
	b.Target = BigIntToHash(util.BigAdd(MaxTarget, c.Big1))
	valid, code := bc.ValidBlock(b)

	if valid {
		t.Fail()
	}
	if code != BadTarget {
		t.Fail()
	}

	b.Target = BigIntToHash(util.BigSub(MinTarget, c.Big1))
	valid, code = bc.ValidBlock(b)

	if valid {
		t.Fail()
	}
	if code != BadTarget {
		t.Fail()
	}
}

func TestValidBlock(t *testing.T) {
	bc, b := NewValidTestChainAndBlock()

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
	bc, b := NewValidTestChainAndBlock()
	b.Transactions = append(b.Transactions, b.Transactions[1])

	valid, code := bc.ValidBlock(b)

	if valid {
		t.Fail()
	}
	if code != DoubleSpend {
		t.Fail()
	}
}

func TestValidBlockBigNumber(t *testing.T) {
	bc, b := NewValidTestChainAndBlock()
	b.BlockNumber = uint32(len(bc.Blocks)) + 1

	valid, code := bc.ValidBlock(b)

	if valid {
		t.Fail()
	}
	if code != BadBlockNumber {
		t.Fail()
	}
}

func TestValidBlockBadCloudBaseTransaction(t *testing.T) {
	bc, b := NewValidTestChainAndBlock()
	b.Transactions[0] = NewTestTransaction()

	valid, code := bc.ValidBlock(b)

	if valid {
		t.Fail()
	}
	if code != BadCloudBaseTransaction {
		t.Fail()
	}
}

func TestValidCloudBaseNilCloudBase(t *testing.T) {
	valid, code := ValidCloudBase(nil)

	if valid {
		t.Fail()
	}
	if code != NilCloudBaseTransaction {
		t.Fail()
	}
}

func TestValidCloudBaseTransaction(t *testing.T) {
	cbTx, _ := NewValidCloudBaseTestTransaction()
	valid, code := ValidCloudBase(cbTx)

	if !valid {
		t.Fail()
	}
	if code != ValidCloudBaseTransaction {
		t.Fail()
	}
}

func TestValidCloudBaseBadSender(t *testing.T) {
	w := NewWallet()
	cbTx, _ := NewValidCloudBaseTestTransaction()
	cbTx.Sender = w.Public()
	valid, code := ValidCloudBase(cbTx)

	if valid {
		t.Fail()
	}
	if code != BadCloudBaseSender {
		t.Fail()
	}
}

func TestValidCloudBaseBadBadInput(t *testing.T) {
	cbTx, _ := NewValidCloudBaseTestTransaction()
	cbTx.TxBody.Input.BlockNumber = 1
	valid, code := ValidCloudBase(cbTx)

	if valid {
		t.Fail()
	}
	if code != BadCloudBaseInput {
		t.Fail()
	}

	cbTx, _ = NewValidCloudBaseTestTransaction()
	cbTx.TxBody.Input.Hash = NewTestHash()
	valid, code = ValidCloudBase(cbTx)

	if valid {
		t.Fail()
	}
	if code != BadCloudBaseInput {
		t.Fail()
	}

	cbTx, _ = NewValidCloudBaseTestTransaction()
	cbTx.TxBody.Input.Index = 1
	valid, code = ValidCloudBase(cbTx)

	if valid {
		t.Fail()
	}
	if code != BadCloudBaseInput {
		t.Fail()
	}
}

func TestValidCloudBaseBadOutput(t *testing.T) {
	cbTx, _ := NewValidCloudBaseTestTransaction()
	w := NewWallet()
	cbTx.Outputs = append(cbTx.Outputs, TxOutput{25, w.Public()})
	valid, code := ValidCloudBase(cbTx)

	if valid {
		t.Fail()
	}
	if code != BadCloudBaseOutput {
		t.Fail()
	}

	cbTx, _ = NewValidCloudBaseTestTransaction()
	var emptyOutputs []TxOutput
	cbTx.Outputs = emptyOutputs
	valid, code = ValidCloudBase(cbTx)

	if valid {
		t.Fail()
	}
	if code != BadCloudBaseOutput {
		t.Fail()
	}

	cbTx, _ = NewValidCloudBaseTestTransaction()
	cbTx.Outputs[0].Amount = 0
	valid, code = ValidCloudBase(cbTx)

	if valid {
		t.Fail()
	}
	if code != BadCloudBaseOutput {
		t.Fail()
	}

	cbTx, _ = NewValidCloudBaseTestTransaction()
	cbTx.Outputs[0].Recipient = NilAddr
	valid, code = ValidCloudBase(cbTx)

	if valid {
		t.Fail()
	}
	if code != BadCloudBaseOutput {
		t.Fail()
	}
}

func TestValidCloudBaseBadSig(t *testing.T) {
	cbTx, _ := NewValidCloudBaseTestTransaction()
	w := NewWallet()
	cbTx.Sig, _ = w.Sign(NewTestHash(), rand.Reader)
	valid, code := ValidCloudBase(cbTx)

	if valid {
		t.Fail()
	}
	if code != BadCloudBaseSig {
		t.Fail()
	}
}

func TestValidGenesisBlock(t *testing.T) {
	miner := NewWallet()
	currentTarget := BigIntToHash(MaxTarget)
	currentBlockReward := uint64(25)
	gb := Genesis(miner.Public(), currentTarget, currentBlockReward, []byte{})
	bc := &BlockChain{
		Blocks: []*Block{gb},
		Head:   HashSum(gb),
	}

	valid, code := bc.ValidGenesisBlock(gb)

	if !valid {
		t.Fail()
	}

	if code != ValidGenesisBlock {
		t.Fail()
	}
}

func TestValidGenesisBlockNilGenesisBlock(t *testing.T) {

	bc := &BlockChain{
		Blocks: []*Block{nil},
		Head:   NewTestHash(),
	}

	valid, code := bc.ValidGenesisBlock(nil)

	if valid {
		t.Fail()
	}

	if code != NilGenesisBlock {
		t.Fail()
	}
}

func TestValidGenesisBlockBadGenesisBlockNumber(t *testing.T) {
	miner := NewWallet()
	currentTarget := BigIntToHash(MaxTarget)
	currentBlockReward := uint64(25)
	gb := Genesis(miner.Public(), currentTarget, currentBlockReward, []byte{})
	gb.BlockHeader.BlockNumber = 1
	bc := &BlockChain{
		Blocks: []*Block{gb},
		Head:   HashSum(gb),
	}

	valid, code := bc.ValidGenesisBlock(gb)

	if valid {
		t.Fail()
	}

	if code != BadGenesisBlockNumber {
		t.Fail()
	}

	gb.BlockHeader.BlockNumber = 0
	bc = &BlockChain{
		Blocks: []*Block{NewTestBlock(), gb},
		Head:   HashSum(gb),
	}

	valid, code = bc.ValidGenesisBlock(gb)

	if valid {
		t.Fail()
	}

	if code != BadGenesisBlockNumber {
		t.Fail()
	}
}

func TestValidGenesisBlockBadGenesisLastBlock(t *testing.T) {
	miner := NewWallet()
	currentTarget := BigIntToHash(MaxTarget)
	currentBlockReward := uint64(25)
	gb := Genesis(miner.Public(), currentTarget, currentBlockReward, []byte{})
	gb.LastBlock = NewTestHash()

	bc := &BlockChain{
		Blocks: []*Block{gb},
		Head:   HashSum(gb),
	}

	valid, code := bc.ValidGenesisBlock(gb)

	if valid {
		t.Fail()
	}

	if code != BadGenesisLastBlock {
		t.Fail()
	}
}

func TestValidGenesisBlockBadGenesisTransactions(t *testing.T) {
	miner := NewWallet()
	currentTarget := BigIntToHash(MaxTarget)
	currentBlockReward := uint64(25)
	gb := Genesis(miner.Public(), currentTarget, currentBlockReward, []byte{})
	gb.Transactions = append(gb.Transactions, NewTestTransaction())
	bc := &BlockChain{
		Blocks: []*Block{gb},
		Head:   HashSum(gb),
	}

	valid, code := bc.ValidGenesisBlock(gb)

	if valid {
		t.Fail()
	}

	if code != BadGenesisTransactions {
		t.Fail()
	}
}

func TestValidGenesisBlockBadGenesisCloudBaseTransaction(t *testing.T) {
	miner := NewWallet()
	currentTarget := BigIntToHash(MaxTarget)
	currentBlockReward := uint64(25)
	gb := Genesis(miner.Public(), currentTarget, currentBlockReward, []byte{})
	gb.Transactions[0] = NewTestTransaction()
	bc := &BlockChain{
		Blocks: []*Block{gb},
		Head:   HashSum(gb),
	}

	valid, code := bc.ValidGenesisBlock(gb)

	if valid {
		t.Fail()
	}

	if code != BadGenesisCloudBaseTransaction {
		t.Fail()
	}
}

func TestValidGenesisBlockBadGenesisTarget(t *testing.T) {
	miner := NewWallet()
	currentTarget := BigIntToHash(MaxTarget)
	currentBlockReward := uint64(25)
	gb := Genesis(miner.Public(), currentTarget, currentBlockReward, []byte{})
	gb.Target = BigIntToHash(util.BigExp(2, 255))
	bc := &BlockChain{
		Blocks: []*Block{gb},
		Head:   HashSum(gb),
	}

	valid, code := bc.ValidGenesisBlock(gb)

	if valid {
		t.Fail()
	}

	if code != BadGenesisTarget {
		t.Fail()
	}
}

func TestValidGenesisBlockBadGenesisTime(t *testing.T) {
	miner := NewWallet()
	currentTarget := BigIntToHash(MaxTarget)
	currentBlockReward := uint64(25)
	gb := Genesis(miner.Public(), currentTarget, currentBlockReward, []byte{})
	gb.Time = 0
	bc := &BlockChain{
		Blocks: []*Block{gb},
		Head:   HashSum(gb),
	}

	valid, code := bc.ValidGenesisBlock(gb)

	if valid {
		t.Fail()
	}

	if code != BadGenesisTime {
		t.Fail()
	}
}
