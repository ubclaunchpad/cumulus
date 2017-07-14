package constants

import (
	"math/big"

	"github.com/ubclaunchpad/cumulus/common/util"
)

// Commonly used "math/big" type numbers
var (
	// Big0 is 0 represented as type "math/big"
	Big0 = big.NewInt(0)
	// Big1 is 1 represented as type "math/big"
	Big1 = big.NewInt(1)
	// Big2Exp256 is 2^256 respresented as type "math/big"
	Big2Exp256 = util.BigExp(2, 256)
)

// Commonly used max values represented as type "math/big"
var (
	// MaxUint256 is the maximum uint256 number represented as type "math/big"
	MaxUint256 = util.BigSub(Big2Exp256, Big1)
)
