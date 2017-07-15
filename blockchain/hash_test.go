package blockchain

import (
	"math/big"
	"testing"

	c "github.com/ubclaunchpad/cumulus/common/constants"
	"github.com/ubclaunchpad/cumulus/common/util"
)

func TestBigIntToHash(t *testing.T) {
	x := util.BigAdd(util.BigExp(2, 256), c.Big1)
	h := BigIntToHash(x)

	for i := 0; i < HashLen; i++ {
		if h[i] != 0 {
			t.Fail()
		}
	}

	h = BigIntToHash(c.Big0)

	for i := 0; i < HashLen; i++ {
		if h[i] != 0 {
			t.Fail()
		}
	}

	h = BigIntToHash(c.MaxUint256)

	for i := 0; i < HashLen; i++ {
		if h[i] != 255 {
			t.Fail()
		}
	}
}

func TestHashToBigInt(t *testing.T) {
	// Max hash value
	x := c.MaxUint256
	h := BigIntToHash(x)

	if HashToBigInt(h).Cmp(x) != 0 {
		t.Fail()
	}

	// Half of max hash value
	x = x.Div(x, big.NewInt(2))
	h = BigIntToHash(x)

	if HashToBigInt(h).Cmp(x) != 0 {
		t.Fail()
	}

	// Min hash value
	x = big.NewInt(1)
	h = BigIntToHash(x)

	if HashToBigInt(h).Cmp(x) != 0 {
		t.Fail()
	}
}

func TestLessThan(t *testing.T) {
	a := BigIntToHash(big.NewInt(0))
	b := BigIntToHash(big.NewInt(0))

	// Test equal to
	if a.LessThan(b) {
		t.Fail()
	}

	// Test less than
	b = BigIntToHash(big.NewInt(1))
	if !a.LessThan(b) {
		t.Fail()
	}

	// Test greater than
	a = BigIntToHash(big.NewInt(2))
	if a.LessThan(b) {
		t.Fail()
	}
}
