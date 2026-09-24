package totp

import (
	"crypto/hmac"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/binary"
	"errors"
	"fmt"
	"hash"
	"math/big"
)

func Generate(secret []byte, algorithm, digits int, period uint64, seconds *big.Int) (string, error) {
	if len(secret) < 1 || len(secret) > 128 || digits < 6 || digits > 8 || period == 0 || period > 0xffffffff || seconds == nil || seconds.Sign() < 0 {
		return "", errors.New("invalid TOTP input")
	}
	counter := new(big.Int).Quo(seconds, new(big.Int).SetUint64(period))
	if counter.BitLen() > 64 {
		return "", errors.New("counter overflow")
	}
	var hf func() hash.Hash
	switch algorithm {
	case 1:
		hf = sha1.New
	case 2:
		hf = sha256.New
	case 3:
		hf = sha512.New
	default:
		return "", errors.New("algorithm")
	}
	var b [8]byte
	binary.BigEndian.PutUint64(b[:], counter.Uint64())
	h := hmac.New(hf, secret)
	h.Write(b[:])
	sum := h.Sum(nil)
	off := int(sum[len(sum)-1] & 15)
	code := binary.BigEndian.Uint32(sum[off:off+4]) & 0x7fffffff
	mod := uint32(1)
	for i := 0; i < digits; i++ {
		mod *= 10
	}
	return fmt.Sprintf("%0*d", digits, code%mod), nil
}
