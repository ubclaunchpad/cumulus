package cumuluspeer_test

import (
    "testing"

    cumuluspeer "github.com/ubclaunchpad/cumulus/cumuluspeer"
)

// Tests if we can make a basic host on a valid TCP port
func TestMakeBasicHostValidPort(t *testing.T) {
    h, ps, err := cumuluspeer.MakeBasicHost(cumuluspeer.DefaultIP, 8000)
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

// Make sure MakeBasicHost fails with invalid IP
func TestMakeBasicHostInvalidIP(t *testing.T) {
    _, _, err := cumuluspeer.MakeBasicHost("asdfasdf", 123)
    if err == nil {
        t.Fail()
    }
}

// TODO Test ExtractPeerInfo
// TODO Test BasicStreamHandler
