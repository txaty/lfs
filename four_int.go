package lfs

import (
	"math/big"
	"sort"
	"strings"
)

// FourInt represents a group of four big.Int values.
type FourInt [4]*big.Int

// NewFourInt creates a new FourInt with its components sorted in descending order.
func NewFourInt(w1, w2, w3, w4 *big.Int) FourInt {
	ints := []*big.Int{
		new(big.Int).Abs(w1),
		new(big.Int).Abs(w2),
		new(big.Int).Abs(w3),
		new(big.Int).Abs(w4),
	}
	sort.Slice(ints, func(i, j int) bool {
		return ints[i].Cmp(ints[j]) > 0
	})
	return FourInt{ints[0], ints[1], ints[2], ints[3]}
}

// Mul multiplies each component of FourInt by n.
func (f *FourInt) Mul(n *big.Int) {
	for i := range f {
		f[i].Mul(f[i], n)
	}
}

// Div divides each component of FourInt by n.
func (f *FourInt) Div(n *big.Int) {
	for i := range f {
		f[i].Div(f[i], n)
	}
}

// String returns a string representation of FourInt.
func (f *FourInt) String() string {
	parts := make([]string, len(f))
	for i, num := range f {
		parts[i] = num.String()
	}
	return "{" + strings.Join(parts, ", ") + "}"
}
