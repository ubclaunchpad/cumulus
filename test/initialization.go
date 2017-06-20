package test

import (
	"os/exec"
	"strconv"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/ubclaunchpad/cumulus/conn"
	"github.com/ubclaunchpad/cumulus/peer"
)

var _ = Describe("Initialization", func() {
	var instance *exec.Cmd

	BeforeEach(func() {
		var err error
		instance, err = startInstance()
		Expect(err).ToNot(HaveOccurred())
	})

	AfterEach(func() {
		if instance.Process != nil {
			instance.Process.Kill()
		}
	})

	// TODO fails until new peer implementation is merged in
	PIt("should accept a TCP connection", func() {
		_, err := conn.Dial(peer.DefaultIP + ":" + strconv.Itoa(peer.DefaultPort))
		Expect(err).ToNot(HaveOccurred())
	})
})
