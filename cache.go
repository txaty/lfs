package lfs

import (
	"math"
	"math/big"
	"sync"

	comp "github.com/txaty/go-bigcomplex"
)

var (
	big0 = big.NewInt(0)
	big1 = big.NewInt(1)
	big2 = big.NewInt(2)
	big3 = big.NewInt(3)
	big4 = big.NewInt(4)

	// sync pool for big integers, lease GC and improve performance
	iPool = sync.Pool{
		New: func() interface{} { return new(big.Int) },
	}
	// sync pool for Gaussian integers
	giPool = sync.Pool{
		New: func() interface{} { return new(comp.GaussianInt) },
	}
	// sync pool for Hurwitz integers
	hiPool = sync.Pool{
		New: func() interface{} { return new(comp.HurwitzInt) },
	}
	// cache for precomputed prime numbers and prime number products
	pCache = newPrimeCache(16)
	// cache for computed Gaussian integers
	giCache = sync.Map{}
)

// ResetCachePrime resets the prime cache
func ResetCachePrime() {
	pCache = newPrimeCache(0)
}

type primeCache struct {
	l   []int    // list of prime numbers
	m   sync.Map // map of prime numbers and the products
	max int      // the largest prime number included
}

// findPrimeProd finds the product of primes less than log n using binary search
func (p *primeCache) findPrimeProd(logN int) *big.Int {
	var (
		l int
		r = len(p.l) - 1
	)
	for l <= r {
		mid := (l-r)/2 + r
		pll := p.l[mid]
		if mid == len(p.l)-1 {
			bi, _ := p.m.Load(pll)
			return bi.(*big.Int)
		}
		plr := p.l[mid+1]
		if pll < logN && plr >= logN {
			bi, _ := p.m.Load(pll)
			return bi.(*big.Int)
		}
		if pll >= logN {
			r = mid - 1
		} else {
			l = mid + 1
		}
	}
	return big.NewInt(2)
}

func newPrimeCache(lmt int) *primeCache {
	ps := &primeCache{
		l:   []int{1, 2, 3, 5, 7},
		m:   sync.Map{},
		max: 7,
	}
	ps.m.Store(1, big.NewInt(1))
	ps.m.Store(2, big.NewInt(2))
	ps.m.Store(3, big.NewInt(6))
	ps.m.Store(5, big.NewInt(30))
	ps.m.Store(7, big.NewInt(210))

	prod := iPool.Get().(*big.Int).SetInt64(210)
	defer iPool.Put(prod)
	opt := iPool.Get().(*big.Int)
	defer iPool.Put(opt)
	for idx := 9; idx <= lmt; idx += 2 {
		ps.checkAddPrime(idx, prod, opt)
	}
	return ps
}

func (p *primeCache) checkAddPrime(n int, prod, opt *big.Int) {
	isPrime := true
	sqrtN := int(math.Sqrt(float64(n)))
	for idx := 1; idx < len(p.l); idx++ {
		if sqrtN < p.l[idx] {
			break
		}
		if n%p.l[idx] == 0 && n != p.l[idx] {
			isPrime = false
			break
		}
	}
	if !isPrime {
		return
	}
	p.l = append(p.l, n)
	opt.SetInt64(int64(n))
	prod.Mul(prod, opt)
	p.m.Store(n, new(big.Int).Set(prod))
	p.max = n
}

// ResetCacheGaussianInt resets the Gaussian integer cache
func ResetCacheGaussianInt() {
	giCache = sync.Map{}
}

// CacheGaussianInt caches (1+i)^n, n <= e
func CacheGaussianInt(e int) {
	giCache = sync.Map{}
	gaussianProd := giPool.Get().(*comp.GaussianInt).Update(big1, big0)
	defer giPool.Put(gaussianProd)
	for i := 0; i <= e; i++ {
		giCache.Store(i, gaussianProd.Copy())
		gaussianProd.Prod(gaussianProd, gaussianProd)
	}
}
