package app

import (
	"github.com/ubclaunchpad/cumulus/blockchain"
	"github.com/ubclaunchpad/cumulus/msg"
	"github.com/ubclaunchpad/cumulus/peer"
	"github.com/ubclaunchpad/cumulus/pool"
)

func createNewTestBlockRequest(lastBlock interface{}) *msg.Request {
	params := make(map[string]interface{}, 1)
	params["lastBlock"] = lastBlock
	return &msg.Request{
		ResourceType: msg.ResourceBlock,
		Params:       params,
	}
}

func createNewTestApp() *App {
	chain, _ := blockchain.NewValidTestChainAndBlock()
	return &App{
		PeerStore:   peer.NewPeerStore("127.0.0.1:8000"),
		CurrentUser: NewUser(),
		Chain:       chain,
		Pool:        pool.New(),
	}
}
