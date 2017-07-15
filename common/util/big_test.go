package util

import (
	"math/big"
	"testing"
)

func TestBigExp(t *testing.T) {
	a := big.NewInt(1)
	b := BigExp(0, 0)

	if a.Cmp(b) != 0 {
		t.Fail()
	}

	a = big.NewInt(1)
	b = BigExp(10, -2)

	if a.Cmp(b) != 0 {
		t.Fail()
	}

	a = new(big.Int).Exp(
		big.NewInt(int64(2)),
		big.NewInt(int64(256)),
		big.NewInt(0),
	)
	b = BigExp(2, 256)

	if a.Cmp(b) != 0 {
		t.Fail()
	}
}

func TestBigSub(t *testing.T) {
	a := big.NewInt(0)
	b := BigSub(big.NewInt(0), big.NewInt(0))

	if a.Cmp(b) != 0 {
		t.Fail()
	}

	a = big.NewInt(1)
	b = BigSub(big.NewInt(1), big.NewInt(0))

	if a.Cmp(b) != 0 {
		t.Fail()
	}

	a = big.NewInt(-1)
	b = BigSub(big.NewInt(0), big.NewInt(1))

	if a.Cmp(b) != 0 {
		t.Fail()
	}

	a = big.NewInt(2)
	b = BigSub(big.NewInt(7), big.NewInt(5))

	if a.Cmp(b) != 0 {
		t.Fail()
	}

	a = big.NewInt(2)
	b = BigSub(big.NewInt(4), big.NewInt(2))

	if a.Cmp(b) != 0 {
		t.Fail()
	}
}

func TestBigAdd(t *testing.T) {
	a := big.NewInt(0)
	b := BigAdd(big.NewInt(0), big.NewInt(0))

	if a.Cmp(b) != 0 {
		t.Fail()
	}

	a = big.NewInt(1)
	b = BigAdd(big.NewInt(1), big.NewInt(0))

	if a.Cmp(b) != 0 {
		t.Fail()
	}

	a = big.NewInt(1)
	b = BigAdd(big.NewInt(0), big.NewInt(1))

	if a.Cmp(b) != 0 {
		t.Fail()
	}

	a = big.NewInt(12)
	b = BigAdd(big.NewInt(7), big.NewInt(5))

	if a.Cmp(b) != 0 {
		t.Fail()
	}

	a = big.NewInt(6)
	b = BigAdd(big.NewInt(4), big.NewInt(2))

	if a.Cmp(b) != 0 {
		t.Fail()
	}
}
