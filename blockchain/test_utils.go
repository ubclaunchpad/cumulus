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
	digest := HashSum(tbody)
	sig := sender.Sign(digest, crand.Reader)
	return &Transaction{
		TxBody: tbody,
		Sig:    sig,
	}
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
