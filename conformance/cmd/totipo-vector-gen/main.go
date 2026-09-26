// totipo-vector-gen builds reviewable fixtures with the same reference codec.
// Existing signature fixtures are reused, so regeneration is byte stable. It is
// not an independent interoperability consumer and must not be counted as one.
package main

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"encoding/asn1"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"math/big"
	"os"
	"path/filepath"
	"strings"
	"totipo/conformance/internal/cryptov1"
	"totipo/conformance/internal/object"
	"totipo/conformance/internal/tlv"
	"totipo/conformance/internal/vectors"
)

var rootDir string
var root = bytes.Repeat([]byte{0x11}, 32)
var keys, _ = cryptov1.Derive(root)
var private = func() *ecdsa.PrivateKey {
	x, y := elliptic.P256().ScalarBaseMult([]byte{1})
	return &ecdsa.PrivateKey{PublicKey: ecdsa.PublicKey{Curve: elliptic.P256(), X: x, Y: y}, D: big.NewInt(1)}
}()
var public = cryptov1.PublicBytes(&private.PublicKey)
var deviceID = object.DeviceID(public)
var manifest = vectors.Manifest{Format: "totipo-vector-manifest-v1", Protocol: "totipo-v1", Revision: "r12"}

func must(e error) {
	if e != nil {
		panic(e)
	}
}
func hx(b []byte) string  { return hex.EncodeToString(b) }
func raw(s string) []byte { p, e := hex.DecodeString(s); must(e); return p }
func jsonBytes(v any) []byte {
	p, e := json.MarshalIndent(v, "", "  ")
	must(e)
	return append(p, '\n')
}
func put(path string, v any) {
	full := filepath.Join(rootDir, path)
	must(os.MkdirAll(filepath.Dir(full), 0755))
	must(os.WriteFile(full, jsonBytes(v), 0644))
}
func casePath(id string) string { return "cases/" + strings.Split(id, ".")[1] + "/" + id + ".json" }
func add(c vectors.Case, kind string, sections ...string) {
	c.Format = "totipo-case-v1"
	p := casePath(c.ID)
	put("vectors/"+p, c)
	manifest.Cases = append(manifest.Cases, vectors.Entry{ID: c.ID, Category: strings.Split(c.ID, ".")[1], Kind: kind, Normative: true, Path: p, Expected: c.Expected, Sections: sections, SHA256: vectors.Hash(jsonBytes(c))})
}
func token() object.Object {
	return object.Object{Routing: object.Routing{Version: 1, Type: object.Token, Identity: bytes.Repeat([]byte{0x22}, 32), Author: deviceID}, Status: 1, Issuer: "Example", Account: "alice@example.test", Algorithm: 1, Digits: 6, Period: 30, Secret: []byte("12345678901234567890")}
}
func device() object.Object {
	return object.Object{Routing: object.Routing{Version: 1, Type: object.Device, Identity: deviceID}, PublicKey: public, DisplayName: "Fixture device"}
}
func parents(n int) [][]byte {
	out := [][]byte{}
	for i := 0; i < n; i++ {
		b := make([]byte, 32)
		b[31] = byte(i + 1)
		out = append(out, b)
	}
	return out
}
func encode(o object.Object) []byte { p, e := o.Encode(); must(e); return p }
func signed(id string, o object.Object) object.Object {
	// Reuse the committed fixed signature only when it still verifies for these
	// exact fields. Fail on input drift instead of silently replacing evidence.
	p, e := os.ReadFile(filepath.Join(rootDir, "vectors", casePath(id)))
	if e == nil {
		var c vectors.Case
		must(vectors.Decode(p, &c))
		if c.Crypto == nil {
			panic("missing fixed signature")
		}
		o.Signature = raw(c.Crypto.Signature)
		if keys.Provenance(o, public) != "VERIFIED" {
			panic("fixture inputs changed; explicit fixture review required")
		}
		return o
	}
	if !os.IsNotExist(e) {
		must(e)
	}
	o, e = keys.Sign(o, private)
	must(e)
	return o
}
func envelope(id string, p []byte, input *object.Object, signedFixture bool, expected, kind string, sections ...string) {
	oid, b, e := keys.Seal(p)
	must(e)
	idBytes := raw(oid)
	padded, e := cryptov1.Padded(p)
	must(e)
	x := &vectors.Crypto{ObjectID: oid, IDKey: hx(keys.ID), ObjectRootKey: hx(keys.ObjectRoot), SignatureContext: hx(keys.SignatureContext), ObjectKey: hx(keys.ObjectKey(idBytes)), Nonce: hx(idBytes[:12]), AAD: hx(cryptov1.AAD(idBytes)), SemanticLength: len(p), Padded: hx(padded), Ciphertext: hx(b[:1008]), Tag: hx(b[1008:]), Object: hx(b)}
	op := "dispatch"
	if signedFixture {
		op = "crypto"
		x.PrivateKey = fmt.Sprintf("%064x", private.D)
		x.PublicKey = hx(public)
		u, e := input.Unsigned()
		must(e)
		x.Unsigned = hx(u)
		s, e := keys.SignatureInput(*input)
		must(e)
		x.SignatureInput = hx(s)
		x.Signature = hx(input.Signature)
	}
	c := vectors.Case{ID: id, Operation: op, Expected: expected, Input: input, Semantic: hx(p), Root: hx(root), Crypto: x}
	if expected == object.Opaque {
		r, tail, e := object.ParseRouting(p)
		must(e)
		c.Future = &vectors.Future{Routing: r, Tail: hx(tail)}
	}
	add(c, kind, sections...)
}
func replace(p []byte, tag uint16, value []byte) []byte {
	var out []byte
	for len(p) > 0 {
		f, r, e := tlv.Take(p)
		must(e)
		if f.Tag == tag {
			f.Value = value
		}
		b, e := tlv.Encode(f.Tag, f.Value)
		must(e)
		out = append(out, b...)
		p = r
	}
	return out
}
func initial() {
	t, d := token(), device()
	for _, v := range []struct {
		id string
		o  object.Object
	}{{"v1.routing.token-v1.001", t}, {"v1.routing.device-v1.001", d}, {"v1.encoding.token-root.001", t}, {"v1.encoding.device-root.001", d}, {"v1.device.explicit-id.001", d}} {
		o := v.o
		envelope(v.id, encode(o), &o, false, object.Supported, "bytes", "12", "20", "41–44")
	}
	for _, v := range []struct {
		id string
		o  object.Object
	}{{"v1.routing.token-future-opaque.001", t}, {"v1.routing.device-future-opaque.001", d}, {"v1.crypto.future-token-opaque.001", t}, {"v1.crypto.future-device-opaque.001", d}} {
		o := v.o
		o.Version = 2
		p, e := o.Routing.Encode()
		must(e)
		p = append(p, 0xff, 0x00, 0xff, 0xfe, 0x80)
		envelope(v.id, p, nil, false, object.Opaque, "bytes", "12", "14–15", "37")
	}
	unknown := replace(encode(t), 2, []byte{99})
	envelope("v1.routing.unknown-type-unscoped.001", unknown, nil, false, object.Unscoped, "bytes", "12", "15")
	for _, v := range []struct {
		id  string
		o   object.Object
		tag uint16
	}{{"v1.routing.future-token-malformed-prefix.001", t, 0x101}, {"v1.routing.future-device-malformed-prefix.001", d, 0x200}} {
		p := replace(encode(v.o), 1, []byte{2})
		p = replace(p, v.tag, []byte{1})
		envelope(v.id, p, nil, false, object.Unscoped, "negative", "12", "15")
	}
	t.Parents = parents(2)
	p := encode(t)
	p = replace(p, 5, bytes.Repeat([]byte{1}, 32))
	envelope("v1.encoding.parent-order.001", p, nil, false, object.Invalid, "negative", "43–44")
	p = replace(encode(t), 4, tlv.U16(1))
	envelope("v1.encoding.parent-count-mismatch.001", p, nil, false, object.Invalid, "negative", "43–44")
	p = encode(token())
	p = append(p[:10], append(bytes.Clone(p[5:10]), p[10:]...)...)
	envelope("v1.encoding.duplicate-nonrepeatable.001", p, nil, false, object.Invalid, "negative", "41–44")
	p = replace(encode(d), 0x200, make([]byte, 32))
	envelope("v1.device.id-mismatch.001", p, nil, false, object.Invalid, "negative", "20", "21.1")
	for _, v := range []struct {
		id string
		n  uint64
	}{{"v1.encoding.author-time-zero.001", 0}, {"v1.encoding.author-time-u64max.001", ^uint64(0)}, {"v1.timestamp.zero.001", 0}, {"v1.timestamp.normal.001", 1700000000}, {"v1.timestamp.i64max.001", 1<<63 - 1}, {"v1.timestamp.u64max.001", ^uint64(0)}} {
		o := token()
		o.AuthorTime = v.n
		envelope(v.id, encode(o), &o, false, object.Supported, "bytes", "20.1", "44")
	}
	t = token()
	t.Issuer = strings.Repeat("é", 128)
	envelope("v1.encoding.utf8-boundary.001", encode(t), &t, false, object.Supported, "bytes", "44–46")
	for _, v := range []struct {
		id          string
		o           object.Object
		n, res, fan int
		fits        bool
	}{{"v1.size.token-max-4.001", token(), 4, 999, 4, true}, {"v1.size.token-max-5-fold.001", token(), 5, 1035, 4, false}, {"v1.size.device-max-14.001", device(), 14, 973, 14, true}, {"v1.size.device-max-15-fold.001", device(), 15, 1009, 14, false}, {"v1.size.short-der-no-extra-parent.001", device(), 15, 1009, 14, false}} {
		o := v.o
		o.Parents = parents(v.n)
		if o.Type == object.Token {
			o.Issuer = strings.Repeat("I", 256)
			o.Account = strings.Repeat("A", 256)
			o.Secret = bytes.Repeat([]byte{7}, 128)
		} else {
			o.DisplayName = strings.Repeat("D", 256)
		}
		if v.id == "v1.size.short-der-no-extra-parent.001" {
			// A valid fixed signature can physically fit, while reserved planning
			// still rejects the same shape. Low-S normalization after one
			// fixture signing bounds DER to <=71 bytes; there is no retry.
			o.DisplayName = strings.Repeat("D", 254)
			data, err := os.ReadFile(filepath.Join(rootDir, "vectors", casePath(v.id)))
			if err == nil {
				var previous vectors.Case
				must(vectors.Decode(data, &previous))
				if previous.Input.DisplayName == o.DisplayName && len(previous.Input.Signature) > 0 {
					o.Signature = previous.Input.Signature
				}
			} else if !os.IsNotExist(err) {
				must(err)
			}
			if len(o.Signature) == 0 {
				in, err := keys.SignatureInput(o)
				must(err)
				digest := sha256.Sum256(in)
				r, s, err := ecdsa.Sign(rand.Reader, private, digest[:])
				must(err)
				half := new(big.Int).Rsh(new(big.Int).Set(private.Params().N), 1)
				if s.Cmp(half) > 0 {
					s.Sub(private.Params().N, s)
				}
				o.Signature, err = asn1.Marshal(struct{ R, S *big.Int }{r, s})
				must(err)
			}
			if keys.Provenance(o, public) != "VERIFIED" {
				panic("capacity signature rejected")
			}
			v.res = 1007
		}
		expected := "FOLD"
		if v.fits {
			expected = "FITS"
		}
		add(vectors.Case{ID: v.id, Operation: "size", Expected: expected, Input: &o, Root: hx(root), PublicKey: hx(public), Size: &vectors.Size{Reserved: v.res, FanIn: v.fan, Fits: v.fits}}, "semantic", "45")
	}
}
func fullCrypto() {
	var rootID []byte
	for _, v := range []struct {
		id string
		o  object.Object
	}{{"v1.crypto.token-root.001", token()}, {"v1.crypto.token-child.001", token()}, {"v1.crypto.device-root.001", device()}} {
		o := v.o
		if strings.Contains(v.id, "child") {
			o.Parents = [][]byte{rootID}
			o.Account = "updated@example.test"
		}
		o = signed(v.id, o)
		p := encode(o)
		envelope(v.id, p, &o, true, object.Supported, "bytes", "11–16", "19–21")
		if strings.Contains(v.id, "token-root") {
			rootID = cryptov1.MAC(keys.ID, p)
		}
	}
	o := signed("v1.crypto.token-root.001", token())
	for _, v := range []struct {
		id, expected string
		key          []byte
		reject       bool
	}{{"v1.provenance.token-verified.001", "VERIFIED", public, false}, {"v1.provenance.token-rejected.001", "REJECTED", public, true}, {"v1.provenance.token-unresolved.001", "UNRESOLVED", nil, false}} {
		x := o
		if v.reject {
			x.Signature = []byte{0x30, 0}
		}
		add(vectors.Case{ID: v.id, Operation: "provenance", Expected: v.expected, Input: &x, Root: hx(root), PublicKey: hx(v.key)}, "semantic", "16", "19", "21.2")
	}
	// Fixed valid signatures exercising DER length. This offline fixture-only
	// selection is unrelated to capacity planning; writers never retry to fit.
	for _, v := range []struct {
		id     string
		length int
	}{{"v1.provenance.der-short-valid.001", 70}, {"v1.provenance.der-max-valid.001", 72}} {
		path := filepath.Join(rootDir, "vectors", casePath(v.id))
		b, e := os.ReadFile(path)
		x := token()
		if e == nil {
			var c vectors.Case
			must(vectors.Decode(b, &c))
			x = *c.Input
		} else {
			if !os.IsNotExist(e) {
				must(e)
			}
			for i := 0; i < 512; i++ {
				x, e = keys.Sign(x, private)
				must(e)
				if len(x.Signature) == v.length {
					break
				}
			}
		}
		if len(x.Signature) != v.length || keys.Provenance(x, public) != "VERIFIED" {
			panic("DER length fixture invalid")
		}
		add(vectors.Case{ID: v.id, Operation: "provenance", Expected: "VERIFIED", Input: &x, Root: hx(root), PublicKey: hx(public)}, "bytes", "16", "19", "45.1")
	}
	for _, v := range []struct {
		id       string
		password []byte
	}{{"v1.bootstrap.ascii.001", []byte("fixture password")}, {"v1.bootstrap.empty.001", nil}, {"v1.bootstrap.unicode.001", []byte("é\u0301 密碼")}} {
		salt := bytes.Repeat([]byte{0x33}, 16)
		nonce := bytes.Repeat([]byte{0x44}, 12)
		record, e := cryptov1.Wrap(v.password, root, salt, nonce)
		must(e)
		key, e := cryptov1.WrapKey(v.password, salt)
		must(e)
		binding, e := cryptov1.Binding(root)
		must(e)
		add(vectors.Case{ID: v.id, Operation: "bootstrap", Expected: "VALID", Root: hx(root), Bootstrap: &vectors.Bootstrap{Password: hx(v.password), Salt: hx(salt), Nonce: hx(nonce), WrapKey: hx(key), Header: hx(record[:39]), Record: hx(record), Binding: hx(binding)}}, "bytes", "5–11")
	}
}
func profile() {
	put("vectors/manifest.json", manifest)
	hashFile := func(path string) string {
		b, e := os.ReadFile(filepath.Join(rootDir, path))
		must(e)
		return vectors.Hash(b)
	}
	p := vectors.Profile{Format: "totipo-requirements-v1", Status: "moving-pre-rc", Protocol: "totipo-v1", Revision: "r12", ManifestSHA256: hashFile("vectors/manifest.json"), SpecSHA256: hashFile("spec/totipo-vault-format-v1.md"), SchemaSHA256: hashFile("vectors/manifest.schema.json"), CaseSchemaSHA256: hashFile("vectors/case.schema.json")}
	for _, e := range manifest.Cases {
		p.Required = append(p.Required, vectors.Pin{ID: e.ID, SHA256: e.SHA256})
	}
	put("requirements/v1-pre-rc.json", p)
}
func main() {
	flag.StringVar(&rootDir, "root", ".", "repository root")
	flag.Parse()
	initial()
	fullCrypto()
	graphCases()
	totpCases()
	storageCases()
	hardeningCases()
	profile()
	fmt.Printf("wrote %d cases and moving pre-RC profile\n", len(manifest.Cases))
}
