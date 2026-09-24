// Isolated public-fixture arithmetic reviewer. Not constant-time; never use with secrets.
// No conformance imports, no crypto/ed25519, no third-party curve code.
package main

import (
	"bytes"
	"crypto/sha512"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"math/big"
	"os"
	"sort"
)

var p = new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), 255), big.NewInt(19))
var l = integer("7237005577332262213973186563042994240857116359379907606001950938285454250989")
var d = mod(new(big.Int).Mul(big.NewInt(-121665), new(big.Int).ModInverse(big.NewInt(121666), p)))

type point struct{ x, y *big.Int }

var identity = point{big.NewInt(0), big.NewInt(1)}

func integer(s string) *big.Int {
	v, ok := new(big.Int).SetString(s, 10)
	if !ok {
		panic(s)
	}
	return v
}
func mod(x *big.Int) *big.Int    { return new(big.Int).Mod(x, p) }
func add(a, b *big.Int) *big.Int { return mod(new(big.Int).Add(a, b)) }
func sub(a, b *big.Int) *big.Int { return mod(new(big.Int).Sub(a, b)) }
func mul(a, b *big.Int) *big.Int { return mod(new(big.Int).Mul(a, b)) }
func divide(a, b *big.Int) *big.Int {
	inv := new(big.Int).ModInverse(b, p)
	if inv == nil {
		panic("noninvertible denominator")
	}
	return mul(a, inv)
}
func plus(a, b point) point {
	xy := mul(d, mul(mul(a.x, b.x), mul(a.y, b.y)))
	return point{divide(add(mul(a.x, b.y), mul(a.y, b.x)), add(big.NewInt(1), xy)), divide(add(mul(a.y, b.y), mul(a.x, b.x)), sub(big.NewInt(1), xy))}
}
func times(a point, n *big.Int) point {
	q := identity
	for i := n.BitLen() - 1; i >= 0; i-- {
		q = plus(q, q)
		if n.Bit(i) != 0 {
			q = plus(q, a)
		}
	}
	return q
}
func equal(a, b point) bool { return a.x.Cmp(b.x) == 0 && a.y.Cmp(b.y) == 0 }
func little(b []byte) *big.Int {
	r := append([]byte{}, b...)
	for i, j := 0, len(r)-1; i < j; i, j = i+1, j-1 {
		r[i], r[j] = r[j], r[i]
	}
	return new(big.Int).SetBytes(r)
}
func encoded(n *big.Int, size int) []byte {
	b := n.Bytes()
	r := make([]byte, size)
	for i := range b {
		r[i] = b[len(b)-1-i]
	}
	return r
}
func encode(a point) []byte { r := encoded(a.y, 32); r[31] |= byte(a.x.Bit(0)) << 7; return r }
func decode(b []byte) (point, bool) {
	if len(b) != 32 {
		return point{}, false
	}
	v := append([]byte{}, b...)
	sign := uint(v[31] >> 7)
	v[31] &= 127
	y := mod(little(v))
	yy := mul(y, y)
	xx := divide(sub(yy, big.NewInt(1)), add(mul(d, yy), big.NewInt(1)))
	x := new(big.Int).ModSqrt(xx, p)
	if x == nil {
		return point{}, false
	}
	if x.Bit(0) != sign {
		x = mod(new(big.Int).Neg(x))
	}
	return point{x, y}, true
}
func hx(b []byte) string    { return hex.EncodeToString(b) }
func unhex(s string) []byte { b, e := hex.DecodeString(s); must(e); return b }
func must(e error) {
	if e != nil {
		panic(e)
	}
}

var base = func() point {
	a, ok := decode(unhex("5866666666666666666666666666666666666666666666666666666666666666"))
	if !ok {
		panic("base")
	}
	return a
}()

