package vectors

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"strings"
	"totipo/conformance/internal/cryptov1"
	"totipo/conformance/internal/graph"
	"totipo/conformance/internal/object"
)

func Decode(p []byte, v any) error {
	d := json.NewDecoder(bytes.NewReader(p))
	d.DisallowUnknownFields()
	if e := d.Decode(v); e != nil {
		return e
	}
	var extra any
	if e := d.Decode(&extra); e != io.EOF {
		return fmt.Errorf("trailing JSON")
	}
	return nil
}
func Hash(p []byte) string { h := sha256.Sum256(p); return hex.EncodeToString(h[:]) }
func unhex(s string) ([]byte, error) {
	p, e := hex.DecodeString(s)
	if e != nil || hex.EncodeToString(p) != s {
		return nil, fmt.Errorf("noncanonical hex")
	}
	return p, nil
}
func equalHex(label, s string, p []byte) error {
	b, e := unhex(s)
	if e != nil {
		return fmt.Errorf("%s: %w", label, e)
	}
	if !bytes.Equal(p, b) {
		return fmt.Errorf("%s mismatch", label)
	}
	return nil
}

var idPattern = regexp.MustCompile(`^v1\.[a-z0-9-]+(?:\.[a-z0-9-]+)+$`)
var hashPattern = regexp.MustCompile(`^[0-9a-f]{64}$`)

