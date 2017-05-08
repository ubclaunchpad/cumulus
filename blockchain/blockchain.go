package blockchain

import (
	"crypto/sha256"
	"encoding/binary"
)

type BlockHeader struct {
	blockNumber uint32
	lastBlock   Hash
	miner       Wallet
}

func (bh *BlockHeader) Marshal() []byte {
	buf := []byte{}
	binary.LittleEndian.PutUint32(buf, bh.blockNumber)
	buf = append(buf, bh.lastBlock...)
	buf = append(buf, bh.miner.Marshal()...)
	return buf
}

type Block struct {
	BlockHeader
	transactions []Transaction
}

func (b *Block) Marshal() []byte {
	buf := b.BlockHeader.Marshal()
	for _, t := range b.transactions {
		buf = append(buf, t.Marshal()...)
	}
	return buf
}

func (b *Block) Hash() []byte {
	sum := sha256.Sum256(b.Marshal())
	return sum[:]
}

type BlockChain []Block

func (bc BlockChain) Verify(b *Block) {

}
