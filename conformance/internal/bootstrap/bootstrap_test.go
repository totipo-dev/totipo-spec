package bootstrap

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"testing"
	"totipo/conformance/internal/corpus"
)

func TestEveryVersionRejectedBeforeKDF(t *testing.T) {
	file := make([]byte, 87)
	copy(file, Magic)
	for v := 1; v < 256; v++ {
		file[10] = byte(v)
		_, s := Unwrap(file, nil, func(p, s []byte) []byte { t.Fatal("KDF called"); return nil })
		if s != "BOOTSTRAP_VERSION_REJECTED" {
			t.Fatal(s)
		}
	}
}
func TestPasswordBoundsAndNoNormalization(t *testing.T) {
	root := make([]byte, 32)
	salt := make([]byte, 16)
	nonce := make([]byte, 12)
	for _, p := range [][]byte{nil, bytes.Repeat([]byte("\U0001f600"), 256)} {
		file, _, e := Wrap(root, p, salt, nonce, nil)
		if e != nil {
			t.Fatal(e)
		}
		got, s := Unwrap(file, p, nil)
		if s != "VERIFIED" || !bytes.Equal(got, root) {
			t.Fatal(s)
		}
	}
	a := []byte("é")
	b := []byte("e\u0301")
	if bytes.Equal(Derive(a, salt), Derive(b, salt)) {
		t.Fatal("passwords normalized")
	}
	if _, e := EncodeRunes([]rune{0xd800}); e == nil {
		t.Fatal("surrogate accepted")
	}
}
func TestAADBoundary(t *testing.T) {
	cases, e := corpus.Load("../../../vectors/v0")
	if e != nil {
		t.Fatal(e)
	}
	for _, c := range cases {
		if c.Kind != "bootstrap" {
			continue
		}
		file, _ := corpus.Hex(c.Expected["vault_file_hex"])
		key, _ := corpus.Hex(c.Expected["k_wrap_hex"])
		block, e := aes.NewCipher(key)
		if e != nil {
			t.Fatal(e)
		}
		g, e := cipher.NewGCM(block)
		if e != nil {
			t.Fatal(e)
		}
		for _, n := range []int{0, 10, 11, 27, 38, 40, 71, 87} {
			if _, e := g.Open(nil, file[27:39], file[39:], file[:n]); e == nil {
				t.Fatalf("%s accepted AAD length %d", c.ID, n)
			}
		}
	}
}
