package lfs

import (
	"context"
	comp "github.com/txaty/go-bigcomplex"
	"log"
	"lukechampine.com/frand"
	"math"
	"math/big"
	"runtime"
)

const (
	randLmtThreshold = 16
	maxFindUIter     = 10
)

// Solve calculates the Lagrange four squares representation of a positive integer
// The algorithm is modified from the paper "Finding the Four Squares in Lagrange’s Theorem"
func Solve(n *big.Int, numRoutine int) FourInt {
	if n.Sign() == 0 {
		return NewFourInt(precomputedHurwitzGCRDs[0].ValInt())
	}
	if numRoutine <= 0 {
		numRoutine = runtime.NumCPU()
	}
	nc, e := divideN(n)
	var hurwitzGCRD *comp.HurwitzInt

	if nc.Cmp(bigPrecomputeLmt) <= 0 {
		hurwitzGCRD = precomputedHurwitzGCRDs[nc.Int64()]
	} else {
		var gcd *comp.GaussianInt
		nBitLen := nc.BitLen()
		if nBitLen < randLmtThreshold {
			gcd = randTrail(nc, precompute(nc), numRoutine)
		} else {
			gcd = randTrailLarge(nc, nBitLen, numRoutine)
		}
		hurwitzGCRD = denouement(nc, gcd)
	}

	// if x'^2 + Y'^2 + Z'^2 + W'^2 = n'
	// then x^2 + Y^2 + Z^2 + W^2 = n for x, Y, Z, W defined by
	// (1 + i)^e * (x' + Y'i + Z'j + W'k) = (x + Yi + Zj + Wk)
	// Gaussian integer: 1 + i
	gi := gaussian1PlusIPow(e)
	hurwitzProd := comp.NewHurwitzInt(gi.R, gi.I, big0, big0, false)
	hurwitzProd.Prod(hurwitzProd, hurwitzGCRD)
	w1, w2, w3, w4 := hurwitzProd.ValInt()
	return NewFourInt(w1, w2, w3, w4)
}

func isValidGaussianIntGCD(gcd *comp.GaussianInt) bool {
	if gcd == nil {
		return false
	}
	absR := iPool.Get().(*big.Int).Abs(gcd.R)
	defer iPool.Put(absR)
	absI := iPool.Get().(*big.Int).Abs(gcd.I)
	defer iPool.Put(absI)
	rCmp1 := absR.Cmp(big1)
	rSign := absR.Sign()
	iCmp1 := absI.Cmp(big1)
	iSign := absI.Sign()
	if rCmp1 == 0 && iSign == 0 {
		return false
	}
	if rSign == 0 && iCmp1 == 0 {
		return false
	}
	if rCmp1 == 0 && iCmp1 == 0 {
		return false
	}
	return true
}

func divideN(n *big.Int) (*big.Int, int) {
	// n = 2^e * n', n' is odd
	nc := new(big.Int).Set(n)
	var e int
	for nc.Bit(0) == 0 {
		nc.Rsh(nc, 1)
		e++
	}
	return nc, e
}

// gaussian1PlusIPow calculates Gaussian integer (1 + i)^e
func gaussian1PlusIPow(e int) *comp.GaussianInt {
	if e == 0 {
		return comp.NewGaussianInt(big1, big0)
	}
	if gi, ok := giCache.Load(e); ok {
		return gi.(*comp.GaussianInt)
	}
	gaussian1PlusI := giPool.Get().(*comp.GaussianInt).Update(big1, big1)
	defer giPool.Put(gaussian1PlusI)

	gaussianProd := comp.NewGaussianInt(big1, big0)
	idx := e
	for idx > 0 {
		gaussianProd.Prod(gaussianProd, gaussian1PlusI)
		idx--
	}
	gi := new(comp.GaussianInt).Update(gaussianProd.R, gaussianProd.I)
	giCache.Store(e, gi)
	return gaussianProd
}

