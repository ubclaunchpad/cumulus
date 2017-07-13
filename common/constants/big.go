package constants

import "github.com/ubclaunchpad/cumulus/common/math"
import "math/big"

// Commonly used "math/big" type numbers
var (
	// Big0 is 0 represented as type big
	Big0 = big.NewInt(0)
	// Big1 is 1 represented as type big
	Big1 = big.NewInt(1)
	// Big2Exp256 is 2^256 respresented as type big
	Big2Exp256 = math.BigExp(2, 256)
)

// Commonly used max values represented as type "math/big"
var (
	// MaxUint256 is the maximum uint256 number
	MaxUint256 = math.BigSub(Big2Exp256, Big1)
)
