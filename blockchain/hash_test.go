package blockchain

import (
	"math/big"
	"testing"
)

func TestHashToBigInt(t *testing.T) {
	// Max hash value
	x := new(big.Int).Sub(new(big.Int).Exp(big.NewInt(int64(2)), big.NewInt(int64(256)), big.NewInt(0)), big.NewInt(1))
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
