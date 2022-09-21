package lfs

import (
	"math/big"
)

// FourInt is the 4-number big integer group
type FourInt [4]*big.Int

// NewFourInt creates a new 4-number group, in descending order
func NewFourInt(w1 *big.Int, w2 *big.Int, w3 *big.Int, w4 *big.Int) FourInt {
	w1.Abs(w1)
	w2.Abs(w2)
	w3.Abs(w3)
	w4.Abs(w4)
	// sort the four big integers in descending order
	if w1.Cmp(w2) == -1 {
		w1, w2 = w2, w1
	}
	if w1.Cmp(w3) == -1 {
		w1, w3 = w3, w1
	}
	if w1.Cmp(w4) == -1 {
		w1, w4 = w4, w1
	}
	if w2.Cmp(w3) == -1 {
		w2, w3 = w3, w2
	}
	if w2.Cmp(w4) == -1 {
		w2, w4 = w4, w2
	}
	if w3.Cmp(w4) == -1 {
		w3, w4 = w4, w3
	}
	return FourInt{w1, w2, w3, w4}
}

// Mul multiplies all the 4 numbers by n
func (f *FourInt) Mul(n *big.Int) {
	for i := 0; i < 4; i++ {
		f[i].Mul(f[i], n)
	}
}

// Div divides all the 4 numbers by n
func (f *FourInt) Div(n *big.Int) {
	for i := 0; i < 4; i++ {
		f[i].Div(f[i], n)
	}
}

// String convert the FourInt object to string
func (f *FourInt) String() string {
	res := "{"
	for i := 0; i < 3; i++ {
		res += f[i].String()
		res += ", "
	}
	res += f[3].String()
	res += "}"
	return res
}
