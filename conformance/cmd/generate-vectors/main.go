// Command generate-vectors explicitly regenerates r18 fixtures and moving pins.
// It is never run by tests. Deterministic crypto fixtures use public test keys.
package main

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"totipo/conformance/internal/cryptov1"
	"totipo/conformance/internal/graph"
	"totipo/conformance/internal/object"
	"totipo/conformance/internal/tlv"
	"totipo/conformance/internal/vectors"
)

var root string
var cases []vectors.Case
var rootKey = bytes.Repeat([]byte{0x11}, 32)

func must(e error) {
	if e != nil {
		panic(e)
	}
}
func hx(b []byte) string                { return hex.EncodeToString(b) }
func raw(s string) []byte               { b, e := hex.DecodeString(s); must(e); return b }
func id(n byte) []byte                  { return bytes.Repeat([]byte{n}, 32) }
func field(tag uint16, b []byte) []byte { p, e := tlv.Encode(tag, b); must(e); return p }
func write(path string, v any) {
	b, e := json.MarshalIndent(v, "", "  ")
	must(e)
	must(os.MkdirAll(filepath.Dir(path), 0755))
	must(os.WriteFile(path, append(b, '\n'), 0644))
}
func base(name, op string) vectors.Case {
	return vectors.Case{Format: "totipo-case-v1", ID: "v1." + name + ".001", Operation: op, Expected: "PASS"}
}
func token() object.Object {
	return object.Object{Identity: id(0x22), Parents: [][]byte{}, Status: 1, Issuer: "Example", Account: "alice@example.test", Algorithm: 1, Digits: 6, Period: 30, Secret: []byte("12345678901234567890")}
}
func encoded(o object.Object) []byte { p, e := o.Encode(); must(e); return p }
func envelope(name string, p []byte, input *object.Object, valid bool) vectors.Case {
	c := base(name, "crypto")
	c.Root = hx(rootKey)
	c.Semantic = hx(p)
	if input != nil {
		copyInput := *input
		c.Input = &copyInput
	}
	c.Expected = object.Supported
	if !valid {
		c.Operation = "dispatch"
		c.Expected = object.Invalid
	}
	k, e := cryptov1.Derive(rootKey)
	must(e)
	nameID, b, e := k.Seal(p)
	must(e)
	rid := raw(nameID)
	padded, e := cryptov1.Padded(p)
	must(e)
	c.Crypto = &vectors.Crypto{ObjectID: nameID, IDKey: hx(k.ID), ObjectRootKey: hx(k.ObjectRoot), ObjectKey: hx(k.ObjectKey(rid)), Nonce: hx(rid[:12]), AAD: hx(cryptov1.AAD(rid)), SemanticLength: len(p), Padded: hx(padded), Ciphertext: hx(b[:1008]), Tag: hx(b[1008:]), Object: hx(b)}
	cases = append(cases, c)
	return c
}

