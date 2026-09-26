package totp

import (
	"strings"
	"testing"
)

func TestBoundaryAndLeadingZero(t *testing.T) {
	secret := []byte("12345678901234567890")
	for _, v := range []struct {
		time   uint64
		digits byte
		want   string
	}{{59, 8, "94287082"}, {30, 8, "94287082"}, {60, 8, "37359152"}, {1111111109, 8, "07081804"}, {1111111109, 7, "7081804"}, {1111111109, 6, "081804"}} {
		got, e := Code(1, secret, v.digits, 30, v.time)
		if e != nil || got != v.want {
			t.Fatalf("time %d: got %q, err %v", v.time, got, e)
		}
	}
}
func TestInvalidCredential(t *testing.T) {
	for _, v := range []struct {
		algorithm, digits byte
		period            uint32
		secret            string
	}{{0, 8, 30, "s"}, {4, 8, 30, "s"}, {1, 5, 30, "s"}, {1, 9, 30, "s"}, {1, 8, 0, "s"}, {1, 8, 30, ""}, {1, 8, 30, strings.Repeat("s", 129)}} {
		if _, e := Code(v.algorithm, []byte(v.secret), v.digits, v.period, 59); e == nil {
			t.Fatal("invalid credential accepted")
		}
	}
}
func TestUnsignedTimeDomain(t *testing.T) {
	for _, period := range []uint32{1, ^uint32(0)} {
		got, e := Code(3, []byte("s"), 8, period, ^uint64(0))
		if e != nil || len(got) != 8 {
			t.Fatal("unsigned time domain", e)
		}
	}
}
