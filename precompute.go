package lfs

import (
	"math/big"

	comp "github.com/txaty/go-bigcomplex"
)

const precomputeLmt = 20

var (
	// precomputed Hurwitz GCRDs for small integers
	precomputedHurwitzGCRDs = [precomputeLmt + 1]*comp.HurwitzInt{
		comp.NewHurwitzInt(big0, big0, big0, big0, false), // 0: 0, 0, 0, 0
		comp.NewHurwitzInt(big1, big0, big0, big0, false), // 1: 1, 0, 0, 0
		comp.NewHurwitzInt(big1, big1, big0, big0, false), // 2: 1, 1, 0, 0
		comp.NewHurwitzInt(big1, big1, big1, big0, false), // 3: 1, 1, 1, 0
		comp.NewHurwitzInt(big2, big0, big0, big0, false), // 4: 2, 0, 0, 0
		comp.NewHurwitzInt(big2, big1, big0, big0, false), // 5: 2, 1, 0, 0
		comp.NewHurwitzInt(big2, big1, big1, big0, false), // 6: 2, 1, 1, 0
		comp.NewHurwitzInt(big2, big1, big1, big1, false), // 7: 2, 1, 1, 1
		comp.NewHurwitzInt(big2, big2, big0, big0, false), // 8: 2, 2, 0, 0
		comp.NewHurwitzInt(big2, big2, big1, big0, false), // 9: 2, 2, 1, 0
		comp.NewHurwitzInt(big2, big2, big1, big1, false), // 10: 2, 2, 1, 1
		comp.NewHurwitzInt(big3, big1, big1, big0, false), // 11: 3, 1, 1, 0
		comp.NewHurwitzInt(big3, big1, big1, big1, false), // 12: 3, 1, 1, 1
		comp.NewHurwitzInt(big3, big2, big0, big0, false), // 13: 3, 2, 0, 0
		comp.NewHurwitzInt(big3, big2, big1, big0, false), // 14: 3, 2, 1, 0
		comp.NewHurwitzInt(big3, big2, big1, big1, false), // 15: 3, 2, 1, 1
		comp.NewHurwitzInt(big4, big0, big0, big0, false), // 16: 4, 0, 0, 0
		comp.NewHurwitzInt(big4, big1, big0, big0, false), // 17: 4, 1, 0, 0
		comp.NewHurwitzInt(big4, big1, big1, big0, false), // 18: 4, 1, 1, 0
		comp.NewHurwitzInt(big4, big1, big1, big1, false), // 19: 4, 1, 1, 1
		comp.NewHurwitzInt(big4, big2, big0, big0, false), // 20: 4, 2, 0, 0
	}
	bigPrecomputeLmt = big.NewInt(precomputeLmt)
	tinyPrimeProd    = big.NewInt(210) // 2 * 3 * 5 * 7
)

func log2(n *big.Int) int {
	return n.BitLen() - 1
}
