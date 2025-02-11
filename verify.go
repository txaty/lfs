package lfs

import "math/big"

// Verify checks if the four-square sum is equal to the original integer
// i.e. target = w1^2 + w2^2 + w3^2 + w4^2
func Verify(target *big.Int, fi FourInt) bool {
	sum := iPool.Get().(*big.Int).SetInt64(0)
	defer iPool.Put(sum)
	opt := iPool.Get().(*big.Int)
	defer iPool.Put(opt)
	for i := 0; i < 4; i++ {
		sum.Add(sum, opt.Mul(fi[i], fi[i]))
	}
	return sum.Cmp(target) == 0
}
