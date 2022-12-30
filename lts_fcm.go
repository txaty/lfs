package lfs

import (
	"context"
	"math/big"

	comp "github.com/txaty/go-bigcomplex"
	"lukechampine.com/frand"
)

var (
	thresholdFCM = new(big.Int).Lsh(big1, 500)
)

// SolveFCM finds the Lagrange four square solution for a very large integer.
// It uses Fermat's Christmas Theorem one more time to further reduce the integer size.
func SolveFCM(n *big.Int, numRoutine int) FourInt {
	if n.Cmp(thresholdFCM) < 0 {
		return Solve(n, numRoutine)
	}

	nc, e := divideN(n)
	gcd, l := randTrailFCM(nc, numRoutine)
	hurwitzGCRD := denouementFCM(nc, l, gcd)

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

func randTrailFCM(nc *big.Int, numRoutine int) (*comp.GaussianInt, *big.Int) {
	preP := iPool.Get().(*big.Int).Lsh(nc, 1)
	defer iPool.Put(preP)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	resChan := make(chan findResultFCM)
	randLmt := iPool.Get().(*big.Int).Lsh(big1, setRandBitLenFCM(preP))
	defer iPool.Put(randLmt)
	for i := 0; i < numRoutine; i++ {
		go routineFindSfcm(ctx, randLmt, preP, resChan)
	}
	res := <-resChan
	return res.gcd, res.l
}

func setRandBitLenFCM(n *big.Int) uint {
	bitLen := n.BitLen()
	ret := uint(float32(bitLen) / 2)
	if ret < 10 {
		ret = 10
	}
	return ret
}

type findResultFCM struct {
	gcd *comp.GaussianInt
	l   *big.Int
}

func routineFindSfcm(ctx context.Context, randLmt, preP *big.Int, resChan chan<- findResultFCM) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
			s, p, l, ok := pickSfcm(randLmt, preP)
			if !ok {
				continue
			}
			gcd := gaussianIntGCD(s, p)
			if !isValidGaussianIntGCD(gcd) {
				continue
			}
			ctx.Done()
			select {
			case resChan <- findResultFCM{gcd: gcd, l: l}:
				return
			default:
				return
			}
		}
	}
}

func pickSfcm(randLmt, preP *big.Int) (s, p, l *big.Int, found bool) {
	l = frand.BigIntn(randLmt)
	l.Lsh(l, 1)
	l.Add(l, big1)
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
	for i := 0; i < maxFindUIter; i++ {
		u = frand.BigIntn(halfP)
		u.Lsh(u, 1)

		// test if s^2 = -1 (mod p)
		// if so, continue to the next step, otherwise, repeat this step
		opt.Exp(u, powU, p)
		if opt.Cmp(pMinus1) == 0 {
			found = true
			break
		}
	}
	if !found {
		return nil, nil, nil, false
	}

	// compute s = u^((p - 1) / 4) mod p
	powU.Rsh(powU, 1)
	s = new(big.Int).Exp(u, powU, p)
	return
}

func denouementFCM(n, l *big.Int, gcd *comp.GaussianInt) *comp.HurwitzInt {
	// compute gcrd(A + Bi + Lj, n), normalized to have integer component
	// Hurwitz integer: A + Bi + Lj
	hurwitzInt := hiPool.Get().(*comp.HurwitzInt).Update(gcd.R, gcd.I, l, big0, false)
	defer hiPool.Put(hurwitzInt)
	// Hurwitz integer: n
	hurwitzN := hiPool.Get().(*comp.HurwitzInt).Update(n, big0, big0, big0, false)
	defer hiPool.Put(hurwitzN)
	gcrd := new(comp.HurwitzInt).GCRD(hurwitzInt, hurwitzN)

	return gcrd
}