func challenge(r, a, m []byte) *big.Int {
	h := sha512.New()
	h.Write(r)
	h.Write(a)
	h.Write(m)
	return new(big.Int).Mod(little(h.Sum(nil)), l)
}
func equation(a point, pk, m, sig []byte) bool {
	r, ok := decode(sig[:32])
	return ok && equal(times(base, little(sig[32:])), plus(r, times(a, challenge(sig[:32], pk, m))))
}
func review(pk, m, sig []byte) string {
	if len(pk) != 32 {
		return "A_LENGTH"
	}
	if len(sig) != 64 {
		return "SIGNATURE_LENGTH"
	}
	a, ok := decode(pk)
	if !ok {
		return "A_DECODE"
	}
	r, ok := decode(sig[:32])
	if !ok {
		return "R_DECODE"
	}
	if !bytes.Equal(encode(a), pk) {
		return "A_CANONICAL"
	}
	if !bytes.Equal(encode(r), sig[:32]) {
		return "R_CANONICAL"
	}
	if equal(times(a, big.NewInt(8)), identity) {
		return "A_SMALL_ORDER"
	}
	if equal(times(r, big.NewInt(8)), identity) {
		return "R_SMALL_ORDER"
	}
	if !equal(times(a, l), identity) {
		return "A_SUBGROUP"
	}
	if !equal(times(r, l), identity) {
		return "R_SUBGROUP"
	}
	if little(sig[32:]).Cmp(l) >= 0 {
		return "S_RANGE"
	}
	if !equation(a, pk, m, sig) {
		return "EQUATION"
	}
	return "ACCEPT"
}
func sign(seed, m []byte) ([]byte, []byte) {
	h := sha512.Sum512(seed)
	aBytes := append([]byte{}, h[:32]...)
	aBytes[0] &= 248
	aBytes[31] &= 63
	aBytes[31] |= 64
	a := little(aBytes)
	pk := encode(times(base, a))
	nonceHash := sha512.Sum512(append(append([]byte{}, h[32:]...), m...))
	r := new(big.Int).Mod(little(nonceHash[:]), l)
	rBytes := encode(times(base, r))
	s := new(big.Int).Mod(new(big.Int).Add(r, new(big.Int).Mul(challenge(rBytes, pk, m), a)), l)
	return pk, append(rBytes, encoded(s, 32)...)
}

type provenance struct {
	Source  string `json:"source"`
	Status  string `json:"status"`
	Locator string `json:"locator"`
}
type vector struct {
	ID          string            `json:"id"`
	Kind        string            `json:"kind"`
	Description string            `json:"description"`
	Provenance  provenance        `json:"provenance"`
	Input       map[string]string `json:"input"`
	Expected    map[string]string `json:"expected"`
}
type bundle struct {
	Schema int      `json:"schema"`
	Cases  []vector `json:"cases"`
}

