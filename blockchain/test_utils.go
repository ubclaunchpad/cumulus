package blockchain

import (
	"crypto/ecdsa"
	crand "crypto/rand"
	"crypto/sha256"
	mrand "math/rand"
)

func newHash() Hash {
	message := make([]byte, 256)
	crand.Read(message)
	return sha256.Sum256(message)
}

func newWallet() Wallet {
	priv, _ := ecdsa.GenerateKey(curve, crand.Reader)
	return (*wallet)(priv)
}

func newTxHashPointer() TxHashPointer {
	return TxHashPointer{
		BlockNumber: mrand.Uint32(),
		Hash:        newHash(),
	}
}

func newTxOutput() TxOutput {
	return TxOutput{
		Amount:    uint64(mrand.Int63()),
		Recipient: newWallet().Public(),
	}
}

func newTxBody() TxBody {
	// Uniform distribution on [1, 4]
	nOutputs := mrand.Intn(4) + 1
	body := TxBody{
		Sender:  newWallet().Public(),
		Input:   newTxHashPointer(),
		Outputs: make([]TxOutput, nOutputs),
	}
	for i := 0; i < nOutputs; i++ {
		body.Outputs[i] = newTxOutput()
	}
	return body
}

func newTransaction() *Transaction {
	sender := newWallet()
	tbody := newTxBody()
	t, _ := tbody.Sign(sender, crand.Reader)
	return t
}

func newBlockHeader() BlockHeader {
	return BlockHeader{
		BlockNumber: mrand.Uint32(),
		LastBlock:   newHash(),
		Miner:       newWallet().Public(),
	}
}

func newBlock() *Block {
	// Uniform distribution on [500, 999]
	nTransactions := mrand.Intn(500) + 500
	b := Block{
		BlockHeader:  newBlockHeader(),
		Transactions: make([]*Transaction, nTransactions),
	}
	for i := 0; i < nTransactions; i++ {
		b.Transactions[i] = newTransaction()
	}
	return &b
}

func newBlockChain() *BlockChain {
	// Uniform distribution on [10, 50]
	nBlocks := mrand.Intn(40) + 10
	bc := BlockChain{Blocks: make([]*Block, nBlocks)}
	for i := 0; i < nBlocks; i++ {
		bc.Blocks[i] = newBlock()
	}
	bc.Head = HashSum(bc.Blocks[nBlocks-1])
	return &bc
}

func newInputBlock(t []*Transaction) *Block {
	return &Block{
		BlockHeader: BlockHeader{
			BlockNumber: 0,
			LastBlock:   newHash(),
			Miner:       newWallet().Public(),
		},
		Transactions: t,
	}
}

func newOutputBlock(t []*Transaction, input *Block) *Block {
	return &Block{
		BlockHeader: BlockHeader{
			BlockNumber: input.BlockNumber + 1,
			LastBlock:   HashSum(input),
			Miner:       newWallet().Public(),
		},
		Transactions: t,
	}
}

func newTransactionValue(s, r Wallet, a uint64) (*Transaction, error) {
	tbody := TxBody{
		Sender: s.Public(),
		Input: TxHashPointer{
			BlockNumber: 0,
			Hash:        newHash(),
		},
		Outputs: make([]TxOutput, 1),
	}
	tbody.Outputs[0] = TxOutput{
		Amount:    a,
		Recipient: r.Public(),
	}
	return tbody.Sign(s, crand.Reader)
}

func newValidBlockChainFixture() (*BlockChain, Wallet) {
	original := newWallet()
	sender := newWallet()
	recipient := newWallet()

	trA, _ := newTransactionValue(original, sender, 2)
	trA.Outputs = append(trA.Outputs, TxOutput{
		Amount:    2,
		Recipient: sender.Public(),
	})

	trB, _ := newTransactionValue(sender, recipient, 4)
	trB.Input.Hash = HashSum(trA)

	trB, _ = trB.TxBody.Sign(sender, crand.Reader)

	inputTransactions := []*Transaction{trA}
	outputTransactions := []*Transaction{trB}

	inputBlock := newInputBlock(inputTransactions)
	outputBlock := newOutputBlock(outputTransactions, inputBlock)

	return &BlockChain{
		Blocks: []*Block{inputBlock, outputBlock},
		Head:   newHash(),
	}, recipient
}
