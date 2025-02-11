// solver_test.go
package lfs

import (
	"math/big"
	"reflect"
	"runtime"
	"testing"
)

func TestNewSolver(t *testing.T) {
	defaultThreshold := new(big.Int).Lsh(big.NewInt(1), 500)
	defaultNumRoutines := runtime.NumCPU()

	tests := []struct {
		name string
		opts []Option
		want *Solver
	}{
		{
			name: "default options",
			opts: nil,
			want: &Solver{
				FCMThreshold: defaultThreshold,
				NumRoutines:  defaultNumRoutines,
			},
		},
		{
			name: "custom options",
			opts: []Option{
				WithFCMThreshold(new(big.Int).Lsh(big.NewInt(1), 600)), // 2^600
				WithNumRoutines(8),
			},
			want: &Solver{
				FCMThreshold: new(big.Int).Lsh(big.NewInt(1), 600),
				NumRoutines:  8,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewSolver(tt.opts...)
			// Compare NumRoutines.
			if got.NumRoutines != tt.want.NumRoutines {
				t.Errorf("NewSolver() NumRoutines = %d, want %d", got.NumRoutines, tt.want.NumRoutines)
			}
			// Compare FCMThreshold via Cmp since they are big.Int pointers.
			if got.FCMThreshold.Cmp(tt.want.FCMThreshold) != 0 {
				t.Errorf("NewSolver() FCMThreshold = %v, want %v", got.FCMThreshold, tt.want.FCMThreshold)
			}
		})
	}
}

func TestSolver_Solve(t *testing.T) {
	tests := []struct {
		name   string
		fields struct {
			FCMThreshold *big.Int
			NumRoutines  int
		}
		args struct {
			n *big.Int
		}
		want FourInt
	}{
		{
			name: "solve zero",
			fields: struct {
				FCMThreshold *big.Int
				NumRoutines  int
			}{
				FCMThreshold: new(big.Int).Lsh(big.NewInt(1), 500),
				NumRoutines:  runtime.NumCPU(),
			},
			args: struct{ n *big.Int }{
				n: big.NewInt(0),
			},
			want: NewFourInt(big.NewInt(0), big.NewInt(0), big.NewInt(0), big.NewInt(0)),
		},
		{
			name: "solve four",
			fields: struct {
				FCMThreshold *big.Int
				NumRoutines  int
			}{
				FCMThreshold: new(big.Int).Lsh(big.NewInt(1), 500),
				NumRoutines:  4,
			},
			args: struct{ n *big.Int }{
				n: big.NewInt(4),
			},
			want: NewFourInt(big.NewInt(2), big.NewInt(0), big.NewInt(0), big.NewInt(0)),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Solver{
				FCMThreshold: tt.fields.FCMThreshold,
				NumRoutines:  tt.fields.NumRoutines,
			}
			got := s.Solve(tt.args.n)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Solve() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSolver_SolveBasic(t *testing.T) {
	tests := []struct {
		name   string
		fields struct {
			FCMThreshold *big.Int
			NumRoutines  int
		}
		args struct {
			n *big.Int
		}
		want FourInt
	}{
		{
			name: "solve basic four",
			fields: struct {
				FCMThreshold *big.Int
				NumRoutines  int
			}{
				FCMThreshold: new(big.Int).Lsh(big.NewInt(1), 500),
				NumRoutines:  4,
			},
			args: struct{ n *big.Int }{
				n: big.NewInt(4),
			},
			want: NewFourInt(big.NewInt(2), big.NewInt(0), big.NewInt(0), big.NewInt(0)),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Solver{
				FCMThreshold: tt.fields.FCMThreshold,
				NumRoutines:  tt.fields.NumRoutines,
			}
			got := s.SolveBasic(tt.args.n)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("SolveBasic() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestWithFCMThreshold(t *testing.T) {
	customThreshold := new(big.Int).Lsh(big.NewInt(1), 600)
	opt := WithFCMThreshold(customThreshold)
	s := &Solver{
		FCMThreshold: new(big.Int).Lsh(big.NewInt(1), 500),
		NumRoutines:  runtime.NumCPU(),
	}
	opt(s)
	if s.FCMThreshold.Cmp(customThreshold) != 0 {
		t.Errorf("WithFCMThreshold() did not set threshold correctly: got %v, want %v", s.FCMThreshold, customThreshold)
	}
}

func TestWithNumRoutines(t *testing.T) {
	opt := WithNumRoutines(8)
	s := &Solver{
		FCMThreshold: new(big.Int).Lsh(big.NewInt(1), 500),
		NumRoutines:  runtime.NumCPU(),
	}
	opt(s)
	if s.NumRoutines != 8 {
		t.Errorf("WithNumRoutines() did not set NumRoutines correctly: got %d, want %d", s.NumRoutines, 8)
	}
}
