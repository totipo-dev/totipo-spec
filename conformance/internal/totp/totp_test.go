package totp

import (
	"math/big"
	"testing"
)

func TestUnsignedCounter(t *testing.T) {
	max := new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), 64), big.NewInt(1))
	for _, period := range []uint64{1, 30, 0xffffffff} {
		seconds := new(big.Int).Mul(max, new(big.Int).SetUint64(period))
		code, e := Generate([]byte("12345678901234567890"), 1, 6, period, seconds)
		if e != nil || len(code) != 6 {
			t.Fatalf("maximum: %s %v", code, e)
		}
		seconds.Add(seconds, new(big.Int).SetUint64(period))
		if _, e := Generate([]byte("12345678901234567890"), 1, 6, period, seconds); e == nil {
			t.Fatal("overflow accepted")
		}
	}
}
