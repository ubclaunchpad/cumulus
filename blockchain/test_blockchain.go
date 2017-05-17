package blockchain

import (
	"testing"

	log "github.com/Sirupsen/logrus"
)

// TestMain sets logging levels for tests.
func TestMain(t *testing.T) {
	log.SetLevel(log.DebugLevel)
}

// TestValidTransaction tests the three cases in which a
// transaction can fail to be valid.
func TestValidTransaction() {
	// t does not appear in the input block.
	// The output(s) in the inputBlock do not equal the outputs in t.
	// The signature is invalid
}

// TestValidBlock tests the three cases in which a block can fail to be valid.
func TestValidBlock() {
	// b conains invalid transaction.
	// Block number for b is not one greater than the last block.
	// The hash of the last block is incorrect.
}