// precompute determine the primes not exceeding log n and compute their product
// the function only handles positive integers larger than precomputed range (20)
func precompute(n *big.Int) *big.Int {
	if n.Cmp(bigPrecomputeLmt) <= 0 {
		log.Panicf("n should be larger than %d", precomputeLmt)
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

func randTrail(n, primeProd *big.Int, numRoutine int) *comp.GaussianInt {
	// use goroutines to choose a random number between [0, n^5 / 2 / numRoutine]
	// then construct k based on the random number
	// and check the validity of the trails
	// p = M * n * k - 1, pre-p = M * n
	preP := iPool.Get().(*big.Int).Mul(primeProd, n)
	defer iPool.Put(preP)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	resChan := make(chan *comp.GaussianInt)
	randLmt := setInitRandLmt(n)
	randLmt.Rsh(randLmt, 1)
	randLmt.Div(randLmt, big.NewInt(int64(numRoutine)))
	//randLmt.Add(randLmt, big1)

	mul := iPool.Get().(*big.Int).SetInt64(int64(2 * numRoutine)) // 2 * numRoutine
	defer iPool.Put(mul)
	var adds []*big.Int
	for i := 0; i <= numRoutine; i++ {
		adds = append(adds, big.NewInt(int64(2*i+1))) // 2i+1
	}
	for _, add := range adds {
		go routineFindS(ctx, add, mul, randLmt, preP, resChan)
	}
	return <-resChan
}

func randTrailLarge(n *big.Int, bitLen, numRoutine int) *comp.GaussianInt {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	resChan := make(chan *comp.GaussianInt)
	bl := setInitRandBitLen(bitLen)
	preP := iPool.Get().(*big.Int).Mul(tinyPrimeProd, n)
	defer iPool.Put(preP)
	randLmt := iPool.Get().(*big.Int).Lsh(big1, uint(bl))
	defer iPool.Put(randLmt)
	for i := 0; i < numRoutine; i++ {
		go routineFindSLarge(ctx, randLmt, preP, resChan)
	}
	return <-resChan
}

func setInitRandLmt(n *big.Int) *big.Int {
	bitLen := n.BitLen()
	exp := iPool.Get().(*big.Int).SetInt64(4)
	defer iPool.Put(exp)
	bitLen >>= 2 // bitLen / 4
	for bitLen > 1 {
		exp.Sub(exp, big1)
		bitLen >>= 1
	}
	return new(big.Int).Exp(n, exp, nil)
}

func setInitRandBitLen(bitLen int) int {
	lenF := 20 + 2*math.Log(float64(bitLen))
	return int(math.Round(lenF))
}

func routineFindS(ctx context.Context, mul, add, randLmt, preP *big.Int, resChan chan<- *comp.GaussianInt) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
			s, p, ok, err := pickS(mul, add, randLmt, preP)
			if err != nil {
				panic(err)
			}
			if !ok {
				continue
			}
			gcd := gaussianIntGCD(s, p)
			if !isValidGaussianIntGCD(gcd) {
				continue
			}
			ctx.Done()
			select {
			case resChan <- gcd:
				return
			default:
				return
			}
		}
	}
}

func pickS(mul, add, randLmt, preP *big.Int) (*big.Int, *big.Int, bool, error) {
	// choose k' in [0, randLmt)
	k := frand.BigIntn(randLmt)
	// construct k, k = k' * mul + add
	k.Mul(k, mul)
	k.Add(k, add)
	return determineSAndP(k, preP)
}

func routineFindSLarge(ctx context.Context, randLmt, preP *big.Int, resChan chan<- *comp.GaussianInt) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
			s, p, ok, err := pickSLarge(randLmt, preP)
			if err != nil {
				panic(err)
			}
			if !ok {
				continue
			}
			gcd := gaussianIntGCD(s, p)
			if !isValidGaussianIntGCD(gcd) {
				continue
			}
			ctx.Done()
			select {
			case resChan <- gcd:
				return
			default:
				return
			}
		}
	}
}

func pickSLarge(randLmt, preP *big.Int) (*big.Int, *big.Int, bool, error) {
	k := frand.BigIntn(randLmt)
	k.Or(k, big1)
	return determineSAndP(k, preP)
}

func determineSAndP(k, preP *big.Int) (*big.Int, *big.Int, bool, error) {
	// p = {Product of primes} * n * k - 1 = preP * k - 1
	p := iPool.Get().(*big.Int).Mul(preP, k)
	defer iPool.Put(p)
	p.Sub(p, big1)

	// we want to find a prime number p,
	// so perform probably_prime checking to reject number which is not prime potentially,
	// quick restart if p cannot pass Baillie-PSW test
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

	// choose u from [1, p - 1]
	// here we can pick u in [0, p)
	// if u is 0, then the accepting condition will not pass
	// use normal rand source to prevent acquiring crypto rand reader mutex
	// to reduce the probability of picking up a prime number, we only choose even numbers
	findValidU := false
	for i := 0; i < maxFindUIter; i++ {
		u = frand.BigIntn(halfP)
		u.Lsh(u, 1)

		// test if s^2 = -1 (mod p)
		// if so, continue to the next step, otherwise, repeat this step
		opt.Exp(u, powU, p)
		if opt.Cmp(pMinus1) == 0 {
			findValidU = true
			//return nil, nil, false, nil
			break
		}
	}
	if !findValidU {
		return nil, nil, false, nil
	}

	// compute s = u^((p - 1) / 4) mod p
	powU.Rsh(powU, 1)
	s := new(big.Int).Exp(u, powU, p)
	return s, new(big.Int).Set(p), true, nil
}

func gaussianIntGCD(s, p *big.Int) *comp.GaussianInt {
	// compute A + Bi := gcd(s + i, p)
	// Gaussian integer: s + i
	gaussianInt := giPool.Get().(*comp.GaussianInt).Update(s, big1)
	defer giPool.Put(gaussianInt)
	// Gaussian integer: p
	gaussianP := giPool.Get().(*comp.GaussianInt).Update(p, big0)
	defer giPool.Put(gaussianP)
	// compute gcd(s + i, p)
	gcd := new(comp.GaussianInt)
	gcd.GCD(gaussianInt, gaussianP)
	return gcd
}

func denouement(n *big.Int, gcd *comp.GaussianInt) *comp.HurwitzInt {
	// compute gcrd(A + Bi + j, n), normalized to have integer component
	// Hurwitz integer: A + Bi + j
	hurwitzInt := hiPool.Get().(*comp.HurwitzInt).Update(gcd.R, gcd.I, big1, big0, false)
	defer hiPool.Put(hurwitzInt)
	// Hurwitz integer: n
	hurwitzN := hiPool.Get().(*comp.HurwitzInt).Update(n, big0, big0, big0, false)
	defer hiPool.Put(hurwitzN)
	gcrd := new(comp.HurwitzInt).GCRD(hurwitzInt, hurwitzN)

	return gcrd
}

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
