package lfs

import (
	"context"
	"log"
	"math"
	"math/big"

	comp "github.com/txaty/go-bigcomplex"
	"lukechampine.com/frand"
)

const (
	randLimitThreshold = 16
	maxIterFindU       = 10
)

// solveBasic implements the basic Lagrange four‐square solution algorithm.
func (s *Solver) solveBasic(n *big.Int) FourInt {
	// Factor out powers of 2: n = 2^e * nOdd, with nOdd odd.
	nOdd, e := extractOddComponent(n)

	var hurwitzGCRD *comp.HurwitzInt
	if nOdd.Cmp(bigPrecomputeLmt) <= 0 {
		// For small nOdd, use a precomputed Hurwitz GCRD.
		hurwitzGCRD = precomputedHurwitzGCRDs[nOdd.Int64()]
	} else {
		// Otherwise, use a randomized trail search.
		var gaussianGCD *comp.GaussianInt
		if nOdd.BitLen() < randLimitThreshold {
			gaussianGCD = findGaussianGCDSmall(nOdd, computePrimeProduct(nOdd), s.NumRoutines)
		} else {
			gaussianGCD = findGaussianGCDLarge(nOdd, nOdd.BitLen(), s.NumRoutines)
		}
		hurwitzGCRD = finalizeHurwitzGCRD(nOdd, gaussianGCD)
	}

	// Adjust the solution using (1+i)^e.
	gi := computeGaussianOnePlusIPower(e)
	hurwitzProd := comp.NewHurwitzInt(gi.R, gi.I, big0, big0, false)
	hurwitzProd.Prod(hurwitzProd, hurwitzGCRD)
	w1, w2, w3, w4 := hurwitzProd.ValInt()
	return NewFourInt(w1, w2, w3, w4)
}

// extractOddComponent factors n as n = 2^e * nOdd (with nOdd odd).
func extractOddComponent(n *big.Int) (*big.Int, int) {
	nOdd := new(big.Int).Set(n)
	e := 0
	for nOdd.Bit(0) == 0 {
		nOdd.Rsh(nOdd, 1)
		e++
	}
	return nOdd, e
}

// computeGaussianOnePlusIPower computes (1+i)^e using exponentiation by squaring.
// It caches the result for efficiency.
func computeGaussianOnePlusIPower(e int) *comp.GaussianInt {
	if e == 0 {
		return comp.NewGaussianInt(big1, big0)
	}
	if cached, ok := giCache.Load(e); ok {
		return cached.(*comp.GaussianInt)
	}
	base := comp.NewGaussianInt(big1, big1)
	result := comp.NewGaussianInt(big1, big0)
	exp := e
	for exp > 0 {
		if exp&1 == 1 {
			result.Prod(result, base)
		}
		base.Prod(base, base)
		exp >>= 1
	}
	gi := new(comp.GaussianInt).Update(result.R, result.I)
	giCache.Store(e, gi)
	return result
}

// computePrimeProduct computes the product of primes (up to log2(n)) using the prime cache.
func computePrimeProduct(n *big.Int) *big.Int {
	if n.Cmp(bigPrecomputeLmt) <= 0 {
		return big.NewInt(1)
	}
	logN := log2(n)
	if logN <= pCache.max {
		return pCache.findPrimeProd(logN)
	}
	pm, _ := pCache.m.Load(pCache.max)
	prod := iPool.Get().(*big.Int).Set(pm.(*big.Int))
	defer iPool.Put(prod)
	opt := iPool.Get().(*big.Int)
	defer iPool.Put(opt)
	for idx := pCache.max + 2; idx < logN; idx += 2 {
		pCache.checkAddPrime(idx, prod, opt)
	}
	return new(big.Int).Set(prod)
}

// findGaussianGCDSmall performs random search for a valid Gaussian GCD for small nOdd.
func findGaussianGCDSmall(n, primeProd *big.Int, numRoutines int) *comp.GaussianInt {
	preP := iPool.Get().(*big.Int).Mul(primeProd, n)
	defer iPool.Put(preP)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	resChan := make(chan *comp.GaussianInt)
	randLimit := computeInitialRandLimit(n)
	randLimit.Rsh(randLimit, 1)
	randLimit.Div(randLimit, big.NewInt(int64(numRoutines)))

	mul := iPool.Get().(*big.Int).SetInt64(int64(2 * numRoutines))
	defer iPool.Put(mul)
	var offsets []*big.Int
	for i := 0; i <= numRoutines; i++ {
		offsets = append(offsets, big.NewInt(int64(2*i+1)))
	}
	for _, off := range offsets {
		go workerFindS(ctx, mul, off, randLimit, preP, resChan)
	}
	return <-resChan
}

// findGaussianGCDLarge performs random search for a valid Gaussian GCD for large nOdd.
func findGaussianGCDLarge(n *big.Int, bitLen, numRoutines int) *comp.GaussianInt {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	resChan := make(chan *comp.GaussianInt)
	bl := computeRandBitLength(bitLen)
	preP := iPool.Get().(*big.Int).Mul(tinyPrimeProd, n)
	defer iPool.Put(preP)
	randLimit := iPool.Get().(*big.Int).Lsh(big1, uint(bl))
	defer iPool.Put(randLimit)
	for i := 0; i < numRoutines; i++ {
		go workerFindSLarge(ctx, randLimit, preP, resChan)
	}
	return <-resChan
}

