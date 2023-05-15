package lfs

import (
	"math/big"
	"testing"
)

func TestFCMSolve(t *testing.T) {
	type args struct {
		n          *big.Int
		numRoutine int
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "test_8",
			args: args{
				n:          big.NewInt(8),
				numRoutine: 4,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := FCMSolve(tt.args.n, tt.args.numRoutine); !Verify(tt.args.n, got) {
				t.Errorf("FCMSolve() verify failed, got: %v != %v", got, tt.args.n)
				return
			}
		})
	}
}
