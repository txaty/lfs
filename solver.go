// Package lfs provides functionality to compute Lagrange four-square representations
// of positive integers using different algorithms. Both a basic method and a more
// complicated Fermat-Christmas Method (FCM) are provided.
// Settings are configurable via functional options.
package lfs

import (
	"math/big"
	"runtime"
)

// Option defines a functional option for configuring the Solver.
type Option func(*Solver)

// Solver encapsulates configuration for computing the four-square representation.
type Solver struct {
	// FCMThreshold determines when to use the FCM-based algorithm.
	// If the input n is greater than or equal to FCMThreshold, the FCM algorithm is used.
	FCMThreshold *big.Int

	// NumRoutines specifies the number of goroutines to use for parallel randomized search.
	NumRoutines int
}

// NewSolver creates a new Solver with the provided options.
// By default, FCMThreshold is set to 2^500 and NumRoutines to the number of available CPUs.
func NewSolver(opts ...Option) *Solver {
	s := &Solver{
		FCMThreshold: new(big.Int).Lsh(big1, 500),
		NumRoutines:  runtime.NumCPU(),
	}
	for _, opt := range opts {
		opt(s)
	}
	return s
}

// WithFCMThreshold configures the FCM threshold.
func WithFCMThreshold(th *big.Int) Option {
	return func(s *Solver) {
		s.FCMThreshold = th
	}
}

// WithNumRoutines configures the number of goroutines used in random search.
func WithNumRoutines(n int) Option {
	return func(s *Solver) {
		s.NumRoutines = n
	}
}

// Solve computes the Lagrange four-square representation for n.
// It automatically selects between the basic algorithm and the FCM algorithm.
func (s *Solver) Solve(n *big.Int) FourInt {
	if n.Sign() == 0 {
		// Special case: 0 = 0^2 + 0^2 + 0^2 + 0^2
		return NewFourInt(precomputedHurwitzGCRDs[0].ValInt())
	}
	if n.Cmp(s.FCMThreshold) < 0 {
		return s.solveBasic(n)
	}
	return s.solveFCM(n)
}

// SolveBasic computes the representation using the basic algorithm.
func (s *Solver) SolveBasic(n *big.Int) FourInt {
	return s.solveBasic(n)
}
