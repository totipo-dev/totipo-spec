// Package cryptov1 implements the fixed v1 envelope and provenance primitives.
package cryptov1

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"golang.org/x/crypto/argon2"
	"golang.org/x/crypto/hkdf"
	"io"
	"totipo/conformance/internal/object"
	"unicode/utf8"
)

var ErrCrypto = errors.New("invalid v1 cryptographic input")

type Keys struct{ ID, ObjectRoot, SignatureContext []byte }

func MAC(k, p []byte) []byte { h := hmac.New(sha256.New, k); h.Write(p); return h.Sum(nil) }
func expand(k, info []byte) []byte {
	p := make([]byte, 32)
	if _, e := io.ReadFull(hkdf.Expand(sha256.New, k, info), p); e != nil {
		panic(e)
	}
	return p
}
func Derive(root []byte) (Keys, error) {
	if len(root) != 32 {
		return Keys{}, ErrCrypto
	}
	prk := hkdf.Extract(sha256.New, root, make([]byte, 32))
	return Keys{expand(prk, []byte("totipo/v1/object-id")), expand(prk, []byte("totipo/v1/object-key-root")), expand(prk, []byte("totipo/v1/signature-context"))}, nil
}
func Binding(root []byte) ([]byte, error) {
	if len(root) != 32 {
		return nil, ErrCrypto
	}
	return MAC(root, []byte("totipo/v1/local-vault-binding")), nil
}
func (k Keys) ObjectKey(id []byte) []byte {
	return expand(k.ObjectRoot, append([]byte("totipo/v1/object-key"), id...))
}
func AAD(id []byte) []byte { return append([]byte("totipo/v1/object"), id...) }
func gcm(key []byte) (cipher.AEAD, error) {
	b, e := aes.NewCipher(key)
	if e != nil {
		return nil, e
	}
	return cipher.NewGCM(b)
}
func Padded(p []byte) ([]byte, error) {
	if len(p) > object.Capacity {
		return nil, ErrCrypto
	}
	b := make([]byte, 1008)
	binary.BigEndian.PutUint16(b, uint16(len(p)))
	copy(b[2:], p)
	return b, nil
}
func (k Keys) Seal(p []byte) (string, []byte, error) {
	plain, e := Padded(p)
	if e != nil {
		return "", nil, e
	}
	id := MAC(k.ID, p)
	a, e := gcm(k.ObjectKey(id))
	if e != nil {
		return "", nil, e
	}
	return hex.EncodeToString(id), a.Seal(nil, id[:12], plain, AAD(id)), nil
}
func (k Keys) Open(name string, b []byte) ([]byte, error) {
	if len(name) != 64 || len(b) != 1024 {
		return nil, ErrCrypto
	}
	id, e := hex.DecodeString(name)
	if e != nil || hex.EncodeToString(id) != name {
		return nil, ErrCrypto
	}
	a, e := gcm(k.ObjectKey(id))
	if e != nil {
		return nil, e
	}
	p, e := a.Open(nil, id[:12], b, AAD(id))
	if e != nil {
		return nil, ErrCrypto
	}
	n := int(binary.BigEndian.Uint16(p))
	if n > object.Capacity {
		return nil, ErrCrypto
	}
	for _, x := range p[2+n:] {
		if x != 0 {
			return nil, ErrCrypto
		}
	}
	semantic := p[2 : 2+n]
	if !hmac.Equal(MAC(k.ID, semantic), id) {
		return nil, ErrCrypto
	}
	return bytes.Clone(semantic), nil
}
func (k Keys) SignatureInput(o object.Object) ([]byte, error) {
	p, e := o.Unsigned()
	if e != nil {
		return nil, e
	}
	domain := "totipo/v1/token"
	if o.Type == object.Device {
		domain = "totipo/v1/device"
	}
	b := append([]byte(domain), k.SignatureContext...)
	return append(b, p...), nil
}
func PublicBytes(k *ecdsa.PublicKey) []byte { return elliptic.Marshal(elliptic.P256(), k.X, k.Y) }
func (k Keys) Provenance(o object.Object, public []byte) string {
	if len(o.Signature) == 0 {
		return "REJECTED"
	}
	if o.Type == object.Device {
		public = o.PublicKey
	} else if len(public) == 0 {
		return "UNRESOLVED"
	}
	want := o.Author
	if o.Type == object.Device {
		want = o.Identity
	}
	if !bytes.Equal(want, object.DeviceID(public)) {
		return "REJECTED"
	}
	x, y := elliptic.Unmarshal(elliptic.P256(), public)
	if x == nil {
		return "REJECTED"
	}
	input, e := k.SignatureInput(o)
	if e != nil {
		return "REJECTED"
	}
	h := sha256.Sum256(input)
	if !ecdsa.VerifyASN1(&ecdsa.PublicKey{Curve: elliptic.P256(), X: x, Y: y}, h[:], o.Signature) {
		return "REJECTED"
	}
	return "VERIFIED"
}

// Sign plans capacity before one signing operation and verifies locally. It
// never retries to shorten a signature or change the chosen parent set.
func (k Keys) Sign(o object.Object, private *ecdsa.PrivateKey) (object.Object, error) {
	n, e := o.ReservedSize()
	if e != nil {
		return o, e
	}
	if n > object.Capacity {
		return o, ErrCrypto
	}
	input, e := k.SignatureInput(o)
	if e != nil {
		return o, e
	}
	h := sha256.Sum256(input)
	o.Signature, e = ecdsa.SignASN1(rand.Reader, private, h[:])
	if e != nil {
		return o, e
	}
	if k.Provenance(o, PublicBytes(&private.PublicKey)) != "VERIFIED" {
		return o, ErrCrypto
	}
	return o, nil
}
func PasswordValid(password []byte) bool { return len(password) <= 1024 && utf8.Valid(password) }
func WrapKey(password, salt []byte) ([]byte, error) {
	if !PasswordValid(password) || len(salt) != 16 {
		return nil, ErrCrypto
	}
	return argon2.IDKey(password, salt, 3, 65536, 4, 32), nil
}

// Wrap accepts explicit salt/nonce to support fixtures. Application writers must
// use fresh CSPRNG values for both on every creation or rewrap (sections 8–9).
func Wrap(password, root, salt, nonce []byte) ([]byte, error) {
	if len(root) != 32 || len(salt) != 16 || len(nonce) != 12 {
		return nil, ErrCrypto
	}
	k, e := WrapKey(password, salt)
	if e != nil {
		return nil, e
	}
	a, e := gcm(k)
	if e != nil {
		return nil, e
	}
	h := append([]byte("TOTIPO-VLT\x01"), salt...)
	h = append(h, nonce...)
	return append(h, a.Seal(nil, nonce, root, h)...), nil
}
func Unwrap(password, record []byte) ([]byte, error) {
	if len(record) != 87 || !bytes.Equal(record[:11], []byte("TOTIPO-VLT\x01")) || !PasswordValid(password) {
		return nil, ErrCrypto
	}
	k, e := WrapKey(password, record[11:27])
	if e != nil {
		return nil, e
	}
	a, e := gcm(k)
	if e != nil {
		return nil, e
	}
	root, e := a.Open(nil, record[27:39], record[39:], record[:39])
	if e != nil {
		return nil, ErrCrypto
	}
	return root, nil
}
