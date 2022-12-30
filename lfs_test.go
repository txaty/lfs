package lfs

import (
	"math/big"
	"reflect"
	"testing"
)

var (
	big2Pow20 big.Int
	big2Pow32 big.Int
)

func setup() {
	var ok bool
	if _, ok = big2Pow20.SetString("1048576", 10); !ok {
		panic("failed to set big2Pow20")
	}
	if _, ok = big2Pow32.SetString("4294967296", 10); !ok {
		panic("failed to set big2Pow32")
	}
}

func Test_precompute(t *testing.T) {
	type args struct {
		n *big.Int
	}
	tests := []struct {
		name string
		args args
		want *big.Int
	}{
		{
			name: "test_2^20",
			args: args{
				n: &big2Pow20,
			},
			want: big.NewInt(9699690),
		},
		{
			name: "test_2^32",
			args: args{
				n: &big2Pow32,
			},
			want: big.NewInt(200560490130),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setup()
			if got := precompute(tt.args.n); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("precompute() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestVerify(t *testing.T) {
	type args struct {
		target *big.Int
		w1     *big.Int
		w2     *big.Int
		w3     *big.Int
		w4     *big.Int
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "test_verify_success_4",
			args: args{
				target: big.NewInt(4),
				w1:     big.NewInt(2),
				w2:     big.NewInt(0),
				w3:     big.NewInt(0),
				w4:     big.NewInt(0),
			},
			want: true,
		},
		{
			name: "test_verify_success_35955023",
			args: args{
				target: big.NewInt(35955023),
				w1:     big.NewInt(2323),
				w2:     big.NewInt(5454),
				w3:     big.NewInt(893),
				w4:     big.NewInt(123),
			},
			want: true,
		},
		{
			name: "test_verify_fail_35955024",
			args: args{
				target: big.NewInt(35955024),
				w1:     big.NewInt(2323),
				w2:     big.NewInt(5454),
				w3:     big.NewInt(893),
				w4:     big.NewInt(123),
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fs := [4]*big.Int{
				tt.args.w1,
				tt.args.w2,
				tt.args.w3,
				tt.args.w4,
			}
			if got := Verify(tt.args.target, fs); got != tt.want {
				t.Errorf("Verify() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSolve(t *testing.T) {
	type args struct {
		n *big.Int
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "test_4",
			args: args{
				n: big.NewInt(4),
			},
		},
		{
			name: "test_35955023",
			args: args{
				n: big.NewInt(35955023),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Solve(tt.args.n); !Verify(tt.args.n, got) {
				t.Errorf("Solve() verify failed, got: %v != %v", got, tt.args.n)
			}
		})
	}
}
