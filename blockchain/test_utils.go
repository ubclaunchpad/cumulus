package blockchain

import (
	crand "crypto/rand"
	"crypto/sha256"
	"math/big"
	mrand "math/rand"
	"time"

	"github.com/ubclaunchpad/cumulus/common"
	c "github.com/ubclaunchpad/cumulus/common/constants"
	"github.com/ubclaunchpad/cumulus/common/math"
)

// NewTestHash produces a hash.
func NewTestHash() Hash {
	message := make([]byte, 256)
	crand.Read(message)
	return sha256.Sum256(message)
}

// NewTestTxHashPointer produces transaction hash pointer.
func NewTestTxHashPointer() TxHashPointer {
	return TxHashPointer{
		BlockNumber: mrand.Uint32(),
		Hash:        NewTestHash(),
		Index:       mrand.Uint32(),
	}
}

// NewTestTxOutput random txn output.
func NewTestTxOutput() TxOutput {
	return TxOutput{
		Amount:    uint64(mrand.Int63()),
		Recipient: NewWallet().Public(),
	}
}

// NewTestTxBody random txn body.
func NewTestTxBody() TxBody {
	// Uniform distribution on [1, 4]
	nOutputs := mrand.Intn(4) + 1
	body := TxBody{
		Sender:  NewWallet().Public(),
		Input:   NewTestTxHashPointer(),
		Outputs: make([]TxOutput, nOutputs),
	}
	for i := 0; i < nOutputs; i++ {
		body.Outputs[i] = NewTestTxOutput()
	}
	return body
}

// NewTestTransaction prodcues random txn.
func NewTestTransaction() *Transaction {
	sender := NewWallet()
	tbody := NewTestTxBody()
	t, _ := tbody.Sign(sender, crand.Reader)
	return t
}

// NewTestBlockHeader prodcues random block header.
func NewTestBlockHeader() BlockHeader {
	return BlockHeader{
		BlockNumber: mrand.Uint32(),
		LastBlock:   NewTestHash(),
		Target:      NewValidTestTarget(),
		Time:        mrand.Uint32(),
		Nonce:       0,
	}
}

// NewTestBlock prodcues random block.
func NewTestBlock() *Block {
	// Uniform distribution on [500, 999]
	nTransactions := mrand.Intn(500) + 500
	b := Block{
		BlockHeader:  NewTestBlockHeader(),
		Transactions: make([]*Transaction, nTransactions),
	}
	for i := 0; i < nTransactions; i++ {
		b.Transactions[i] = NewTestTransaction()
	}
	return &b
}

// NewTestBlockChain produces random blockchain.
func NewTestBlockChain() *BlockChain {
	// Uniform distribution on [10, 50]
	nBlocks := mrand.Intn(40) + 10
	bc := BlockChain{Blocks: make([]*Block, nBlocks)}
	for i := 0; i < nBlocks; i++ {
		bc.Blocks[i] = NewTestBlock()
	}
	bc.Head = HashSum(bc.Blocks[nBlocks-1])
	return &bc
}

// NewTestInputBlock produces new block with given transactions.
func NewTestInputBlock(t []*Transaction) *Block {
	return &Block{
		BlockHeader: BlockHeader{
			BlockNumber: 0,
			LastBlock:   NewTestHash(),
			Target:      NewValidTestTarget(),
			Time:        common.UnixNow(),
			Nonce:       0,
		},
		Transactions: t,
	}
}

// NewTestOutputBlock produces new block with given transactions and given input
// block.
func NewTestOutputBlock(t []*Transaction, input *Block) *Block {
	return &Block{
		BlockHeader: BlockHeader{
			BlockNumber: input.BlockNumber + 1,
			LastBlock:   HashSum(input),
			Target:      NewValidTestTarget(),
			Time:        common.UnixNow(),
			Nonce:       0,
		},
		Transactions: t,
	}
}

