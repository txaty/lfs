package lfs

import (
	"context"
	"math/big"

	comp "github.com/txaty/go-bigcomplex"
	"lukechampine.com/frand"
)

// solveFCM implements the FCM algorithm for very large n.
func (s *Solver) solveFCM(n *big.Int) FourInt {
	// For n below the FCM threshold, fallback to the basic method.
	if n.Cmp(s.FCMThreshold) < 0 {
		return s.solveBasic(n)
	}
	nOdd, e := extractOddComponent(n)
	gcd, l := fcmRandTrail(nOdd, s.NumRoutines)
	hurwitzGCRD := fcmFinalizeHurwitzGCRD(nOdd, l, gcd)
	gi := computeGaussianOnePlusIPower(e)
	hurwitzProd := comp.NewHurwitzInt(gi.R, gi.I, big0, big0, false)
	hurwitzProd.Prod(hurwitzProd, hurwitzGCRD)
	w1, w2, w3, w4 := hurwitzProd.ValInt()
	return NewFourInt(w1, w2, w3, w4)
}

// fcmRandTrail performs a random search tailored for the FCM algorithm.
// It returns a Gaussian GCD along with the candidate l.
func fcmRandTrail(nOdd *big.Int, numRoutines int) (*comp.GaussianInt, *big.Int) {
	preP := iPool.Get().(*big.Int).Lsh(nOdd, 1) // preP = 2 * nOdd
	defer iPool.Put(preP)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	resChan := make(chan fcmFindResult)
	randLimit := iPool.Get().(*big.Int).Lsh(big1, fcmComputeRandBitLen(preP))
	defer iPool.Put(randLimit)
	for i := 0; i < numRoutines; i++ {
		go fcmWorkerFindS(ctx, randLimit, preP, resChan)
	}
	res := <-resChan
	return res.gcd, res.l
}

// fcmComputeRandBitLen computes a bit length for random candidate generation in FCM.
func fcmComputeRandBitLen(n *big.Int) uint {
	bitLen := n.BitLen()
	ret := uint(bitLen / 2)
	if ret < 10 {
		ret = 10
	}
	return ret
}

type fcmFindResult struct {
	gcd *comp.GaussianInt
	l   *big.Int
}

// fcmWorkerFindS repeatedly searches for a valid candidate in the FCM algorithm.
func fcmWorkerFindS(ctx context.Context, randLimit, preP *big.Int, resChan chan<- fcmFindResult) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
			s, p, l, ok := fcmPickCandidate(randLimit, preP)
			if !ok {
				continue
			}
			gcd := computeGaussianGCD(s, p)
			if !isValidGaussianGCD(gcd) {
				continue
			}
			select {
			case resChan <- fcmFindResult{gcd: gcd, l: l}:
				return
			default:
				return
			}
		}
	}
}

// fcmPickCandidate generates a candidate for the FCM algorithm.
func fcmPickCandidate(randLimit, preP *big.Int) (s, p, l *big.Int, found bool) {
	l = frand.BigIntn(randLimit)
	l.Lsh(l, 1)
	l.Add(l, big1) // ensure l is odd
	lSq := iPool.Get().(*big.Int).Mul(l, l)
	defer iPool.Put(lSq)
	p = new(big.Int).Sub(preP, lSq)
	if p.Sign() <= 0 {
		return nil, nil, nil, false
	}
	if !p.ProbablyPrime(0) {
		return nil, nil, nil, false
	}
	pMinus1 := iPool.Get().(*big.Int).Sub(p, big1)
	defer iPool.Put(pMinus1)
	powU := iPool.Get().(*big.Int).Rsh(pMinus1, 1)
	defer iPool.Put(powU)
	halfP := iPool.Get().(*big.Int).Rsh(p, 1)
	defer iPool.Put(halfP)
	u := iPool.Get().(*big.Int)
	defer iPool.Put(u)
	opt := iPool.Get().(*big.Int)
	defer iPool.Put(opt)
	valid := false
	for i := 0; i < maxIterFindU; i++ {
		u = frand.BigIntn(halfP)
		u.Lsh(u, 1)
		opt.Exp(u, powU, p)
		if opt.Cmp(pMinus1) == 0 {
			valid = true
			break
		}
	}
	if !valid {
		return nil, nil, nil, false
	}
	powU.Rsh(powU, 1)
	s = new(big.Int).Exp(u, powU, p)
	return s, new(big.Int).Set(p), l, true
}

// fcmFinalizeHurwitzGCRD computes the Hurwitz GCRD for the FCM algorithm.
func fcmFinalizeHurwitzGCRD(n, l *big.Int, gcd *comp.GaussianInt) *comp.HurwitzInt {
	hurwitzCandidate := hiPool.Get().(*comp.HurwitzInt).Update(gcd.R, gcd.I, l, big0, false)
	defer hiPool.Put(hurwitzCandidate)
	hurwitzN := hiPool.Get().(*comp.HurwitzInt).Update(n, big0, big0, big0, false)
	defer hiPool.Put(hurwitzN)
	gcrd := new(comp.HurwitzInt).GCRD(hurwitzCandidate, hurwitzN)
	return gcrd
}