// Malformed fixtures deliberately bypass canonical Seal in this generator only.
// The production parser/writer APIs stay strict. Encrypt under the filename ID's
// context so AEAD succeeds even in the keyed-ID-mismatch case.
func postAEAD(defect string) {
	p := encoded(token())
	k, e := cryptov1.Derive(rootKey)
	must(e)
	rid := cryptov1.MAC(k.ID, p)
	plain, e := cryptov1.Padded(p)
	must(e)
	switch defect {
	case "nonzero-padding":
		plain[len(plain)-1] = 1
	case "object-id-mismatch":
		rid[0] ^= 1
	case "semantic-length-invalid":
		binary.BigEndian.PutUint16(plain, 1007)
	default:
		panic("unknown defect")
	}
	block, e := aes.NewCipher(k.ObjectKey(rid))
	must(e)
	a, e := cipher.NewGCM(block)
	must(e)
	b := a.Seal(nil, rid[:12], plain, cryptov1.AAD(rid))
	c := base("crypto."+defect, "post-aead")
	c.Expected = "INVALID_STORAGE"
	c.Root = hx(rootKey)
	c.Semantic = hx(p)
	c.Notes = "GCM authentication succeeds under the exact filename-derived context. The named post-AEAD boundary rejects; no TOKEN state is accepted. No HMAC collision is claimed."
	c.PostAEAD = &vectors.PostAEAD{ObjectID: hx(rid), Plaintext: hx(plain), Object: hx(b), Defect: defect}
	cases = append(cases, c)
}
func node(n, identity, value string, parents ...string) *graph.Node {
	return &graph.Node{ID: n, Identity: identity, Parents: parents, Canonical: n + "/" + identity + "/" + value + "/" + strings.Join(parents, ","), Value: value}
}
func metadata(n *graph.Node, name string, time uint64) *graph.Node {
	n.ClientName = &name
	n.ClientTime = &time
	n.Canonical += fmt.Sprintf("/metadata/%q/%d", name, time)
	return n
}
func add(n *graph.Node) vectors.Step { return vectors.Step{Action: "add", Node: n} }
func result(identity string, heads, values, missing []string) vectors.Step {
	if heads == nil {
		heads = []string{}
	}
	if values == nil {
		values = []string{}
	}
	if missing == nil {
		missing = []string{}
	}
	return vectors.Step{Action: "evaluate", Identity: identity, Want: &graph.Result{Heads: heads, Values: values, Unresolved: missing, Conflicting: len(values) > 1}}
}
func graphCase(name string, steps ...vectors.Step) {
	c := base("graph."+name, "graph")
	c.Notes = "Abstract validated TOKEN facts; symbolic IDs and cycles are not encrypted collision fixtures."
	// Expected head IDs/values are authored above, not queried from an evaluator.
	// Attach each declared head's exact fixture record to that explicit expectation.
	records := map[string]graph.Node{}
	for i := range steps {
		step := &steps[i]
		if step.Action == "add" {
			if step.IntegrityError {
				delete(records, step.Node.ID)
			} else {
				records[step.Node.ID] = *step.Node
			}
		}
		if step.Action == "remove" {
			delete(records, step.ID)
		}
		if step.Want != nil {
			step.Want.HeadObjects = []graph.Node{}
			for _, id := range step.Want.Heads {
				n, ok := records[id]
				if !ok {
					panic("unknown expected head")
				}
				step.Want.HeadObjects = append(step.Want.HeadObjects, n)
			}
		}
	}
	c.Graph = &vectors.GraphCase{Steps: steps}
	cases = append(cases, c)
}
func workflow(name string, w vectors.Workflow) {
	c := base(name, "workflow")
	c.Workflow = &w
	cases = append(cases, c)
}
func storageCase(name string, entries []vectors.StorageEntry, classes []string, diagnostics bool) {
	c := base("storage."+name, "storage")
	c.Root = hx(rootKey)
	c.Storage = &vectors.StorageCase{NamespaceKind: "directory", Entries: entries, Classes: classes, Diagnostics: diagnostics}
	cases = append(cases, c)
}
func main() {
	flag.StringVar(&root, "root", ".", "repository root")
	flag.Parse()
	// RFC known answers remain independent, byte-identical input fixtures.
	for _, alg := range []string{"sha1", "sha256", "sha512"} {
		b, e := os.ReadFile(filepath.Join(root, "vectors/cases/totp/v1.totp.rfc6238-"+alg+".001.json"))
		must(e)
		var c vectors.Case
		must(vectors.Decode(b, &c))
		cases = append(cases, c)
	}
	for _, defect := range []string{"nonzero-padding", "object-id-mismatch", "semantic-length-invalid"} {
		postAEAD(defect)
	}
	o := token()
	r := envelope("crypto.token-root", encoded(o), &o, true)
	o = token()
	o.Parents = [][]byte{raw(r.Crypto.ObjectID)}
	child := envelope("crypto.token-child", encoded(o), &o, true)
	o = token()
	envelope("encoding.token-root", encoded(o), &o, true)
	for _, n := range []int{1, 2, 3, 4} {
		o = token()
		for i := 0; i < n; i++ {
			o.Parents = append(o.Parents, id(byte(i+1)))
		}
		envelope(fmt.Sprintf("encoding.parents-%d", n), encoded(o), &o, true)
	}
	for _, name := range []string{"absent", "empty", "max"} {
		o = token()
		if name != "absent" {
			s := ""
			if name == "max" {
				s = strings.Repeat("é", 64)
			}
			o.ClientName = &s
		}
		envelope("metadata.client-name-"+name, encoded(o), &o, true)
	}
	for _, name := range []string{"absent", "zero", "normal", "u64max"} {
		o = token()
		if name != "absent" {
			var t uint64
			if name == "normal" {
				t = 1700000000
			}
			if name == "u64max" {
				t = ^uint64(0)
			}
			o.ClientTime = &t
		}
		envelope("metadata.client-time-"+name, encoded(o), &o, true)
	}
	o = token()
	o.Status = 2
	envelope("encoding.complete-tombstone", encoded(o), &o, true)
	o = token()
	o.Issuer = strings.Repeat("é", 128)
	o.Account = "\x00\n"
	envelope("encoding.utf8-boundary", encoded(o), &o, true)
	o = token()
	o.Issuer = strings.Repeat("i", 256)
	o.Account = strings.Repeat("a", 256)
	o.Secret = bytes.Repeat([]byte{0xab}, 128)
	s := strings.Repeat("c", 128)
	t := ^uint64(0)
	o.ClientName = &s
	o.ClientTime = &t
	o.Period = ^uint32(0)
	o.Algorithm = 3
	o.Digits = 8
	for i := 0; i < 4; i++ {
		o.Parents = append(o.Parents, id(byte(i+1)))
	}
	max := encoded(o)
	if len(max) != 1005 {
		panic("maximum arithmetic")
	}
	envelope("size.token-max-4", max, &o, true)
	// Malformed plaintexts are still genuinely authenticated and keyed, isolating
	// semantic rejection from the outer envelope validation boundary.
	p := encoded(token())
	envelope("encoding.parent-count-mismatch", append(append(bytes.Clone(p[:36]), field(2, tlv.U16(1))...), p[42:]...), nil, false)
	for _, x := range []struct {
		name    string
		parents [][]byte
	}{{"parent-order", [][]byte{id(2), id(1)}}, {"parent-duplicate", [][]byte{id(1), id(1)}}, {"five-parents", [][]byte{id(1), id(2), id(3), id(4), id(5)}}} {
		q := append(bytes.Clone(p[:36]), field(2, tlv.U16(uint16(len(x.parents))))...)
		for _, parent := range x.parents {
			q = append(q, field(3, parent)...)
		}
		q = append(q, p[42:]...)
		envelope("encoding."+x.name, q, nil, false)
	}
	envelope("encoding.duplicate-nonrepeatable", append(bytes.Clone(p), field(10, []byte{1})...), nil, false)
	envelope("encoding.unknown-field", append(bytes.Clone(p), field(13, nil)...), nil, false)
	envelope("encoding.missing-secret", p[:len(p)-24], nil, false)
	envelope("encoding.trailing-byte", append(bytes.Clone(p), 0), nil, false)
	envelope("encoding.client-name-too-long", append(bytes.Clone(p), field(11, bytes.Repeat([]byte{'n'}, 129))...), nil, false)
	envelope("encoding.client-name-invalid-utf8", append(bytes.Clone(p), field(11, []byte{0xff})...), nil, false)
	envelope("encoding.client-time-width", append(bytes.Clone(p), field(12, make([]byte, 7))...), nil, false)
	q := append(bytes.Clone(p), field(12, tlv.U64(0))...)
	q = append(q, field(11, nil)...)
	envelope("encoding.metadata-order", q, nil, false)
	envelope("encoding.non-token-plaintext", []byte("authenticated but not a TOKEN"), nil, false)
	for _, x := range []struct {
		name  string
		tag   uint16
		value []byte
	}{{"status-range", 4, []byte{3}}, {"algorithm-range", 7, []byte{4}}, {"digits-range", 8, []byte{5}}, {"period-zero", 9, tlv.U32(0)}, {"secret-empty", 10, nil}, {"secret-too-long", 10, make([]byte, 129)}, {"issuer-too-long", 5, make([]byte, 257)}, {"issuer-invalid-utf8", 5, []byte{0xff}}, {"account-too-long", 6, make([]byte, 257)}, {"identity-width", 1, make([]byte, 31)}} {
		rest := p
		q := []byte{}
		for len(rest) > 0 {
			f, r, e := tlv.Take(rest)
			must(e)
			rest = r
			if f.Tag == x.tag {
				f.Value = x.value
			}
			q = append(q, field(f.Tag, f.Value)...)
		}
		envelope("encoding."+x.name, q, nil, false)
	}
	for _, x := range []struct {
		name, password string
		salt, nonce    byte
	}{{"ascii", "fixture password", 0x33, 0x44}, {"empty", "", 0x33, 0x44}, {"unicode", "päss🔑", 0x33, 0x44}, {"rewrap", "replacement password", 0x55, 0x66}} {
		c := base("bootstrap."+x.name, "bootstrap")
		c.Expected = "VALID"
		c.Root = hx(rootKey)
		password := []byte(x.password)
		salt := bytes.Repeat([]byte{x.salt}, 16)
		nonce := bytes.Repeat([]byte{x.nonce}, 12)
		// Preserve the original exact Unicode password and wrapping test inputs.
		if x.name == "unicode" {
			b, e := os.ReadFile(filepath.Join(root, "vectors/cases/bootstrap/v1.bootstrap.unicode.001.json"))
			must(e)
			var old struct {
				Bootstrap struct {
					Password string `json:"password_hex"`
				}
			}
			must(json.Unmarshal(b, &old))
			password = raw(old.Bootstrap.Password)
		}
		rec, e := cryptov1.Wrap(password, rootKey, salt, nonce)
		must(e)
		key, e := cryptov1.WrapKey(password, salt)
		must(e)
		fp, e := cryptov1.Fingerprint(rootKey)
		must(e)
		c.Bootstrap = &vectors.Bootstrap{Password: hx(password), Salt: hx(salt), Nonce: hx(nonce), WrapKey: hx(key), Header: hx(rec[:39]), Record: hx(rec), Fingerprint: hx(fp)}
		cases = append(cases, c)
	}
	graphCase("sequential", add(node("A", "T", "X")), add(node("B", "T", "Y", "A")), result("T", []string{"B"}, []string{"Y"}, nil))
	graphCase("equal-concurrent", add(metadata(node("A", "T", "X"), "Laptop", 0)), add(metadata(node("B", "T", "X"), "Phone", ^uint64(0))), result("T", []string{"A", "B"}, []string{"X"}, nil))
	graphCase("conflicting-concurrent", add(node("A", "T", "X")), add(node("B", "T", "Y")), result("T", []string{"A", "B"}, []string{"X", "Y"}, nil))
	graphCase("late-parent", add(node("B", "T", "Y", "A")), result("T", []string{"B"}, []string{"Y"}, []string{"A"}), add(node("A", "T", "X")), result("T", []string{"B"}, []string{"Y"}, nil))
	graphCase("wrong-identity-parent", add(node("A", "U", "X")), add(node("B", "T", "Y", "A")), result("T", []string{"B"}, []string{"Y"}, []string{"A"}))
	graphCase("intermediate-disappears", add(node("A", "T", "X")), add(node("B", "T", "Y", "A")), add(node("C", "T", "Z", "B")), vectors.Step{Action: "remove", ID: "B"}, result("T", []string{"A", "C"}, []string{"X", "Z"}, []string{"B"}))
	graphCase("reappearance", add(node("A", "T", "X")), add(node("B", "T", "Y", "A")), vectors.Step{Action: "remove", ID: "A"}, result("T", []string{"B"}, []string{"Y"}, []string{"A"}), add(node("A", "T", "X")), result("T", []string{"B"}, []string{"Y"}, nil))
	for _, equal := range []bool{true, false} {
		name, value := "cycle-equal", "X"
		values := []string{"X"}
		if !equal {
			name, value = "cycle-conflicting", "Y"
			values = []string{"X", "Y"}
		}
		graphCase(name, add(metadata(node("A", "T", "X", "B"), "Desktop", 0)), add(metadata(node("B", "T", value, "A"), "Mobile", ^uint64(0))), result("T", []string{"A", "B"}, values, nil), add(node("D", "T", "Z", "A")), result("T", []string{"D"}, []string{"Z"}, nil))
	}
	graphCase("late-cycle", add(node("A", "T", "X", "B")), add(node("B", "T", "X", "C")), result("T", []string{"A"}, []string{"X"}, []string{"C"}), add(node("C", "T", "X", "A")), result("T", []string{"A", "B", "C"}, []string{"X"}, nil))
	graphCase("self-cycle", add(node("A", "T", "X", "A")), result("T", []string{"A"}, []string{"X"}, nil))
	graphCase("same-id-defensive", add(node("A", "T", "X")), vectors.Step{Action: "add", Node: node("A", "T", "Y"), IntegrityError: true}, add(node("B", "T", "Z", "A")), add(node("U", "U", "W")), result("T", []string{"B"}, []string{"Z"}, []string{"A"}), result("U", []string{"U"}, []string{"W"}, nil))
	graphCase("identical-collapse", add(node("A", "T", "X")), add(node("A", "T", "X")), result("T", []string{"A"}, []string{"X"}, nil))
	for _, n := range []int{0, 4, 5, 7, 10, 11} {
		c := base(fmt.Sprintf("fold.width-%d", n), "fold")
		f := &vectors.FoldCase{Token: token(), Frontier: []string{}, StageIDs: []string{}, Parents: [][]string{}}
		switch n {
		case 5:
			name := ""
			time := uint64(0)
			f.Token.ClientName = &name
			f.Token.ClientTime = &time
		case 7:
			time := ^uint64(0)
			f.Token.ClientTime = &time
		case 10:
			name := "Alice's laptop"
			time := uint64(1770000000)
			f.Token.ClientName = &name
			f.Token.ClientTime = &time
		case 11:
			name := "e\u0301 laptop"
			f.Token.ClientName = &name
		}
		for i := 1; i <= n; i++ {
			f.Frontier = append(f.Frontier, hx(id(byte(i))))
		}
		remaining := append([]string{}, f.Frontier...)
		carry := ""
		stage := 0
		for len(remaining) > 0 || stage == 0 {
			capacity := 4
			if carry != "" {
				capacity = 3
			}
			take := min(capacity, len(remaining))
			parents := append([]string{}, remaining[:take]...)
			remaining = remaining[take:]
			if carry != "" {
				parents = append(parents, carry)
			}
			sort.Strings(parents)
			carry = hx(id(byte(100 + stage)))
			f.Parents = append(f.Parents, parents)
			f.StageIDs = append(f.StageIDs, carry)
			stage++
			if len(remaining) == 0 {
				break
			}
		}
		c.Fold = f
		c.Notes = "Symbolic stage IDs test fixed fold planning with one complete operation. Every stage must preserve the token value and exact client metadata. Integration tests also construct and open encrypted stages."
		cases = append(cases, c)
	}
	valid := vectors.StorageEntry{Path: "objects-v1/" + r.Crypto.ObjectID, Kind: "regular", Object: r.Crypto.Object}
	storageCase("objects-v1", []vectors.StorageEntry{valid}, []string{object.Supported}, false)
	ignored := valid
	ignored.Path += ".tmp"
	storageCase("nonobject-name-ignored", []vectors.StorageEntry{ignored}, []string{}, false)
	ignored = valid
	ignored.Path = "objects-v2/" + r.Crypto.ObjectID
	storageCase("unknown-sibling-ignored", []vectors.StorageEntry{ignored}, []string{}, false)
	wrong := valid
	wrong.Object = "00"
	storageCase("wrong-size", []vectors.StorageEntry{wrong}, []string{"INVALID_STORAGE"}, false)
	unreadable := valid
	unreadable.Path = "objects-v1/" + hx(id(0xfa))
	unreadable.Unreadable = true
	storageCase("incomplete-diagnostics", []vectors.StorageEntry{unreadable, valid}, []string{object.Supported}, true)
	wrong = valid
	b := raw(wrong.Object)
	b[0] ^= 1
	wrong.Object = hx(b)
	storageCase("failed-aead", []vectors.StorageEntry{wrong, valid}, []string{"INVALID_STORAGE", object.Supported}, false)
	for _, c := range cases {
		if c.ID == "v1.encoding.non-token-plaintext.001" {
			storageCase("invalid-grammar", []vectors.StorageEntry{{Path: "objects-v1/" + c.Crypto.ObjectID, Kind: "regular", Object: c.Crypto.Object}}, []string{object.Invalid}, false)
			break
		}
	}
	for _, x := range []struct {
		name, kind, existing, result string
		durable, parents             bool
	}{{"exact-existing-publication", "regular", r.Crypto.Object, "ALREADY_PRESENT_EXACT", false, false}, {"non-exact-publication", "regular", "00", "FAILED", true, true}, {"unreadable-target", "unreadable", "", "FAILED", true, true}, {"new-publication", "absent", "", "PUBLISHED_NEW", true, true}, {"missing-parent-publication", "absent", "", "PUBLISHED_NEW", true, false}, {"ambiguous-publication", "absent", "", "FAILED", false, false}} {
		intended := r.Crypto.Object
		if x.name == "missing-parent-publication" {
			intended = child.Crypto.Object
		}
		workflow("storage."+x.name, vectors.Workflow{Action: "publish", Kind: x.kind, Existing: x.existing, Intended: intended, Complete: true, Durable: x.durable, ParentsAvailable: x.parents, Result: x.result})
	}
	for _, x := range []struct {
		name, kind, result string
		orphans, durable   bool
	}{{"create-orphans", "absent", "CREATED", true, true}, {"create-empty", "absent", "CREATED", false, true}, {"create-existing", "regular", "FAILED", false, true}, {"create-ambiguous", "absent", "FAILED", true, false}} {
		workflow("vault."+x.name, vectors.Workflow{Action: "create", Kind: x.kind, Complete: true, Durable: x.durable, OrphanObjects: x.orphans, Result: x.result})
	}
	for _, x := range []struct {
		name, current, result string
		readable, durable     bool
	}{{"replace-equal", "01", "REPLACED", true, true}, {"replace-stale", "02", "STALE", true, true}, {"replace-unreadable", "", "FAILED", false, true}, {"replace-ambiguous", "01", "FAILED", true, false}} {
		workflow("vault."+x.name, vectors.Workflow{Action: "replace", Kind: "regular", Base: "01", Existing: x.current, Complete: true, Readable: x.readable, Durable: x.durable, Result: x.result})
	}
	sort.Slice(cases, func(i, j int) bool { return cases[i].ID < cases[j].ID })
	manifest := vectors.Manifest{Format: "totipo-vector-manifest-v1", Protocol: "totipo-v1", Revision: "r18", Cases: []vectors.Entry{}}
	keep := map[string]bool{}
	for _, c := range cases {
		category := strings.Split(c.ID, ".")[1]
		path := "cases/" + category + "/" + c.ID + ".json"
		full := filepath.Join(root, "vectors", path)
		if c.Operation != "totp" {
			write(full, c)
		}
		b, e := os.ReadFile(full)
		must(e)
		kind := "semantic"
		sections := []string{"15", "16"}
		switch c.Operation {
		case "crypto", "dispatch":
			kind = "bytes"
			sections = []string{"10", "11", "12", "13"}
		case "post-aead":
			kind = "negative"
			sections = []string{"11", "13"}
		case "bootstrap":
			kind = "bytes"
			sections = []string{"4", "5", "6", "9"}
		case "totp":
			kind = "bytes"
			sections = []string{"19"}
		case "fold":
			sections = []string{"17"}
		case "storage":
			sections = []string{"3", "13", "14"}
		case "workflow":
			sections = []string{"7", "8", "18"}
		}
		if c.Expected == object.Invalid {
			kind = "negative"
		}
		manifest.Cases = append(manifest.Cases, vectors.Entry{ID: c.ID, Category: category, Kind: kind, Normative: true, Path: path, Expected: c.Expected, Sections: sections, SHA256: vectors.Hash(b)})
		keep[filepath.Clean(full)] = true
	}
	must(filepath.WalkDir(filepath.Join(root, "vectors/cases"), func(path string, d os.DirEntry, e error) error {
		if e != nil {
			return e
		}
		if !d.IsDir() && !keep[filepath.Clean(path)] {
			return os.Remove(path)
		}
		return nil
	}))
	write(filepath.Join(root, "vectors/manifest.json"), manifest)
	hashFile := func(path string) string {
		b, e := os.ReadFile(filepath.Join(root, path))
		must(e)
		return vectors.Hash(b)
	}
	profile := vectors.Profile{Format: "totipo-requirements-v1", Status: "moving-pre-rc", Protocol: "totipo-v1", Revision: "r18", SpecSHA256: hashFile("spec/totipo-vault-format-v1.md"), ManifestSHA256: hashFile("vectors/manifest.json"), SchemaSHA256: hashFile("vectors/manifest.schema.json"), CaseSchemaSHA256: hashFile("vectors/case.schema.json"), Required: []vectors.Pin{}}
	for _, e := range manifest.Cases {
		profile.Required = append(profile.Required, vectors.Pin{ID: e.ID, SHA256: e.SHA256})
	}
	write(filepath.Join(root, "requirements/v1-pre-rc.json"), profile)
	fmt.Printf("Generated r18: %d cases\n", len(cases))
}
