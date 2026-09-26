// Package totp implements RFC 6238 with the credential bounds in v1 section 40.
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
)

// Code uses T0=0, unsigned Unix seconds, and an eight-byte big-endian counter.
// Its caller must establish credential-use eligibility before using the result.
func Code(algorithm byte, secret []byte, digits byte, period uint32, unixSeconds uint64) (string, error) {
	if digits < 6 || digits > 8 || period == 0 || len(secret) < 1 || len(secret) > 128 {
		return "", errors.New("invalid TOTP credential")
	}
	var newHash func() hash.Hash
	switch algorithm {
	case 1:
		newHash = sha1.New
	case 2:
		newHash = sha256.New
	case 3:
		newHash = sha512.New
	default:
		return "", errors.New("unknown TOTP algorithm")
	}
	var counter [8]byte
	binary.BigEndian.PutUint64(counter[:], unixSeconds/uint64(period))
	mac := hmac.New(newHash, secret)
	mac.Write(counter[:])
	sum := mac.Sum(nil)
	offset := int(sum[len(sum)-1] & 15)
	value := binary.BigEndian.Uint32(sum[offset:offset+4]) & 0x7fffffff
	modulus := uint32(1)
	for i := byte(0); i < digits; i++ {
		modulus *= 10
	}
	return fmt.Sprintf("%0*d", int(digits), value%modulus), nil
}
