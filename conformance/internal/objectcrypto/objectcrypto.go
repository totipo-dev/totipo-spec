package objectcrypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/sha256"
	"errors"
	"totipo/conformance/internal/dispatch"
	"totipo/conformance/internal/envelope"
)

const domain = "TOTP-Vault/v0/"

type Keys struct{ PRK, ID, ObjectRoot, SignatureContext []byte }

func MAC(key, p []byte) []byte { h := hmac.New(sha256.New, key); h.Write(p); return h.Sum(nil) }

// Expand32 is exactly the first (and only) HKDF-SHA-256 Expand block.
func Expand32(key, info []byte) []byte { return MAC(key, append(append([]byte{}, info...), 1)) }
func Derive(root []byte) (Keys, error) {
	if len(root) != 32 {
		return Keys{}, errors.New("root must be 32 bytes")
	}
	prk := MAC(make([]byte, 32), root)
	return Keys{prk, Expand32(prk, []byte(domain+"object-id")), Expand32(prk, []byte(domain+"object-key-root")), Expand32(prk, []byte(domain+"signature-context"))}, nil
}
func SignatureDomain(typ byte) ([]byte, error) {
	if typ == 1 {
		return []byte(domain + "token-update"), nil
	}
	if typ == 2 {
		return []byte(domain + "device-update"), nil
	}
	return nil, errors.New("unknown type")
}
func (k Keys) SignatureInput(typ byte, unsigned []byte) ([]byte, error) {
	d, e := SignatureDomain(typ)
	if e != nil {
		return nil, e
	}
	d = append(d, k.SignatureContext...)
	return append(d, unsigned...), nil
}
func (k Keys) Parameters(id []byte) (key, nonce, aad []byte, err error) {
	if len(id) != 32 {
		return nil, nil, nil, errors.New("ID must be 32 bytes")
	}
	key = Expand32(k.ObjectRoot, append([]byte(domain+"object-key"), id...))
	return key, id[:12], append([]byte(domain+"object"), id...), nil
}
func (k Keys) AEAD(id []byte) (cipher.AEAD, []byte, []byte, error) {
	key, n, a, e := k.Parameters(id)
	if e != nil {
		return nil, nil, nil, e
	}
	block, e := aes.NewCipher(key)
	if e != nil {
		return nil, nil, nil, e
	}
	g, e := cipher.NewGCM(block)
	return g, n, a, e
}
func (k Keys) Seal(p []byte) (id, file []byte, err error) {
	framed, e := envelope.Frame(p)
	if e != nil {
		return nil, nil, e
	}
	id = MAC(k.ID, p)
	g, n, a, e := k.AEAD(id)
	if e != nil {
		return nil, nil, e
	}
	return id, g.Seal(nil, n, framed, a), nil
}

// Open reports a rejection stage and never interprets semantic data prematurely.
// The optional observer is called immediately before each attempted stage.
func (k Keys) Open(id, file []byte, observe func(string)) ([]byte, string) {
	hit := func(s string) {
		if observe != nil {
			observe(s)
		}
	}
	hit("FILE_LENGTH")
	if len(file) != envelope.FileSize {
		return nil, "FILE_LENGTH_REJECTED"
	}
	g, n, a, e := k.AEAD(id)
	if e != nil {
		return nil, "OBJECT_ID_REJECTED"
	}
	hit("AEAD")
	b, e := g.Open(nil, n, file, a)
	if e != nil {
		return nil, "AEAD_REJECTED"
	}
	hit("ENVELOPE")
	p, e := envelope.Unframe(b)
	if e != nil {
		return nil, "ENVELOPE_REJECTED"
	}
	hit("OBJECT_ID")
	if !hmac.Equal(MAC(k.ID, p), id) {
		return nil, "OBJECT_ID_REJECTED"
	}
	hit("DISPATCH")
	return p, dispatch.Classify(p)
}
