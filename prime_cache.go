package lfs

import (
	"math"
	"math/big"
	"sync"
)

type primeCache struct {
	l   []int    // list of prime numbers
	m   sync.Map // map from a prime to the product of primes up to that prime
	max int      // largest prime in the cache
}

// findPrimeProd returns the product of primes less than logN.
func (p *primeCache) findPrimeProd(logN int) *big.Int {
	l, r := 0, len(p.l)-1
	for l <= r {
		mid := (l + r) / 2
		current := p.l[mid]
		if mid == len(p.l)-1 {
			if bi, ok := p.m.Load(current); ok {
				return new(big.Int).Set(bi.(*big.Int))
			}
		}
		next := p.l[mid+1]
		if current < logN && next >= logN {
			if bi, ok := p.m.Load(current); ok {
				return new(big.Int).Set(bi.(*big.Int))
			}
		}
		if current >= logN {
			r = mid - 1
		} else {
			l = mid + 1
		}
	}
	return big.NewInt(2)
}

func newPrimeCache(limit int) *primeCache {
	ps := &primeCache{
		l:   []int{1, 2, 3, 5, 7},
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
	for idx := 9; idx <= limit; idx += 2 {
		ps.checkAddPrime(idx, prod, opt)
	}
	return ps
}

func (p *primeCache) checkAddPrime(n int, prod, opt *big.Int) {
	isPrime := true
	sqrtN := int(math.Sqrt(float64(n)))
	for _, prime := range p.l[1:] { // skip 1
		if prime > sqrtN {
			break
		}
		if n%prime == 0 && n != prime {
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

// ResetCachePrime resets the prime cache.
func ResetCachePrime() {
	pCache = newPrimeCache(0)
}
