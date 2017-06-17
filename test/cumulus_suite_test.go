package test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestCumulus(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Cumulus Suite")
}