func main() {
	write := flag.Bool("write-vectors", false, "explicit maintenance only")
	flag.Parse()
	path := "vectors/v0/ed25519/phase2.json"
	if !*write {
		raw, e := os.ReadFile(path)
		must(e)
		var b bundle
		must(json.Unmarshal(raw, &b))
		for _, c := range b.Cases {
			got := review(unhex(c.Input["public_key_hex"]), unhex(c.Input["message_hex"]), unhex(c.Input["signature_hex"]))
			if got != c.Expected["reason"] {
				panic(c.ID + ": expected " + c.Expected["reason"] + ", got " + got)
			}
			fmt.Println("REVIEW PASS", c.ID, got)
		}
		return
	}
	var cases []vector
	put := func(id, description string, pk, m, sig []byte, reason string) {
		disposition := "REJECT"
		if reason == "ACCEPT" {
			disposition = "ACCEPT"
		}
		actual := review(pk, m, sig)
		if actual != reason {
			panic(id + ": manual reason " + reason + ", arithmetic " + actual)
		}
		cases = append(cases, vector{"v0/ed25519/" + id, "ed25519", description, provenance{"review/ed25519/REVIEW.md", "spec-derived-reviewed", "Section 16; " + id}, map[string]string{"public_key_hex": hx(pk), "message_hex": hx(m), "signature_hex": hx(sig)}, map[string]string{"disposition": disposition, "reason": reason}})
	}
	pk := unhex("d75a980182b10ab7d54bfed3c964073a0ee172f3daa62325af021a68f707511a")
	sig := unhex("e5564300c360ac729086e2cc806e828a84877f1eb8e5d974d873e065224901555fb8821590a33bacc61e39701cf9b46bd25bf5f0595bbe24655141438e7a100b")
	put("rfc8032-empty", "RFC 8032 section 7.1 TEST 1", pk, nil, sig, "ACCEPT")
	cases[len(cases)-1].Provenance = provenance{"https://www.rfc-editor.org/rfc/rfc8032.html#section-7.1", "published-standard", "TEST 1"}
	// Signer arithmetic is independently anchored by reproducing the RFC signature.
	rk, rs := sign(unhex("9d61b19deffd5a60ba844af492ec2cc44449c5697b326919703bac031cae7f60"), nil)
	if !bytes.Equal(rk, pk) || !bytes.Equal(rs, sig) {
		panic("RFC signing anchor")
	}
	ak, as := sign(bytes.Repeat([]byte{0x42}, 32), []byte("Totipo independent arithmetic signature"))
	put("independent-positive", "Signature produced by isolated affine arithmetic", ak, []byte("Totipo independent arithmetic signature"), as, "ACCEPT")
	raw, e := os.ReadFile("vectors/v0/object-crypto/cases.json")
	must(e)
	var objects bundle
	must(json.Unmarshal(raw, &objects))
	for _, c := range objects.Cases {
		id := "totipo-token"
		if c.Input["object_type"] == "DEVICE_UPDATE" {
			id = "totipo-device"
		}
		put(id, "External pinned Totipo signature", unhex(c.Expected["public_key_hex"]), unhex(c.Expected["signature_input_hex"]), unhex(c.Expected["signature_hex"]), "ACCEPT")
		cases[len(cases)-1].Provenance = provenance{"review/source-vectors/r31-object-vectors.txt", "reviewed-pinned", c.Input["object_type"]}
	}
	for _, n := range []int{0, 31, 33} {
		put(fmt.Sprintf("a-length-%d", n), "Public key length gate", make([]byte, n), nil, sig, "A_LENGTH")
	}
	for _, n := range []int{0, 63, 65} {
		put(fmt.Sprintf("sig-length-%d", n), "Signature length gate", pk, nil, make([]byte, n), "SIGNATURE_LENGTH")
	}
	withR := func(r []byte) []byte { return append(append([]byte{}, r...), sig[32:]...) }
	var bad []byte
	for y := int64(2); ; y++ {
		b := encoded(big.NewInt(y), 32)
		if _, ok := decode(b); !ok {
			bad = b
			break
		}
	}
	put("a-nondecode", "y gives a nonsquare x coordinate", bad, nil, sig, "A_DECODE")
	put("r-nondecode", "R has no Edwards point", pk, nil, withR(bad), "R_DECODE")
	noncanon := encoded(new(big.Int).Add(p, big.NewInt(1)), 32)
	put("a-noncanonical-y", "y=p+1 encodes the identity noncanonically", noncanon, nil, sig, "A_CANONICAL")
	put("r-noncanonical-y", "R uses y=p+1", pk, nil, withR(noncanon), "R_CANONICAL")
	negzero := encode(identity)
	negzero[31] |= 128
	put("a-negative-zero", "x=0 sign bit one fails re-encoding", negzero, nil, sig, "A_CANONICAL")
	put("r-negative-zero", "R x=0 sign bit one", pk, nil, withR(negzero), "R_CANONICAL")
	torsion := point{big.NewInt(0), new(big.Int).Sub(p, big.NewInt(1))}
	for name, q := range map[string]point{"identity": identity, "order-two": torsion} {
		put("a-"+name, "[8]A=identity", encode(q), nil, sig, "A_SMALL_ORDER")
		put("r-"+name, "[8]R=identity", pk, nil, withR(encode(q)), "R_SMALL_ORDER")
	}
	zeroSig := append(encode(identity), make([]byte, 32)...)
	put("identity-forgery", "Equation holds for identity A/R and S=0; small-order gate rejects", encode(identity), []byte("identity forgery"), zeroSig, "A_SMALL_ORDER")
	mixedA := plus(base, torsion)
	mixedR := plus(times(base, big.NewInt(2)), torsion)
	if equal(times(mixedA, big.NewInt(8)), identity) || equal(times(mixedA, l), identity) {
		panic("mixed-order construction")
	}
	put("a-mixed-order", "A=B+(0,-1); [L]A=(0,-1), not identity", encode(mixedA), nil, sig, "A_SUBGROUP")
	put("r-mixed-order", "R=2B+(0,-1); [L]R=(0,-1)", pk, nil, withR(encode(mixedR)), "R_SUBGROUP")
	// Build mathematically valid cofactorless signatures with mixed A, varying the
	// public message until the torsion equation u+k=0 mod 2 is satisfied.
	for u := int64(0); u < 2; u++ {
		r := times(base, big.NewInt(2))
		if u == 1 {
			r = plus(r, torsion)
		}
		a := encode(mixedA)
		rb := encode(r)
		for nonce := 0; ; nonce++ {
			m := []byte(fmt.Sprintf("mixed order %d/%d", u, nonce))
			k := challenge(rb, a, m)
			if int64(k.Bit(0)) != u {
				continue
			}
			s := new(big.Int).Mod(new(big.Int).Add(big.NewInt(2), k), l)
			signature := append(append([]byte{}, rb...), encoded(s, 32)...)
			if !equation(mixedA, a, m, signature) {
				panic("torsion equation")
			}
			put(fmt.Sprintf("mixed-equation-%d", u), "Cofactorless equation holds, but A is not in prime subgroup", a, m, signature, "A_SUBGROUP")
			break
		}
	}
	for name, s := range map[string]*big.Int{"equal-l": l, "greater-l": new(big.Int).Add(l, big.NewInt(1)), "maximum": new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), 256), big.NewInt(1))} {
		put("s-"+name, "S is not in [0,L)", pk, nil, append(append([]byte{}, sig[:32]...), encoded(s, 32)...), "S_RANGE")
	}
	// Negating a prime-subgroup point preserves canonical decoding and membership,
	// so these mutations isolate the final equation rather than an earlier gate.
	a, _ := decode(pk)
	a.x = mod(new(big.Int).Neg(a.x))
	put("modified-public-key", "Negated canonical prime-subgroup A fails equation", encode(a), nil, sig, "EQUATION")
	r, _ := decode(sig[:32])
	r.x = mod(new(big.Int).Neg(r.x))
	put("modified-r", "Negated canonical prime-subgroup R fails equation", pk, nil, withR(encode(r)), "EQUATION")
	ms := append([]byte{}, sig...)
	ms[32] ^= 1
	put("modified-s", "In-range scalar mutation fails equation", pk, nil, ms, "EQUATION")
	put("modified-message", "Message mutation fails equation", pk, []byte{1}, sig, "EQUATION")
	put("canonical-invalid-equation", "A=B R=B S=0 are individually canonical but not a signature", encode(base), []byte("invalid equation"), append(encode(base), make([]byte, 32)...), "EQUATION")
	sort.Slice(cases, func(i, j int) bool { return cases[i].ID < cases[j].ID })
	output, e := json.MarshalIndent(bundle{1, cases}, "", "  ")
	must(e)
	must(os.WriteFile(path, append(output, '\n'), 0644))
	fmt.Println("wrote", len(cases), "reviewed cases")
}
