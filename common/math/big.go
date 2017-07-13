package math

import "math/big"

// BigExp returns a big int pointer to the result of base to the power of exp,
// if exp <= 0, the result is 1
func BigExp(base, exp int) *big.Int {
	return new(big.Int).Exp(big.NewInt(int64(base)), big.NewInt(int64(exp)), nil)
}

// BigSub returns a big int pointer to the result of the subtraction of y from
// x.
func BigSub(x, y *big.Int) *big.Int {
	return new(big.Int).Sub(x, y)
}

// BigAdd returns a big int pointer to the result of the addition of x to y.
func BigAdd(x, y *big.Int) *big.Int {
	return new(big.Int).Add(x, y)
}
