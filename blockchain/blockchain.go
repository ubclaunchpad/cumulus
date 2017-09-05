package blockchain

import (
	"encoding/json"
	"errors"
	"os"
	"sync"
)

// BlockChain represents a linked list of blocks
type BlockChain struct {
	Blocks []*Block
	Head   Hash
	lock   *sync.RWMutex
}

// New returns a new blockchain
func New() *BlockChain {
	return &BlockChain{
		Blocks: make([]*Block, 0),
		Head:   NilHash,
		lock:   &sync.RWMutex{},
	}
}

// RLock locks the blockchain for reading
func (bc *BlockChain) RLock() {
	bc.lock.RLock()
}

// RUnlock locks the blockchain for reading
func (bc *BlockChain) RUnlock() {
	bc.lock.RUnlock()
}

// Lock locks the blockchain for reading
func (bc *BlockChain) Lock() {
	bc.lock.Lock()
}

// Unlock locks the blockchain for reading
func (bc *BlockChain) Unlock() {
	bc.lock.Unlock()
}

// Len returns the length of the BlockChain when marshalled
func (bc *BlockChain) Len() int {
	return len(bc.Marshal())
}

// Marshal converts the BlockChain to a byte slice.
func (bc *BlockChain) Marshal() []byte {
	var buf []byte
	for _, b := range bc.Blocks {
		buf = append(buf, b.Marshal()...)
	}
	return append(buf, bc.Head.Marshal()...)
}

// Save writes the blockchain to a file of the given same in the current
// working directory in JSON format. It returns an error if one occurred, or a
// pointer to the file that was written to on success.
func (bc *BlockChain) Save(fileName string) error {
	file, err := os.OpenFile(fileName, os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	blockchainBytes, err := json.Marshal(bc)
	if err != nil {
		return err
	}

	if _, err = file.Write(blockchainBytes); err != nil {
		return err
	}
	return nil
}

// Load attempts to read blockchain info from the file with the given name in the
// current working directory in JSON format. On success this returns
// a pointer to a new user constructed from the information in the file.
// If an error occurrs it is returned.
func Load(fileName string) (*BlockChain, error) {
	file, err := os.OpenFile(fileName, os.O_RDONLY|os.O_CREATE, 0644)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	dec := json.NewDecoder(file)
	dec.UseNumber()

	var bc BlockChain
	if err := dec.Decode(&bc); err != nil {
		return nil, err
	}
	bc.lock = &sync.RWMutex{}
	return &bc, nil
}

// AppendBlock adds a block to the end of the block chain.
func (bc *BlockChain) AppendBlock(b *Block) {
	bc.Blocks = append(bc.Blocks, b)
	bc.Head = HashSum(b)
}

// LastBlock returns a pointer to the last block in the given blockchain, or nil
// if the blockchain is empty.
func (bc *BlockChain) LastBlock() *Block {
	if len(bc.Blocks) == 0 {
		return nil
	}
	return bc.Blocks[len(bc.Blocks)-1]
}

// GetInputTransaction returns the input Transaction referenced by TxHashPointer.
// If the Transaction does not exist, then GetInputTransaction returns nil.
func (bc *BlockChain) GetInputTransaction(t *TxHashPointer) *Transaction {
	if t.BlockNumber > uint32(len(bc.Blocks)) {
		return nil
	}
	b := bc.Blocks[t.BlockNumber]
	if t.Index > uint32(len(b.Transactions)) {
		return nil
	}
	return b.Transactions[t.Index]
}

// GetAllInputs returns all the transactions referenced by a transaction
// as inputs. Returns an error if any of the transactios requested could
// not be found.
func (bc *BlockChain) GetAllInputs(t *Transaction) ([]*Transaction, error) {
	txns := []*Transaction{}
	for _, tx := range t.Inputs {
		nextTxn := bc.GetInputTransaction(&tx)
		if nextTxn == nil {
			return nil, errors.New("input transaction not found")
		}
		txns = append(txns, nextTxn)
	}
	return txns, nil
}

// ContainsTransaction returns true if the BlockChain contains the transaction
// in a block between start and stop as indexes.
func (bc *BlockChain) ContainsTransaction(t *Transaction, start, stop uint32) (bool, uint32, uint32) {
	for i := start; i < stop; i++ {
		if exists, j := bc.Blocks[i].ContainsTransaction(t); exists {
			return true, i, j
		}
	}
	return false, 0, 0
}

// GetBlockByLastBlockHash returns a copy of the block in the local chain that
// comes directly after the block with the given hash. Returns error if no such
// block is found.
func (bc *BlockChain) GetBlockByLastBlockHash(hash Hash) (*Block, error) {
	// Find the block with the given hash
	for _, block := range bc.Blocks {
		if block.LastBlock == hash {
			return block, nil
		}
	}
	return nil, errors.New("No such block")
}

// RollBack removes the last block from the blockchain. Returns the block that
// was removed from the end of the chain, or nil if the blockchain is empty.
func (bc *BlockChain) RollBack() *Block {
	if len(bc.Blocks) == 0 {
		return nil
	}
	prevHead := bc.LastBlock()
	bc.Blocks = bc.Blocks[:len(bc.Blocks)-1]
	if len(bc.Blocks) == 0 {
		bc.Head = NilHash
	} else {
		bc.Head = HashSum(bc.LastBlock())
	}
	return prevHead
}
