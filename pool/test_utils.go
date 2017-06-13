package pool

import (
	"crypto/elliptic"
	crand "crypto/rand"
	"crypto/sha256"
	mrand "math/rand"

	"github.com/ubclaunchpad/cumulus/blockchain"
)

var curve = elliptic.P256()

func newHash() blockchain.Hash {
	message := make([]byte, 256)
	crand.Read(message)
	return sha256.Sum256(message)
}

func newTxHashPointer() blockchain.TxHashPointer {
	return blockchain.TxHashPointer{
		BlockNumber: mrand.Uint32(),
		Hash:        newHash(),
		Index:       mrand.Uint32(),
	}
}

func newTxOutput() blockchain.TxOutput {
	return blockchain.TxOutput{
		Amount:    uint64(mrand.Int63()),
		Recipient: blockchain.NewWallet().Public(),
	}
}

func newTxBody() blockchain.TxBody {
	// Uniform distribution on [1, 4]
	nOutputs := mrand.Intn(4) + 1
	body := blockchain.TxBody{
		Sender:  blockchain.NewWallet().Public(),
		Input:   newTxHashPointer(),
		Outputs: make([]blockchain.TxOutput, nOutputs),
	}
	for i := 0; i < nOutputs; i++ {
		body.Outputs[i] = newTxOutput()
	}
	return body
}

func newTransaction() *blockchain.Transaction {
	sender := blockchain.NewWallet()
	tbody := newTxBody()
	t, _ := tbody.Sign(sender, crand.Reader)
	return t
}

func newBlockHeader() blockchain.BlockHeader {
	return blockchain.BlockHeader{
		BlockNumber: mrand.Uint32(),
		LastBlock:   newHash(),
		Miner:       blockchain.NewWallet().Public(),
	}
}

func newBlock() *blockchain.Block {
	// Uniform distribution on [500, 999]
	nTransactions := mrand.Intn(500) + 500
	b := blockchain.Block{
		BlockHeader:  newBlockHeader(),
		Transactions: make([]*blockchain.Transaction, nTransactions),
	}
	for i := 0; i < nTransactions; i++ {
		b.Transactions[i] = newTransaction()
	}
	return &b
}

func newBlockChain() *blockchain.BlockChain {
	// Uniform distribution on [10, 50]
	nBlocks := mrand.Intn(40) + 10
	bc := blockchain.BlockChain{Blocks: make([]*blockchain.Block, nBlocks)}
	for i := 0; i < nBlocks; i++ {
		bc.Blocks[i] = newBlock()
	}
	bc.Head = blockchain.HashSum(bc.Blocks[nBlocks-1])
	return &bc
}
