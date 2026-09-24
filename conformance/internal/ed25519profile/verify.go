// Package ed25519profile enforces Totipo r36 Section 16 on public inputs.
package ed25519profile

import (
	"bytes"
	"crypto/ed25519"
	"errors"
	"filippo.io/edwards25519"
)

var order = []byte{0xed, 0xd3, 0xf5, 0x5c, 0x1a, 0x63, 0x12, 0x58, 0xd6, 0x9c, 0xf7, 0xa2, 0xde, 0xf9, 0xde, 0x14, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0x10}

// fullOrder multiplies by the integer L using point addition. No scalar API
// may reduce L modulo L: doing so would make every point pass vacuously.
func fullOrder(p *edwards25519.Point) *edwards25519.Point {
	q := edwards25519.NewIdentityPoint()
	for i := 255; i >= 0; i-- {
		q.Add(q, q)
		if order[i/8]&(1<<uint(i%8)) != 0 {
			q.Add(q, p)
		}
	}
	return q
}
func small(p *edwards25519.Point) bool {
	q := new(edwards25519.Point).Set(p)
	for i := 0; i < 3; i++ {
		q.Add(q, q)
	}
	return q.Equal(edwards25519.NewIdentityPoint()) == 1
}
func Verify(publicKey, message, signature []byte) error {
	reject := func(s string) error { return errors.New(s) }
	if len(publicKey) != 32 {
		return reject("A_LENGTH")
	}
	if len(signature) != 64 {
		return reject("SIGNATURE_LENGTH")
	}
	a, e := new(edwards25519.Point).SetBytes(publicKey)
	if e != nil {
		return reject("A_DECODE")
	}
	r, e := new(edwards25519.Point).SetBytes(signature[:32])
	if e != nil {
		return reject("R_DECODE")
	}
	if !bytes.Equal(a.Bytes(), publicKey) {
		return reject("A_CANONICAL")
	}
	if !bytes.Equal(r.Bytes(), signature[:32]) {
		return reject("R_CANONICAL")
	}
	if small(a) {
		return reject("A_SMALL_ORDER")
	}
	if small(r) {
		return reject("R_SMALL_ORDER")
	}
	identity := edwards25519.NewIdentityPoint()
	if fullOrder(a).Equal(identity) != 1 {
		return reject("A_SUBGROUP")
	}
	if fullOrder(r).Equal(identity) != 1 {
		return reject("R_SUBGROUP")
	}
	if _, e := new(edwards25519.Scalar).SetCanonicalBytes(signature[32:]); e != nil {
		return reject("S_RANGE")
	}
	// With canonical non-small prime-subgroup A/R and S<L, the standard
	// PureEd25519 equation is exactly the required cofactorless equation.
	if !ed25519.Verify(publicKey, message, signature) {
		return reject("EQUATION")
	}
	return nil
}
