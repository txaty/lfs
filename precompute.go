package lfs

import (
	"math/big"

	comp "github.com/txaty/go-bigcomplex"
)

const precomputeLmt = 20

var (
	// precomputedHurwitzGCRDs contains precomputed Hurwitz GCRDs for small integers.
	precomputedHurwitzGCRDs = [precomputeLmt + 1]*comp.HurwitzInt{
		comp.NewHurwitzInt(big0, big0, big0, big0, false), // 0
		comp.NewHurwitzInt(big1, big0, big0, big0, false), // 1
		comp.NewHurwitzInt(big1, big1, big0, big0, false), // 2
		comp.NewHurwitzInt(big1, big1, big1, big0, false), // 3
		comp.NewHurwitzInt(big2, big0, big0, big0, false), // 4
		comp.NewHurwitzInt(big2, big1, big0, big0, false), // 5
		comp.NewHurwitzInt(big2, big1, big1, big0, false), // 6
		comp.NewHurwitzInt(big2, big1, big1, big1, false), // 7
		comp.NewHurwitzInt(big2, big2, big0, big0, false), // 8
		comp.NewHurwitzInt(big2, big2, big1, big0, false), // 9
		comp.NewHurwitzInt(big2, big2, big1, big1, false), // 10
		comp.NewHurwitzInt(big3, big1, big1, big0, false), // 11
		comp.NewHurwitzInt(big3, big1, big1, big1, false), // 12
		comp.NewHurwitzInt(big3, big2, big0, big0, false), // 13
		comp.NewHurwitzInt(big3, big2, big1, big0, false), // 14
		comp.NewHurwitzInt(big3, big2, big1, big1, false), // 15
		comp.NewHurwitzInt(big4, big0, big0, big0, false), // 16
		comp.NewHurwitzInt(big4, big1, big0, big0, false), // 17
		comp.NewHurwitzInt(big4, big1, big1, big0, false), // 18
		comp.NewHurwitzInt(big4, big1, big1, big1, false), // 19
		comp.NewHurwitzInt(big4, big2, big0, big0, false), // 20
	}
	bigPrecomputeLmt = big.NewInt(precomputeLmt)
	tinyPrimeProd    = big.NewInt(210) // Product of small primes: 2*3*5*7
)

// log2 returns the floor of the base‑2 logarithm of n.
func log2(n *big.Int) int {
	return n.BitLen() - 1
}