// computeInitialRandLimit computes an initial random limit for candidate generation.
func computeInitialRandLimit(n *big.Int) *big.Int {
	bitLen := n.BitLen()
	exp := iPool.Get().(*big.Int).SetInt64(4)
	defer iPool.Put(exp)
	bitLen >>= 2
	for bitLen > 1 {
		exp.Sub(exp, big1)
		bitLen >>= 1
	}
	return new(big.Int).Exp(n, exp, nil)
}

// computeRandBitLength computes a bit length for random candidate generation.
func computeRandBitLength(bitLen int) int {
	lenF := 20 + 2*math.Log(float64(bitLen))
	return int(math.Round(lenF))
}

// workerFindS is a goroutine that repeatedly searches for a valid candidate.
func workerFindS(ctx context.Context, mul, offset, randLimit, preP *big.Int, resChan chan<- *comp.GaussianInt) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
			s, p, ok, err := pickCandidateS(mul, offset, randLimit, preP)
			if err != nil {
				log.Panic(err)
			}
			if !ok {
				continue
			}
			gcd := computeGaussianGCD(s, p)
			if !isValidGaussianGCD(gcd) {
				continue
			}
			select {
			case resChan <- gcd:
				return
			default:
				return
			}
		}
	}
}

// pickCandidateS generates candidate s and p for workerFindS.
func pickCandidateS(mul, offset, randLimit, preP *big.Int) (*big.Int, *big.Int, bool, error) {
	k := frand.BigIntn(randLimit)
	k.Mul(k, mul)
	k.Add(k, offset)
	return computeCandidateSP(k, preP)
}

// workerFindSLarge is the worker routine for large nOdd.
func workerFindSLarge(ctx context.Context, randLimit, preP *big.Int, resChan chan<- *comp.GaussianInt) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
			s, p, ok, err := pickCandidateSLarge(randLimit, preP)
			if err != nil {
				log.Panic(err)
			}
			if !ok {
				continue
			}
			gcd := computeGaussianGCD(s, p)
			if !isValidGaussianGCD(gcd) {
				continue
			}
			select {
			case resChan <- gcd:
				return
			default:
				return
			}
		}
	}
}

// pickCandidateSLarge generates candidate s and p for large nOdd.
func pickCandidateSLarge(randLimit, preP *big.Int) (*big.Int, *big.Int, bool, error) {
	k := frand.BigIntn(randLimit)
	k.Or(k, big1)
	return computeCandidateSP(k, preP)
}

// computeCandidateSP computes candidate s and p given k and preP.
func computeCandidateSP(k, preP *big.Int) (*big.Int, *big.Int, bool, error) {
	p := iPool.Get().(*big.Int).Mul(preP, k)
	defer iPool.Put(p)
	p.Sub(p, big1)
	if !p.ProbablyPrime(0) {
		return nil, nil, false, nil
	}
	pMinus1 := iPool.Get().(*big.Int).Sub(p, big1)
	defer iPool.Put(pMinus1)
	powU := iPool.Get().(*big.Int).Rsh(pMinus1, 1)
	defer iPool.Put(powU)
	halfP := iPool.Get().(*big.Int).Rsh(p, 1)
	defer iPool.Put(halfP)
	opt := iPool.Get().(*big.Int)
	defer iPool.Put(opt)
	u := iPool.Get().(*big.Int)
	defer iPool.Put(u)
	found := false
	for i := 0; i < maxIterFindU; i++ {
		u = frand.BigIntn(halfP)
		u.Lsh(u, 1)
		opt.Exp(u, powU, p)
		if opt.Cmp(pMinus1) == 0 {
			found = true
			break
		}
	}
	if !found {
		return nil, nil, false, nil
	}
	powU.Rsh(powU, 1)
	s := new(big.Int).Exp(u, powU, p)
	return s, new(big.Int).Set(p), true, nil
}

// computeGaussianGCD computes the Gaussian GCD of (s+i) and p.
func computeGaussianGCD(s, p *big.Int) *comp.GaussianInt {
	gaussS := giPool.Get().(*comp.GaussianInt).Update(s, big1)
	defer giPool.Put(gaussS)
	gaussP := giPool.Get().(*comp.GaussianInt).Update(p, big0)
	defer giPool.Put(gaussP)
	gcd := new(comp.GaussianInt)
	gcd.GCD(gaussS, gaussP)
	return gcd
}

// isValidGaussianGCD verifies that the Gaussian GCD is nontrivial.
func isValidGaussianGCD(gcd *comp.GaussianInt) bool {
	if gcd == nil {
		return false
	}
	absR := iPool.Get().(*big.Int).Abs(gcd.R)
	defer iPool.Put(absR)
	absI := iPool.Get().(*big.Int).Abs(gcd.I)
	defer iPool.Put(absI)
	// Reject trivial cases.
	if absR.Cmp(big1) == 0 && absI.Sign() == 0 {
		return false
	}
	if absR.Sign() == 0 && absI.Cmp(big1) == 0 {
		return false
	}
	if absR.Cmp(big1) == 0 && absI.Cmp(big1) == 0 {
		return false
	}
	return true
}

// finalizeHurwitzGCRD computes the Hurwitz GCRD of (gcd + j) and n.
func finalizeHurwitzGCRD(n *big.Int, gcd *comp.GaussianInt) *comp.HurwitzInt {
	hurwitzCandidate := hiPool.Get().(*comp.HurwitzInt).Update(gcd.R, gcd.I, big1, big0, false)
	defer hiPool.Put(hurwitzCandidate)
	hurwitzN := hiPool.Get().(*comp.HurwitzInt).Update(n, big0, big0, big0, false)
	defer hiPool.Put(hurwitzN)
	gcrd := new(comp.HurwitzInt).GCRD(hurwitzCandidate, hurwitzN)
	return gcrd
}
