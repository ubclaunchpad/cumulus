package consensus

import (
	"testing"

	"github.com/ubclaunchpad/cumulus/blockchain"
	c "github.com/ubclaunchpad/cumulus/common/constants"
)

func TestCurrentTarget(t *testing.T) {
	if blockchain.HashToBigInt(CurrentTarget()).Cmp(c.MaxTarget) != 0 {
		t.Fail()
	}
}
