package blockchain

import (
	crand "crypto/rand"
	"crypto/sha256"
	"math"
	"math/big"
	mrand "math/rand"
	"sync"

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
	bc.lock = &sync.RWMutex{}
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

// NewValidBlockChainFixture creates a valid blockchain of three blocks
// and returns the wallets involved in the transactions.
// The returning wallets will have balances of 3, 1, and 0 respectively.
func NewValidBlockChainFixture() (*BlockChain, map[string]*Wallet) {
	sender := NewWallet()
	alice := NewWallet()
	bob := NewWallet()

	// Cloud base txns for our blocks.
	cbA, _ := NewValidCloudBaseTestTransaction()
	cbB, _ := NewValidCloudBaseTestTransaction()
	cbC, _ := NewValidCloudBaseTestTransaction()

	// Transaction A is at index 1 in block 0 (sender awarded 4 coins).
	tA, _ := TxBody{
		Sender: sender.Public(),
		Inputs: []TxHashPointer{},
		Outputs: []TxOutput{
			TxOutput{
				Amount:    2,
				Recipient: sender.Public().Repr(),
			},
			TxOutput{
				Amount:    2,
				Recipient: sender.Public().Repr(),
			},
		},
	}.Sign(*sender, crand.Reader)

	block0 := Block{
		BlockHeader: BlockHeader{
			BlockNumber: 0,
			LastBlock:   NewTestHash(),
			Target:      NewValidTestTarget(),
			Time:        mrand.Uint32(),
			Nonce:       0,
		},
		// Block0 is a cb and a transaction.
		Transactions: []*Transaction{cbA, tA},
	}

	// Transaction B is at index 1 in block 1 (sender sends 3 coins to alice).
	tB, _ := TxBody{
		Sender: sender.Public(),
		Inputs: []TxHashPointer{
			TxHashPointer{
				// Reference block 0, index 0 for inputs.
				BlockNumber: 0,
				Index:       1, // Cloudbase will bump transactions forward.
				Hash:        HashSum(tA),
			},
		},
		// Send some outputs to alice, some back to sender.
		Outputs: []TxOutput{
			TxOutput{
				Amount:    3,
				Recipient: alice.Public().Repr(),
			},
			TxOutput{
				Amount:    1,
				Recipient: sender.Public().Repr(),
			},
		},
	}.Sign(*sender, crand.Reader)

	// Block1 is a cb and a transaction.
	block1 := Block{
		BlockHeader: BlockHeader{
			BlockNumber: 1,
			LastBlock:   HashSum(block0),
			Target:      NewValidTestTarget(),
			Time:        mrand.Uint32(),
			Nonce:       0,
		},
		Transactions: []*Transaction{cbB, tB},
	}

	// Sender has 1 coin left to send to bob.
	tC, _ := TxBody{
		Sender: sender.Public(),
		Inputs: []TxHashPointer{
			TxHashPointer{
				// Again look at block 1.
				BlockNumber: 1,
				Index:       1, // skip cb
				Hash:        HashSum(tB),
			},
		},
		// One coin output to bob.
		Outputs: []TxOutput{
			TxOutput{
				Amount:    1,
				Recipient: bob.Public().Repr(),
			},
		},
	}.Sign(*sender, crand.Reader)

	// Block2 is a cb and a transaction.
	block2 := Block{
		BlockHeader: BlockHeader{
			BlockNumber: 2,
			LastBlock:   HashSum(block1),
			Target:      NewValidTestTarget(),
			Time:        mrand.Uint32(),
			Nonce:       0,
		},
		Transactions: []*Transaction{cbC, tC},
	}

	wallets := map[string]*Wallet{
		"alice":  alice,
		"bob":    bob,
		"sender": sender,
	}

	bc := New()
	bc.Blocks = []*Block{&block0, &block1, &block2}
	bc.Head = NewTestHash()
	return bc, wallets
}

// NewValidTestChainAndBlock creates a valid BlockChain of 3 blocks,
// and a new block which is valid with respect to the blockchain.
func NewValidTestChainAndBlock() (*BlockChain, *Block) {
	bc, wallets := NewValidBlockChainFixture()

	// Alice wants to send 2 coins to bob and bob wants to send
	// his coin back to the sender.
	aliceToBob, _ := TxBody{
		Sender: wallets["alice"].Public(),
		Inputs: []TxHashPointer{
			TxHashPointer{
				// Block 1, transaction 1 is where this input comes from.
				BlockNumber: 1,
				Index:       1,
				Hash:        HashSum(bc.Blocks[1].Transactions[1]),
			},
		},
		// One output to bob, one back to alice.
		Outputs: []TxOutput{
			TxOutput{
				Amount:    2,
				Recipient: wallets["bob"].Public().Repr(),
			},
			TxOutput{
				Amount:    1,
				Recipient: wallets["alice"].Public().Repr(),
			},
		},
	}.Sign(*wallets["alice"], crand.Reader)

	bobToSender, _ := TxBody{
		Sender: wallets["bob"].Public(),
		Inputs: []TxHashPointer{
			TxHashPointer{
				// Block 2, transaction 1 is where this input comes from.
				BlockNumber: 2,
				Index:       1,
				Hash:        HashSum(bc.Blocks[2].Transactions[1]),
			},
		},
		// One output to sender.
		Outputs: []TxOutput{
			TxOutput{
				Amount:    1,
				Recipient: wallets["sender"].Public().Repr(),
			},
		},
	}.Sign(*wallets["bob"], crand.Reader)

	cb, _ := NewValidCloudBaseTestTransaction()

	// Update CloudBase transaction amount so it fits the blockchain
	timesHalved := float64((len(bc.Blocks) / 210000 /* Block reward halving rate */))
	cb.Outputs[0].Amount = 25 * 2 << 32 /* Starting block reward */ /
		uint64(math.Pow(float64(2), timesHalved))

	blk := Block{
		BlockHeader: BlockHeader{
			BlockNumber: 3,
			LastBlock:   HashSum(bc.Blocks[2]),
			Target:      NewValidTestTarget(),
			Time:        mrand.Uint32(),
			Nonce:       0,
		},
		Transactions: []*Transaction{cb, aliceToBob, bobToSender},
	}

	return bc, &blk
}

// NewValidChainAndTxn creates a valid BlockChain of 3 blocks,
// and a Transaction that is valid with respect to the BlockChain.
func NewValidChainAndTxn() (*BlockChain, *Transaction) {
	bc, b := NewValidTestChainAndBlock()
	return bc, b.Transactions[1]
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
		Amount:    25 * 2 << 32, // Starting block reward
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