// NewTestTransactionValue creates a new transaction with specific value a.
func NewTestTransactionValue(s, r Wallet, a uint64, i uint32) (*Transaction, error) {
	tbody := TxBody{
		Sender: s.Public(),
		Input: TxHashPointer{
			BlockNumber: 0,
			Hash:        NewTestHash(),
			Index:       i,
		},
		Outputs: make([]TxOutput, 1),
	}
	tbody.Outputs[0] = TxOutput{
		Amount:    a,
		Recipient: r.Public(),
	}
	return tbody.Sign(s, crand.Reader)
}

// NewValidBlockChainFixture creates a valid blockchain of two blocks.
func NewValidBlockChainFixture() (*BlockChain, Wallet) {
	original := NewWallet()
	sender := NewWallet()
	recipient := NewWallet()

	trA, _ := NewTestTransactionValue(original, sender, 2, 1)
	trA.Outputs = append(trA.Outputs, TxOutput{
		Amount:    2,
		Recipient: sender.Public(),
	})

	trB, _ := NewTestTransactionValue(sender, recipient, 4, 1)
	trB.Input.Hash = HashSum(trA)

	trB, _ = trB.TxBody.Sign(sender, crand.Reader)

	cbA, _ := NewValidCloudBaseTestTransaction()
	cbB, _ := NewValidCloudBaseTestTransaction()
	inputTransactions := []*Transaction{cbA, trA}
	outputTransactions := []*Transaction{cbB, trB}

	inputBlock := NewTestInputBlock(inputTransactions)
	outputBlock := NewTestOutputBlock(outputTransactions, inputBlock)

	return &BlockChain{
		Blocks: []*Block{inputBlock, outputBlock},
		Head:   NewTestHash(),
	}, recipient
}

// NewValidTestChainAndBlock creates a valid BlockChain and a Block that is valid
// with respect to the BlockChain.
func NewValidTestChainAndBlock() (*BlockChain, *Block) {
	bc, s := NewValidBlockChainFixture()
	inputBlock := bc.Blocks[1]
	inputTransaction := inputBlock.Transactions[0]
	a := inputTransaction.Outputs[0].Amount

	// Create a legit block that does *not* appear in bc.
	tbody := TxBody{
		Sender: s.Public(),
		Input: TxHashPointer{
			BlockNumber: 1,
			Hash:        HashSum(inputTransaction),
			Index:       0,
		},
		Outputs: make([]TxOutput, 1),
	}
	tbody.Outputs[0] = TxOutput{
		Amount:    a,
		Recipient: NewWallet().Public(),
	}

	tr, _ := tbody.Sign(s, crand.Reader)
	cb, _ := NewValidCloudBaseTestTransaction()
	newBlock := NewTestOutputBlock([]*Transaction{cb, tr}, inputBlock)
	return bc, newBlock
}

// NewValidTestTarget creates a new valid target that is a random value between the
// max and min difficulties
func NewValidTestTarget() Hash {
	r := new(big.Int).Rand(
		mrand.New(mrand.NewSource(time.Now().Unix())),
		math.BigAdd(MaxTarget, c.Big1),
	)
	r.Add(r, c.Big1)
	return BigIntToHash(r)
}

// NewValidCloudBaseTestTransaction returns a new valid CloudBase transaction and
// the address of the recipient of the transaction
func NewValidCloudBaseTestTransaction() (*Transaction, Address) {
	w := NewWallet()
	cbInput := TxHashPointer{
		BlockNumber: 0,
		Hash:        NilHash,
		Index:       0,
	}
	cbReward := TxOutput{
		Amount:    25,
		Recipient: w.Public(),
	}
	cbTxBody := TxBody{
		Sender:  NilAddr,
		Input:   cbInput,
		Outputs: []TxOutput{cbReward},
	}
	cbTx := &Transaction{
		TxBody: cbTxBody,
		Sig:    NilSig,
	}
	return cbTx, w.Public()
}
