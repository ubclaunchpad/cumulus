package blockchain

import (
	crand "crypto/rand"
	"crypto/sha256"
	"math/big"
	mrand "math/rand"

	c "github.com/ubclaunchpad/cumulus/common/constants"
	"github.com/ubclaunchpad/cumulus/common/util"
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
		Recipient: NewWallet().Public().Repr(),
	}
}

// NewTestTxBody random txn body.
func NewTestTxBody() TxBody {
	// Uniform distribution on [1, 4]
	nOutputs := mrand.Intn(4) + 1
	nInputs := mrand.Intn(4) + 1
	body := TxBody{
		Sender:  NewWallet().Public(),
		Inputs:  make([]TxHashPointer, nInputs),
		Outputs: make([]TxOutput, nOutputs),
	}
	for i := 0; i < nOutputs; i++ {
		body.Outputs[i] = NewTestTxOutput()
	}
	for i := 0; i < nInputs; i++ {
		body.Inputs[i] = NewTestTxHashPointer()
	}
	return body
}

// NewTestTransaction prodcues random txn.
func NewTestTransaction() *Transaction {
	sender := NewWallet()
	tbody := NewTestTxBody()
	t, _ := tbody.Sign(*sender, crand.Reader)
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
			Time:        util.UnixNow(),
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
			Time:        util.UnixNow(),
			Nonce:       0,
		},
		Transactions: t,
	}
}

// NewTestTransactionValue creates a new transaction with specific value a at
// index i in block number b.
func NewTestTransactionValue(s, r *Wallet, a uint64, i uint32, b uint32) (*Transaction, error) {
	tbody := TxBody{
		Sender:  s.Public(),
		Inputs:  make([]TxHashPointer, 1),
		Outputs: make([]TxOutput, 1),
	}
	tbody.Outputs[0] = TxOutput{
		Amount:    a,
		Recipient: r.Public().Repr(),
	}
	tbody.Inputs[0] = TxHashPointer{
		BlockNumber: b,
		Hash:        NewTestHash(),
		Index:       i,
	}
	return tbody.Sign(*s, crand.Reader)
}

// NewValidBlockChainFixture creates a valid blockchain of two blocks
// and returns the recipient of the only transaction in block 1.
func NewValidBlockChainFixture() (*BlockChain, Wallet) {
	original := NewWallet()
	sender := NewWallet()
	recipient := NewWallet()

	// Transaction A is in block 0 at index 0 (sender awarded 2 coins).
	trA, _ := NewTestTransactionValue(original, sender, 2, 1, 0)
	trA.Outputs = append(trA.Outputs, TxOutput{
		Amount:    2,
		Recipient: sender.Public().Repr(),
	})

	// Transaction B is in block 1 at index 0 (sender sends 2 coins to recipient).
	trB, _ := NewTestTransactionValue(sender, recipient, 2, 1, 1)
	trB.Inputs[1].Hash = HashSum(trA)

	trB, _ = trB.TxBody.Sign(*sender, crand.Reader)

	// CloudBase will bump our transactions forward.
	cbA, _ := NewValidCloudBaseTestTransaction()
	cbB, _ := NewValidCloudBaseTestTransaction()
	inputTransactions := []*Transaction{cbA, trA}
	outputTransactions := []*Transaction{cbB, trB}

	// Create the blocks in the blockchain.
	inputBlock := NewTestInputBlock(inputTransactions)
	outputBlock := NewTestOutputBlock(outputTransactions, inputBlock)

	return &BlockChain{
		Blocks: []*Block{inputBlock, outputBlock},
		Head:   NewTestHash(),
	}, *recipient
}

// NewValidTestChainAndBlock creates a valid BlockChain of 2 blocks,
// and a Block that is valid with respect to the BlockChain.
//	 [ 2in | 2out ]   ->>   [ 2in | 2out ]	 ,	[ 2in | 2out ]
func NewValidTestChainAndBlock() (*BlockChain, *Block) {
	bc, s := NewValidBlockChainFixture()
	inputBlock := bc.Blocks[1]

	// Collect the transaction following the CloudBase.
	inputTransaction := inputBlock.Transactions[1]
	a := inputTransaction.Outputs[1].Amount

	// Create a legit block that does *not* appear in bc.
	tbody := TxBody{
		Sender:  s.Public(),
		Inputs:  make([]TxHashPointer, 1),
		Outputs: make([]TxOutput, 1),
	}
	tbody.Inputs[0] = TxHashPointer{
		BlockNumber: 1,
		Hash:        HashSum(inputTransaction),
		Index:       1,
	}
	tbody.Outputs[0] = TxOutput{
		Amount:    a,
		Recipient: NewWallet().Public().Repr(),
	}

	tr, _ := tbody.Sign(s, crand.Reader)
	cb, _ := NewValidCloudBaseTestTransaction()
	newBlock := NewTestOutputBlock([]*Transaction{cb, tr}, inputBlock)
	return bc, newBlock
}

// NewValidTestTarget creates a new valid target that is a random value between the
// max and min difficulties
func NewValidTestTarget() Hash {
	return BigIntToHash(
		new(big.Int).Div(
			c.MaxTarget,
			c.MinTarget,
		),
	)
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
		Recipient: w.Public().Repr(),
	}
	cbTxBody := TxBody{
		Sender:  NilAddr,
		Inputs:  []TxHashPointer{cbInput},
		Outputs: []TxOutput{cbReward},
	}
	cbTx := &Transaction{
		TxBody: cbTxBody,
		Sig:    NilSig,
	}
	return cbTx, w.Public()
}
