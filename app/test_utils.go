package app

import (
	"github.com/ubclaunchpad/cumulus/blockchain"
	"github.com/ubclaunchpad/cumulus/miner"
	"github.com/ubclaunchpad/cumulus/msg"
	"github.com/ubclaunchpad/cumulus/peer"
	"github.com/ubclaunchpad/cumulus/pool"
)

func newTestBlockRequest(lastBlockHash interface{}) *msg.Request {
	params := make(map[string]interface{}, 1)
	params["lastBlockHash"] = lastBlockHash
	return &msg.Request{
		ResourceType: msg.ResourceBlock,
		Params:       params,
	}
}

func newTestApp() *App {
	chain, _ := blockchain.NewValidTestChainAndBlock()
	return &App{
		PeerStore:        peer.NewPeerStore("127.0.0.1:8000"),
		CurrentUser:      NewUser(),
		Chain:            chain,
		Miner:            miner.New(),
		Pool:             pool.New(),
		blockQueue:       make(chan *blockchain.Block, blockQueueSize),
		transactionQueue: make(chan *blockchain.Transaction, transactionQueueSize),
		quitChan:         make(chan bool),
	}
}
