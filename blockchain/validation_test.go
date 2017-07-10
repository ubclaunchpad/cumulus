package blockchain

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"testing"
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
	bc, b := NewValidChainAndBlock()
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

func TestValidBlockBadBlockSize(t *testing.T) {
	bc, b := NewValidChainAndBlock()

	for i := 0; i <= MaxBlockSize/TxOutputLen; i++ {
		b.Transactions[0].Outputs = append(
			b.Transactions[0].Outputs,
			TxOutput{
				0,
				NewWallet().Public(),
			},
		)
	}

	valid, code := bc.ValidBlock(b)

	if valid {
		t.Fail()
	}
	if code != BadBlockSize {
		t.Fail()
	}
}

func TestValidBlockBadGenesisBlock(t *testing.T) {
	miner := NewWallet()
	currentTarget := BigIntToHash(MaxTarget)
	currentBlockReward := uint64(25)
	gb := Genesis(miner.Public(), currentTarget, currentBlockReward, []byte{})
	gb.Target = BigIntToHash(BigExp(2, 255))
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
}

func TestValidBlockBadTarget(t *testing.T) {
	bc, b := NewValidChainAndBlock()
	b.Target = BigIntToHash(new(big.Int).Add(MaxTarget, big.NewInt(1)))
	valid, code := bc.ValidBlock(b)

	if valid {
		t.Fail()
	}
	if code != BadTarget {
		t.Fail()
	}

	b.Target = BigIntToHash(new(big.Int).Sub(MinTarget, big.NewInt(1)))
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

func TestValidBlockBadCloudBaseTransaction(t *testing.T) {
	bc, b := NewValidChainAndBlock()
	b.Transactions[0] = NewTransaction()

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
	cbTx, _ := NewValidCloudBaseTransaction()
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
	cbTx, _ := NewValidCloudBaseTransaction()
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
	cbTx, _ := NewValidCloudBaseTransaction()
	cbTx.TxBody.Input.BlockNumber = 1
	valid, code := ValidCloudBase(cbTx)

	if valid {
		t.Fail()
	}
	if code != BadCloudBaseInput {
		t.Fail()
	}

	cbTx, _ = NewValidCloudBaseTransaction()
	cbTx.TxBody.Input.Hash = NewHash()
	valid, code = ValidCloudBase(cbTx)

	if valid {
		t.Fail()
	}
	if code != BadCloudBaseInput {
		t.Fail()
	}

	cbTx, _ = NewValidCloudBaseTransaction()
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
	cbTx, _ := NewValidCloudBaseTransaction()
	w := NewWallet()
	cbTx.Outputs = append(cbTx.Outputs, TxOutput{25, w.Public()})
	valid, code := ValidCloudBase(cbTx)

	if valid {
		t.Fail()
	}
	if code != BadCloudBaseOutput {
		t.Fail()
	}

	cbTx, _ = NewValidCloudBaseTransaction()
	var emptyOutputs []TxOutput
	cbTx.Outputs = emptyOutputs
	valid, code = ValidCloudBase(cbTx)

	if valid {
		t.Fail()
	}
	if code != BadCloudBaseOutput {
		t.Fail()
	}

	cbTx, _ = NewValidCloudBaseTransaction()
	cbTx.Outputs[0].Amount = 0
	valid, code = ValidCloudBase(cbTx)

	if valid {
		t.Fail()
	}
	if code != BadCloudBaseOutput {
		t.Fail()
	}

	cbTx, _ = NewValidCloudBaseTransaction()
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
	cbTx, _ := NewValidCloudBaseTransaction()
	w := NewWallet()
	cbTx.Sig, _ = w.Sign(NewHash(), rand.Reader)
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
		Head:   NewHash(),
	}

	valid, code := bc.ValidGenesisBlock(nil)

	if valid {
		t.Fail()
	}

	if code != NilGenesisBlock {
		t.Fail()
	}
}

func TestValidGenesisBlockBadGenesisBlockSize(t *testing.T) {
	miner := NewWallet()
	currentTarget := BigIntToHash(MaxTarget)
	currentBlockReward := uint64(25)
	gb := Genesis(miner.Public(), currentTarget, currentBlockReward, []byte{})

	for i := 0; i <= MaxBlockSize/TxOutputLen; i++ {
		gb.Transactions[0].Outputs = append(
			gb.Transactions[0].Outputs,
			TxOutput{
				0,
				NewWallet().Public(),
			},
		)
	}

	bc := &BlockChain{
		Blocks: []*Block{gb},
		Head:   HashSum(gb),
	}

	valid, code := bc.ValidGenesisBlock(gb)

	if valid {
		t.Fail()
	}

	if code != BadGenesisBlockSize {
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
		Blocks: []*Block{NewBlock(), gb},
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
	gb.LastBlock = NewHash()
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
	gb.Transactions = append(gb.Transactions, NewTransaction())
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
	gb.Transactions[0] = NewTransaction()
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
	gb.Target = BigIntToHash(BigExp(2, 255))
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
