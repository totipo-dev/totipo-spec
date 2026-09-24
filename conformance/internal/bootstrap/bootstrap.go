package bootstrap

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"errors"
	"golang.org/x/crypto/argon2"
	"unicode/utf8"
)

const Magic = "TOTP-VAULT"

// KDF is injectable so tests can prove that preflight failures perform no work.
type KDF func(password, salt []byte) []byte

func Derive(password, salt []byte) []byte { return argon2.IDKey(password, salt, 3, 65536, 4, 32) }
func Password(p []byte) error {
	if len(p) > 1024 || !utf8.Valid(p) {
		return errors.New("PASSWORD_REJECTED")
	}
	return nil
}

// EncodeRunes rejects invalid source code points before Go could replace them.
func EncodeRunes(r []rune) ([]byte, error) {
	for _, c := range r {
		if !utf8.ValidRune(c) {
			return nil, errors.New("PASSWORD_REJECTED")
		}
	}
	p := []byte(string(r))
	return p, Password(p)
}
func Check(file, password []byte) string {
	if len(file) != 87 {
		return "FILE_LENGTH_REJECTED"
	}
	if !bytes.Equal(file[:10], []byte(Magic)) {
		return "MAGIC_REJECTED"
	}
	if file[10] != 0 {
		return "BOOTSTRAP_VERSION_REJECTED"
	}
	if Password(password) != nil {
		return "PASSWORD_REJECTED"
	}
	return "PREFLIGHT_VALID"
}
func gcm(key []byte) (cipher.AEAD, error) {
	b, e := aes.NewCipher(key)
	if e != nil {
		return nil, e
	}
	return cipher.NewGCM(b)
}
func Wrap(root, password, salt, nonce []byte, kdf KDF) (file, key []byte, err error) {
	if Password(password) != nil {
		return nil, nil, errors.New("PASSWORD_REJECTED")
	}
	if len(root) != 32 || len(salt) != 16 || len(nonce) != 12 {
		return nil, nil, errors.New("invalid wrap inputs")
	}
	if kdf == nil {
		kdf = Derive
	}
	key = kdf(password, salt)
	g, e := gcm(key)
	if e != nil {
		return nil, nil, e
	}
	header := append([]byte(Magic), 0)
	header = append(header, salt...)
	header = append(header, nonce...)
	return append(header, g.Seal(nil, nonce, root, header)...), key, nil
}
func Unwrap(file, password []byte, kdf KDF) (root []byte, stage string) {
	if s := Check(file, password); s != "PREFLIGHT_VALID" {
		return nil, s
	}
	if kdf == nil {
		kdf = Derive
	}
	key := kdf(password, file[11:27])
	g, e := gcm(key)
	if e != nil {
		return nil, "AEAD_REJECTED"
	}
	root, e = g.Open(nil, file[27:39], file[39:], file[:39])
	if e != nil {
		return nil, "AEAD_REJECTED"
	}
	return root, "VERIFIED"
}
