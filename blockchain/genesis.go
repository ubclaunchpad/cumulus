package blockchain

import (
	"time"
)

// Genesis creates the Genesis block and returns is.
//
// Properties of the Genesis block:
// 	- BlockNumber = 0
// 	- LastBlock = 0
// 	- There is only one transaction in the block, the CloudBase transaction that
// 	  awards the miner with the block reward.
func Genesis(miner Address, target Hash, blockReward uint64, extraData []byte) *Block {

	cbReward := TxOutput{
		Amount:    blockReward,
		Recipient: miner,
	}

	cbTx := &Transaction{
		TxBody: TxBody{
			Sender: NilAddr,
			Input: TxHashPointer{
				BlockNumber: 0,
				Hash:        NilHash,
				Index:       0,
			},
			Outputs: []TxOutput{cbReward},
		},
		Sig: NilSig,
	}

	genesisBlock := &Block{
		BlockHeader: BlockHeader{
			BlockNumber: 0,
			LastBlock:   NilHash,
			Target:      target,
			Time:        uint32(time.Now().Unix()),
			Nonce:       0,
			ExtraData:   extraData,
		},
		Transactions: []*Transaction{cbTx},
	}

	return genesisBlock
}
