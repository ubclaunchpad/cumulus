package main

import (
    "fmt"
    "log"

    golog "github.com/ipfs/go-log"
    gologging "github.com/whyrusleeping/go-logging"
    cmlPeer "github.com/ubclaunchpad/cumulus/cumulus-peer"
)

func main() {
    fmt.Println("Starting Cumulus Peer")

    // Set up a logger. Change to DEBUG for extra info
    golog.SetAllLoggers(gologging.INFO)

    // Set up a new host on the Cumulus network
    host, err := cmlPeer.MakeHost()
    if err != nil {
        log.Fatal(err)
    }

    // TODO: remove this when we figure out what to do with this host.
    fmt.Println(host.ID())
}
