package lfs

import (
	"math/big"
	"testing"
)

func TestSolveFCM(t *testing.T) {
	type args struct {
		n *big.Int
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "test_8",
			args: args{
				n: big.NewInt(8),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := SolveFCM(tt.args.n); !Verify(tt.args.n, got) {
				t.Errorf("SolveFCM() verify failed, got: %v != %v", got, tt.args.n)
				return
			}
		})
	}
}
