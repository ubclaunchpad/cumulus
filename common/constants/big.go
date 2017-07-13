package constants

import "github.com/ubclaunchpad/cumulus/common/math"
import "math/big"

var (
	// Big0 is 0 represented as type big
	Big0 = big.NewInt(0)
	// Big1 is 1 represented as type big
	Big1 = big.NewInt(1)
	// Big2E256 is 2^256 respresented as type big
	Big2E256 = math.BigExp(2, 256)
)

var (
	// MaxUint256 is the maximum uint256 number
	MaxUint256 = math.BigSub(Big2E256, Big1)
)
