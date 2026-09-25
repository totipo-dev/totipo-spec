package cryptov1

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"encoding/asn1"
	"encoding/binary"
	"encoding/hex"
	"math/big"
	"testing"
	"totipo/conformance/internal/object"
)

func TestEnvelopeRejections(t *testing.T) {
	k, _ := Derive(make([]byte, 32))
	semantic := []byte{1, 2, 3}
	name, b, e := k.Seal(semantic)
	if e != nil {
		t.Fatal(e)
	}
	for _, n := range []string{name[:63], "A" + name[1:], "../" + name, name + ".tmp"} {
		if _, e := k.Open(n, b); e == nil {
			t.Fatal("invalid name accepted")
		}
	}
	for _, length := range []int{0, 1023, 1025} {
		if _, e := k.Open(name, make([]byte, length)); e == nil {
			t.Fatal("length accepted")
		}
	}
	bad := bytes.Clone(b)
	bad[0] ^= 1
	if _, e := k.Open(name, bad); e == nil {
		t.Fatal("AEAD accepted tamper")
	}
	id, _ := hex.DecodeString(name)
	a, _ := gcm(k.ObjectKey(id))
	plain, _ := Padded(semantic)
	for _, mutate := range []func([]byte){func(p []byte) { p[1007] = 1 }, func(p []byte) { binary.BigEndian.PutUint16(p, 1007) }, func(p []byte) { p[2] ^= 1 }} {
		p := bytes.Clone(plain)
		mutate(p)
		b := a.Seal(nil, id[:12], p, AAD(id))
		if _, e := k.Open(name, b); e == nil {
			t.Fatal("authenticated noncanonical plaintext accepted")
		}
	}
	other, _ := Derive(bytes.Repeat([]byte{1}, 32))
	if _, e := other.Open(name, b); e == nil {
		t.Fatal("cross-vault accepted")
	}
}
func TestBootstrapRejectsBeforeKDF(t *testing.T) {
	for _, password := range [][]byte{{0xff}, bytes.Repeat([]byte{'a'}, 1025)} {
		if _, e := WrapKey(password, make([]byte, 16)); e == nil {
			t.Fatal("password")
		}
	}
	for _, record := range [][]byte{nil, make([]byte, 87), append([]byte("TOTIPO-VLT\x02"), make([]byte, 76)...)} {
		if _, e := Unwrap(nil, record); e == nil {
			t.Fatal("bootstrap header")
		}
	}
}
func TestBootstrapAuthentication(t *testing.T) {
	r := bytes.Repeat([]byte{1}, 32)
	b, e := Wrap([]byte("p"), r, make([]byte, 16), make([]byte, 12))
	if e != nil {
		t.Fatal(e)
	}
	if _, e = Unwrap([]byte("wrong"), b); e == nil {
		t.Fatal("wrong password")
	}
	b[86] ^= 1
	if _, e = Unwrap([]byte("p"), b); e == nil {
		t.Fatal("tag tamper")
	}
}
func TestSigningAndProvenance(t *testing.T) {
	priv, e := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if e != nil {
		t.Fatal(e)
	}
	pub := PublicBytes(&priv.PublicKey)
	k, _ := Derive(make([]byte, 32))
	o := object.Object{Routing: object.Routing{Version: 1, Type: object.Token, Identity: make([]byte, 32), Author: object.DeviceID(pub)}, Status: 1, Algorithm: 1, Digits: 6, Period: 30, Secret: []byte{1}}
	o, e = k.Sign(o, priv)
	if e != nil {
		t.Fatal(e)
	}
	if k.Provenance(o, pub) != "VERIFIED" {
		t.Fatal("signature")
	}
	var pair struct{ R, S *big.Int }
	rest, e := asn1.Unmarshal(o.Signature, &pair)
	if e != nil || len(rest) != 0 {
		t.Fatal("DER fixture")
	}
	pair.S.Sub(priv.Params().N, pair.S)
	o.Signature, e = asn1.Marshal(pair)
	if e != nil {
		t.Fatal(e)
	}
	if k.Provenance(o, pub) != "VERIFIED" {
		t.Fatal("high-S/low-S equivalent rejected")
	}
	good := bytes.Clone(o.Signature)
	o.Signature = append(o.Signature, 0)
	if k.Provenance(o, pub) != "REJECTED" {
		t.Fatal("noncanonical trailing DER accepted")
	}
	o.Signature = good

	if k.Provenance(o, nil) != "UNRESOLVED" {
		t.Fatal("missing key")
	}
	o.AuthorTime++
	if k.Provenance(o, pub) != "REJECTED" {
		t.Fatal("timestamp not signed")
	}
	o.Signature = nil
	if k.Provenance(o, nil) != "REJECTED" {
		t.Fatal("empty signature")
	}
}
func FuzzOpen(f *testing.F) {
	k, _ := Derive(make([]byte, 32))
	name, b, _ := k.Seal([]byte{1, 2, 3})
	f.Add(name, b)
	f.Add("", []byte{})
	f.Fuzz(func(t *testing.T, name string, b []byte) {
		p, e := k.Open(name, b)
		if e == nil {
			n, again, e := k.Seal(p)
			if e != nil || n != name || !bytes.Equal(b, again) {
				t.Fatal("noncanonical envelope")
			}
		}
	})
}
