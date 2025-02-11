package lfs

import (
	"math/big"
	"sync"

	comp "github.com/txaty/go-bigcomplex"
)

const primeCacheLimit = 16

var (
	// Common big.Int constants.
	big0 = big.NewInt(0)
	big1 = big.NewInt(1)
	big2 = big.NewInt(2)
	big3 = big.NewInt(3)
	big4 = big.NewInt(4)

	// iPool is a sync.Pool for big.Int to reduce GC pressure.
	iPool = sync.Pool{
		New: func() interface{} { return new(big.Int) },
	}

	// giPool is a sync.Pool for Gaussian integers.
	giPool = sync.Pool{
		New: func() interface{} { return new(comp.GaussianInt) },
	}

	// hiPool is a sync.Pool for Hurwitz integers.
	hiPool = sync.Pool{
		New: func() interface{} { return new(comp.HurwitzInt) },
	}

	// pCache caches prime numbers and their products.
	pCache = newPrimeCache(primeCacheLimit)

	// giCache caches computed Gaussian integers (for example, (1+i)^e).
	giCache = sync.Map{}
)

// ResetCacheGaussianInt resets the Gaussian integer cache.
func ResetCacheGaussianInt() {
	giCache = sync.Map{}
}

// CacheGaussianInt precomputes and caches (1+i)^n for n <= e.
func CacheGaussianInt(e int) {
	giCache = sync.Map{}
	gaussianProd := giPool.Get().(*comp.GaussianInt).Update(big1, big0)
	defer giPool.Put(gaussianProd)
	for i := 0; i <= e; i++ {
		giCache.Store(i, gaussianProd.Copy())
		gaussianProd.Prod(gaussianProd, gaussianProd)
	}
}