func Read(root string) (Manifest, []Case, error) {
	var m Manifest
	b, e := os.ReadFile(filepath.Join(root, "vectors/manifest.json"))
	if e != nil {
		return m, nil, e
	}
	if e = Decode(b, &m); e != nil {
		return m, nil, e
	}
	if m.Format != "totipo-vector-manifest-v1" || m.Protocol != "totipo-v1" || m.Revision != "r9" || len(m.Cases) == 0 {
		return m, nil, fmt.Errorf("invalid manifest header or empty corpus")
	}
	seen, paths := map[string]bool{}, map[string]bool{}
	cases := []Case{}
	for _, entry := range m.Cases {
		if !idPattern.MatchString(entry.ID) || seen[entry.ID] || paths[entry.Path] || entry.Category == "" || !entry.Normative || len(entry.Sections) == 0 || entry.Expected == "" || !hashPattern.MatchString(entry.SHA256) {
			return m, nil, fmt.Errorf("invalid/duplicate manifest entry %q", entry.ID)
		}
		if !strings.HasPrefix(entry.ID, "v1."+entry.Category+".") {
			return m, nil, fmt.Errorf("category mismatch: %s", entry.ID)
		}
		for _, section := range entry.Sections {
			if section == "" {
				return m, nil, fmt.Errorf("empty spec section")
			}
		}
		if entry.Kind != "bytes" && entry.Kind != "negative" && entry.Kind != "semantic" {
			return m, nil, fmt.Errorf("invalid kind")
		}
		if !filepath.IsLocal(entry.Path) || strings.Contains(entry.Path, "\\") || filepath.ToSlash(filepath.Clean(entry.Path)) != entry.Path || !strings.HasPrefix(entry.Path, "cases/") {
			return m, nil, fmt.Errorf("unsafe case path")
		}
		full := filepath.Join(root, "vectors", entry.Path)
		resolved, e := filepath.EvalSymlinks(full)
		if e != nil {
			return m, nil, e
		}
		base, _ := filepath.Abs(filepath.Join(root, "vectors"))
		resolved, _ = filepath.Abs(resolved)
		rel, e := filepath.Rel(base, resolved)
		if e != nil || !filepath.IsLocal(rel) {
			return m, nil, fmt.Errorf("case symlink escapes vectors")
		}
		b, e := os.ReadFile(full)
		if e != nil {
			return m, nil, e
		}
		if Hash(b) != entry.SHA256 {
			return m, nil, fmt.Errorf("%s checksum mismatch", entry.ID)
		}
		var c Case
		if e = Decode(b, &c); e != nil {
			return m, nil, fmt.Errorf("%s: %w", entry.ID, e)
		}
		if c.Format != "totipo-case-v1" || c.ID != entry.ID || c.Expected != entry.Expected {
			return m, nil, fmt.Errorf("case identity/expectation mismatch")
		}
		if e = ValidateShape(c, entry.Kind); e != nil {
			return m, nil, fmt.Errorf("%s: %w", c.ID, e)
		}
		seen[entry.ID] = true
		paths[entry.Path] = true
		cases = append(cases, c)
	}
	return m, cases, nil
}
func ValidateShape(c Case, kind string) error {
	switch c.Operation {
	case "dispatch", "crypto":
		if c.Semantic == "" || c.Root == "" || c.Crypto == nil || c.Graph != nil || c.Size != nil || c.Bootstrap != nil {
			return fmt.Errorf("missing envelope or mixed case payload")
		}
	case "provenance":
		if c.Input == nil || c.Root == "" || c.Crypto != nil || c.Graph != nil || c.Size != nil || c.Bootstrap != nil {
			return fmt.Errorf("invalid provenance payload")
		}
	case "size":
		if c.Input == nil || c.Size == nil || c.Crypto != nil || c.Graph != nil || c.Bootstrap != nil {
			return fmt.Errorf("invalid size payload")
		}
	case "bootstrap":
		if c.Bootstrap == nil || c.Root == "" || c.Crypto != nil || c.Graph != nil || c.Size != nil {
			return fmt.Errorf("invalid bootstrap payload")
		}
	case "graph":
		if kind != "semantic" || c.Graph == nil || len(c.Graph.Steps) == 0 || c.Crypto != nil || c.Size != nil || c.Bootstrap != nil || c.Input != nil {
			return fmt.Errorf("invalid graph payload")
		}
	default:
		return fmt.Errorf("unknown operation %q", c.Operation)
	}
	return nil
}
func VerifyProfile(root string, m Manifest) error {
	p := Profile{}
	b, e := os.ReadFile(filepath.Join(root, "requirements/v1-pre-rc.json"))
	if e != nil {
		return e
	}
	if e = Decode(b, &p); e != nil {
		return e
	}
	if p.Format != "totipo-requirements-v1" || p.Status != "moving-pre-rc" || p.Protocol != "totipo-v1" || p.Revision != "r9" || len(p.Required) != len(m.Cases) {
		return fmt.Errorf("invalid moving profile")
	}
	for file, want := range map[string]string{"vectors/manifest.json": p.ManifestSHA256, "spec/totipo-vault-format-v1.md": p.SpecSHA256, "vectors/manifest.schema.json": p.SchemaSHA256, "vectors/case.schema.json": p.CaseSchemaSHA256} {
		b, e := os.ReadFile(filepath.Join(root, file))
		if e != nil {
			return e
		}
		if Hash(b) != want {
			return fmt.Errorf("profile checksum mismatch: %s", file)
		}
	}
	for i, entry := range m.Cases {
		if p.Required[i].ID != entry.ID || p.Required[i].SHA256 != entry.SHA256 {
			return fmt.Errorf("profile case mismatch at %d", i)
		}
	}
	return nil
}
func Run(c Case) error {
	switch c.Operation {
	case "dispatch", "crypto":
		return runEnvelope(c)
	case "size":
		n, e := c.Input.ReservedSize()
		if e != nil {
			return e
		}
		f, e := c.Input.FanIn()
		if e != nil {
			return e
		}
		if n != c.Size.Reserved || f != c.Size.FanIn || (n <= object.Capacity) != c.Size.Fits {
			return fmt.Errorf("capacity mismatch")
		}
		want := "FOLD"
		if c.Size.Fits {
			want = "FITS"
		}
		if c.Expected != want {
			return fmt.Errorf("size outcome mismatch")
		}
		if len(c.Input.Signature) > 0 {
			root, e := unhex(c.Root)
			if e != nil {
				return e
			}
			k, e := cryptov1.Derive(root)
			if e != nil {
				return e
			}
			pub, e := unhex(c.PublicKey)
			if e != nil {
				return e
			}
			if k.Provenance(*c.Input, pub) != "VERIFIED" {
				return fmt.Errorf("capacity fixture signature invalid")
			}
			p, e := c.Input.Encode()
			if e != nil {
				return e
			}
			if len(p) > object.Capacity || c.Size.Fits {
				return fmt.Errorf("short DER fixture must physically fit but fail planning")
			}
		}
		return nil
	case "provenance":
		r, e := unhex(c.Root)
		if e != nil {
			return e
		}
		k, e := cryptov1.Derive(r)
		if e != nil {
			return e
		}
		p, e := unhex(c.PublicKey)
		if e != nil {
			return e
		}
		if k.Provenance(*c.Input, p) != c.Expected {
			return fmt.Errorf("provenance mismatch")
		}
		return nil
	case "bootstrap":
		return runBootstrap(c)
	case "graph":
		return runGraph(c)
	}
	return fmt.Errorf("unknown operation")
}
func runEnvelope(c Case) error {
	root, e := unhex(c.Root)
	if e != nil {
		return e
	}
	k, e := cryptov1.Derive(root)
	if e != nil {
		return e
	}
	p, e := unhex(c.Semantic)
	if e != nil {
		return e
	}
	if c.Future != nil {
		prefix, e := c.Future.Routing.Encode()
		if e != nil {
			return e
		}
		tail, e := unhex(c.Future.Tail)
		if e != nil {
			return e
		}
		if c.Future.Routing.Version == 1 || !bytes.Equal(append(prefix, tail...), p) {
			return fmt.Errorf("future fixture fields mismatch")
		}
	}
	x := c.Crypto
	id, b, e := k.Seal(p)
	if e != nil {
		return e
	}
	if id != x.ObjectID {
		return fmt.Errorf("object ID mismatch")
	}
	rawID, _ := unhex(id)
	plain, _ := cryptov1.Padded(p)
	checks := []struct {
		name, want string
		got        []byte
	}{
		{"object", x.Object, b}, {"id key", x.IDKey, k.ID}, {"object root", x.ObjectRootKey, k.ObjectRoot}, {"signature context", x.SignatureContext, k.SignatureContext}, {"object key", x.ObjectKey, k.ObjectKey(rawID)}, {"nonce", x.Nonce, rawID[:12]}, {"AAD", x.AAD, cryptov1.AAD(rawID)}, {"padding", x.Padded, plain}, {"ciphertext", x.Ciphertext, b[:1008]}, {"tag", x.Tag, b[1008:]},
	}
	for _, check := range checks {
		if e := equalHex(check.name, check.want, check.got); e != nil {
			return e
		}
	}
	if x.SemanticLength != len(p) {
		return fmt.Errorf("semantic length mismatch")
	}
	fixture, e := unhex(x.Object)
	if e != nil {
		return e
	}
	opened, e := k.Open(x.ObjectID, fixture)
	if e != nil {
		return e
	}
	if !bytes.Equal(opened, p) {
		return fmt.Errorf("roundtrip mismatch")
	}
	class, o := object.Dispatch(opened)
	if class != c.Expected {
		return fmt.Errorf("dispatch: got %s want %s", class, c.Expected)
	}
	if c.Input != nil {
		encoded, e := c.Input.Encode()
		if e != nil {
			return e
		}
		if !bytes.Equal(encoded, p) {
			return fmt.Errorf("field encoding mismatch")
		}
		if !reflect.DeepEqual(o, c.Input) { // nil and empty byte arrays are semantically identical; compare canonical encodings.
			parsed, e := o.Encode()
			if e != nil || !bytes.Equal(parsed, encoded) {
				return fmt.Errorf("parsed object mismatch")
			}
		}
	}
	if x.Unsigned != "" {
		if o == nil || class != object.Supported {
			return fmt.Errorf("signed fixture not supported")
		}
		u, e := o.Unsigned()
		if e != nil {
			return e
		}
		if e = equalHex("unsigned", x.Unsigned, u); e != nil {
			return e
		}
		in, e := k.SignatureInput(*o)
		if e != nil {
			return e
		}
		if e = equalHex("signature input", x.SignatureInput, in); e != nil {
			return e
		}
		if e = equalHex("signature", x.Signature, o.Signature); e != nil {
			return e
		}
		pub, e := unhex(x.PublicKey)
		if e != nil {
			return e
		}
		if k.Provenance(*o, pub) != "VERIFIED" {
			return fmt.Errorf("fixed signature rejected")
		}
	}
	return nil
}
func runBootstrap(c Case) error {
	x := c.Bootstrap
	decode := func(s string) []byte { b, _ := unhex(s); return b }
	for _, s := range []string{c.Root, x.Password, x.Salt, x.Nonce, x.WrapKey, x.Header, x.Record, x.Binding} {
		if _, e := unhex(s); e != nil {
			return e
		}
	}
	root := decode(c.Root)
	record, e := cryptov1.Wrap(decode(x.Password), root, decode(x.Salt), decode(x.Nonce))
	if e != nil {
		return e
	}
	if c.Expected != "VALID" {
		return fmt.Errorf("bootstrap expectation")
	}
	key, e := cryptov1.WrapKey(decode(x.Password), decode(x.Salt))
	if e != nil {
		return e
	}
	binding, e := cryptov1.Binding(root)
	if e != nil {
		return e
	}
	for _, v := range []struct {
		name, want string
		got        []byte
	}{{"record", x.Record, record}, {"header", x.Header, record[:39]}, {"wrap key", x.WrapKey, key}, {"binding", x.Binding, binding}} {
		if e = equalHex(v.name, v.want, v.got); e != nil {
			return e
		}
	}
	opened, e := cryptov1.Unwrap(decode(x.Password), decode(x.Record))
	if e != nil {
		return e
	}
	if !bytes.Equal(opened, root) {
		return fmt.Errorf("root unwrap mismatch")
	}
	return nil
}
func runGraph(c Case) error {
	s := graph.New()
	checks := 0
	for i, step := range c.Graph.Steps {
		var e error
		switch step.Action {
		case "learn", "persist-fails":
			if step.Node == nil {
				return fmt.Errorf("missing node")
			}
			e = s.Learn(*step.Node, step.Value, step.Action == "learn")
		case "disappear":
			s.Disappear(step.ID)
		case "discovery-incomplete":
			s.DiscoveryIncomplete = step.Flag
		case "continuity-unknown":
			s.ContinuityUnknown = step.Flag
		case "query":
			if step.Query == nil || step.Want == nil {
				return fmt.Errorf("missing graph expectation")
			}
			q := step.Query
			before, _ := json.Marshal(s)
			got := s.Evaluate(q.Identity, q.Candidate, q.Device)
			after, _ := json.Marshal(s)
			if !bytes.Equal(before, after) {
				return fmt.Errorf("query mutated state")
			}
			if !reflect.DeepEqual(got, *step.Want) {
				return fmt.Errorf("step %d: got %+v want %+v", i, got, *step.Want)
			}
			checks++
		default:
			return fmt.Errorf("unknown graph action")
		}
		if (e != nil) != step.IntegrityError {
			return fmt.Errorf("step %d integrity error: %v", i, e)
		}
	}
	if checks == 0 || c.Expected != "PASS" {
		return fmt.Errorf("no graph assertions or invalid expectation")
	}
	return nil
}
