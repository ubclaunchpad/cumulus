package cumuluspeer_test

import (
    "testing"

    cumuluspeer "github.com/ubclaunchpad/cumulus/cumuluspeer"
)

// Tests if we can make a basic host on a valid TCP port
func TestMakeBasicHostValidPort(t *testing.T) {
    h, ps, err := cumuluspeer.MakeBasicHost(8000)
    if err != nil {
        t.Fail()
    }

    if h == nil {
        t.Fail()
    }

    if ps == nil {
        t.Fail()
    }

    if ps != h.Peerstore() {
        t.Fail()
    }
}

// TODO Test ExtractPeerInfo
// TODO Test BasicStreamHandler
