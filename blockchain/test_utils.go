package blockchain

import (
	crand "crypto/rand"
	"crypto/sha256"
	"math/big"
	mrand "math/rand"
	"time"
)

// NewHash produces a hash.
func NewHash() Hash {
	message := make([]byte, 256)
	crand.Read(message)
	return sha256.Sum256(message)
}

// NewTxHashPointer produces transaction hash pointer.
func NewTxHashPointer() TxHashPointer {
	return TxHashPointer{
		BlockNumber: mrand.Uint32(),
		Hash:        NewHash(),
		Index:       mrand.Uint32(),
	}
}

// NewTxOutput random txn output.
func NewTxOutput() TxOutput {
	return TxOutput{
		Amount:    uint64(mrand.Int63()),
		Recipient: NewWallet().Public(),
	}
}

// NewTxBody random txn body.
func NewTxBody() TxBody {
	// Uniform distribution on [1, 4]
	nOutputs := mrand.Intn(4) + 1
	body := TxBody{
		Sender:  NewWallet().Public(),
		Input:   NewTxHashPointer(),
		Outputs: make([]TxOutput, nOutputs),
	}
	for i := 0; i < nOutputs; i++ {
		body.Outputs[i] = NewTxOutput()
	}
	return body
}

// NewTransaction prodcues random txn.
func NewTransaction() *Transaction {
	sender := NewWallet()
	tbody := NewTxBody()
	t, _ := tbody.Sign(sender, crand.Reader)
	return t
}

// NewBlockHeader prodcues random block header.
func NewBlockHeader() BlockHeader {
	return BlockHeader{
		BlockNumber: mrand.Uint32(),
		LastBlock:   NewHash(),
	}
}

// NewBlock prodcues random block.
func NewBlock() *Block {
	// Uniform distribution on [500, 999]
	nTransactions := mrand.Intn(500) + 500
	b := Block{
		BlockHeader:  NewBlockHeader(),
		Transactions: make([]*Transaction, nTransactions),
	}
	for i := 0; i < nTransactions; i++ {
		b.Transactions[i] = NewTransaction()
	}
	return &b
}

// NewBlockChain produces random blockchain.
func NewBlockChain() *BlockChain {
	// Uniform distribution on [10, 50]
	nBlocks := mrand.Intn(40) + 10
	bc := BlockChain{Blocks: make([]*Block, nBlocks)}
	for i := 0; i < nBlocks; i++ {
		bc.Blocks[i] = NewBlock()
	}
	bc.Head = HashSum(bc.Blocks[nBlocks-1])
	return &bc
}

// NewInputBlock produces new block with given transactions.
func NewInputBlock(t []*Transaction) *Block {
	return &Block{
		BlockHeader: BlockHeader{
			BlockNumber: 0,
			LastBlock:   NewHash(),
			Target:      NewValidTarget(),
			Time:        uint32(time.Now().Unix()),
			Nonce:       0,
		},
		Transactions: t,
	}
}

// NewOutputBlock produces new block with given transactions and given input
// block.
func NewOutputBlock(t []*Transaction, input *Block) *Block {
	return &Block{
		BlockHeader: BlockHeader{
			BlockNumber: input.BlockNumber + 1,
			LastBlock:   HashSum(input),
			Target:      NewValidTarget(),
			Time:        uint32(time.Now().Unix()),
			Nonce:       0,
		},
		Transactions: t,
	}
}

// NewTransactionValue creates a new transaction with specific value a.
func NewTransactionValue(s, r Wallet, a uint64, i uint32) (*Transaction, error) {
	tbody := TxBody{
		Sender: s.Public(),
		Input: TxHashPointer{
			BlockNumber: 0,
			Hash:        NewHash(),
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

	trA, _ := NewTransactionValue(original, sender, 2, 1)
	trA.Outputs = append(trA.Outputs, TxOutput{
		Amount:    2,
		Recipient: sender.Public(),
	})

	trB, _ := NewTransactionValue(sender, recipient, 4, 0)
	trB.Input.Hash = HashSum(trA)

	trB, _ = trB.TxBody.Sign(sender, crand.Reader)

	inputTransactions := []*Transaction{trA}
	outputTransactions := []*Transaction{trB}

	inputBlock := NewInputBlock(inputTransactions)
	outputBlock := NewOutputBlock(outputTransactions, inputBlock)

	return &BlockChain{
		Blocks: []*Block{inputBlock, outputBlock},
		Head:   NewHash(),
	}, recipient
}

// NewValidChainAndBlock creates a valid BlockChain and a Block that is valid
// with respect to the BlockChain.
func NewValidChainAndBlock() (*BlockChain, *Block) {
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
	newBlock := NewOutputBlock([]*Transaction{tr}, inputBlock)
	return bc, newBlock
}

// NewValidTarget creates a new valid target that is a random value between the
// max and min difficulties
func NewValidTarget() Hash {
	r := new(big.Int).Rand(
		mrand.New(mrand.NewSource(time.Now().Unix())),
		new(big.Int).Add(MinDifficulty, big.NewInt(1)),
	)
	r.Add(r, big.NewInt(1))
	return BigIntToHash(r)
}

// BigIntToHash converts a big integer to a hash
func BigIntToHash(x *big.Int) Hash {
	bytes := x.Bytes()

	var result Hash
	for i := 0; i < HashLen; i++ {
		result[i] = 0
	}

	if len(bytes) > HashLen {
		return result
	}

	for i := 0; i < len(bytes); i++ {
		result[len(bytes)-1-i] = bytes[i]
	}
	return result
}
