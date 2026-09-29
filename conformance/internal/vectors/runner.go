package vectors

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
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
	"totipo/conformance/internal/storage"
	"totipo/conformance/internal/totp"
)

// JSON objects have unique member names at every depth; rejecting duplicates
// avoids consumer-dependent interpretations before decoding typed payloads.
func uniqueJSON(p []byte) error {
	d := json.NewDecoder(bytes.NewReader(p))
	d.UseNumber()
	var walk func() error
	walk = func() error {
		t, e := d.Token()
		if e != nil {
			return e
		}
		delim, ok := t.(json.Delim)
		if !ok {
			return nil
		}
		switch delim {
		case '{':
			seen := map[string]bool{}
			for d.More() {
				key, e := d.Token()
				if e != nil {
					return e
				}
				s, ok := key.(string)
				if !ok || seen[s] {
					return fmt.Errorf("duplicate/invalid JSON member")
				}
				seen[s] = true
				if e := walk(); e != nil {
					return e
				}
			}
		case '[':
			for d.More() {
				if e := walk(); e != nil {
					return e
				}
			}
		default:
			return fmt.Errorf("unexpected delimiter")
		}
		_, e = d.Token()
		return e
	}
	return walk()
}
func Decode(p []byte, v any) error {
	if e := uniqueJSON(p); e != nil {
		return e
	}
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
	if m.Format != "totipo-vector-manifest-v1" || m.Protocol != "totipo-v1" || m.Revision != "r17" || len(m.Cases) == 0 {
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
	// Every physical case must be represented once; unlisted cases are not silently skipped.
	if err := filepath.WalkDir(filepath.Join(root, "vectors/cases"), func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		rel, err := filepath.Rel(filepath.Join(root, "vectors"), path)
		if err != nil {
			return err
		}
		if !paths[filepath.ToSlash(rel)] {
			return fmt.Errorf("unlisted physical case %s", rel)
		}
		return nil
	}); err != nil {
		return m, nil, err
	}

	return m, cases, nil
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
	if p.Format != "totipo-requirements-v1" || p.Status != "moving-pre-rc" || p.Protocol != "totipo-v1" || p.Revision != "r17" {
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
	want := map[string]string{}
	for _, entry := range m.Cases {
		want[entry.ID] = entry.SHA256
	}
	if len(p.Required) != len(want) {
		return fmt.Errorf("profile case count mismatch")
	}
	for _, pin := range p.Required {
		if want[pin.ID] != pin.SHA256 {
			return fmt.Errorf("profile case mismatch %s", pin.ID)
		}
		delete(want, pin.ID)
	}
	return nil
}
func ValidateShape(c Case, kind string) error {
	payloads := 0
	for _, present := range []bool{c.PostAEAD != nil, c.Crypto != nil, c.Bootstrap != nil, c.TOTP != nil, c.Graph != nil, c.Fold != nil, c.Storage != nil, c.Workflow != nil} {
		if present {
			payloads++
		}
	}
	if payloads != 1 {
		return fmt.Errorf("exactly one operation payload required")
	}
	ok := false
	switch c.Operation {
	case "crypto", "dispatch":
		if (c.Operation == "crypto") != (c.Input != nil) {
			return fmt.Errorf("input/operation mismatch")
		}
		ok = c.Crypto != nil && c.Root != "" && c.Semantic != "" && (c.Expected == object.Supported || c.Expected == object.Invalid)
	case "post-aead":
		ok = c.PostAEAD != nil && c.Root != "" && c.Semantic != "" && c.Input == nil && c.Expected == storage.InvalidStorage
	case "bootstrap":
		ok = c.Bootstrap != nil && c.Root != "" && c.Expected == "VALID"
	case "totp":
		ok = c.TOTP != nil && len(c.TOTP.Rows) > 0 && c.TOTP.T0 == 0 && c.Expected == "PASS"
	case "graph":
		ok = c.Graph != nil && len(c.Graph.Steps) > 0 && c.Expected == "PASS"
	case "fold":
		ok = c.Fold != nil && len(c.Fold.StageIDs) > 0 && len(c.Fold.StageIDs) == len(c.Fold.Parents) && c.Expected == "PASS"
	case "storage":
		ok = c.Storage != nil && c.Root != "" && c.Expected == "PASS"
	case "workflow":
		ok = c.Workflow != nil && c.Expected == "PASS"
	}
	if !ok {
		return fmt.Errorf("invalid %s payload", c.Operation)
	}
	if (c.Input != nil || c.Semantic != "") && c.Crypto == nil && c.PostAEAD == nil {
		return fmt.Errorf("unexpected semantic input")
	}
	if c.Root != "" && c.Crypto == nil && c.Bootstrap == nil && c.Storage == nil && c.PostAEAD == nil {
		return fmt.Errorf("unexpected root")
	}
	wantKind := "semantic"
	if c.Crypto != nil || c.Bootstrap != nil || c.TOTP != nil {
		wantKind = "bytes"
	}
	if c.Expected == object.Invalid || c.PostAEAD != nil {
		wantKind = "negative"
	}
	if kind != wantKind {
		return fmt.Errorf("wrong kind")
	}
	return nil
}
func Run(c Case) error {
	switch c.Operation {
	case "post-aead":
		return runPostAEAD(c)
	case "crypto", "dispatch":
		return runEnvelope(c)
	case "bootstrap":
		return runBootstrap(c)
	case "totp":
		return runTOTP(c)
	case "graph":
		s := graph.New()
		for _, step := range c.Graph.Steps {
			switch step.Action {
			case "add":
				if step.Node == nil {
					return fmt.Errorf("missing node")
				}
				e := s.Add(*step.Node)
				if (e != nil) != step.IntegrityError {
					return fmt.Errorf("integrity result")
				}
			case "remove":
				s.Remove(step.ID)
			case "evaluate":
				if step.Want == nil || !reflect.DeepEqual(s.Evaluate(step.Identity), *step.Want) {
					return fmt.Errorf("graph: got %+v want %+v", s.Evaluate(step.Identity), step.Want)
				}
			default:
				return fmt.Errorf("unknown graph action")
			}
		}
		return nil
	case "fold":
		f := c.Fold
		ids := [][]byte{}
		for _, id := range f.Frontier {
			b, e := unhex(id)
			if e != nil {
				return e
			}
			ids = append(ids, b)
		}
		i := 0
		_, e := graph.FoldObjects(ids, f.Token, func(stage object.Object) ([]byte, error) {
			parents := stage.Parents
			p, e := stage.Encode()
			if e != nil {
				return nil, e
			}
			_, decoded := object.Dispatch(p)
			if !bytes.Equal(decoded.ValueBytes(), f.Token.ValueBytes()) || !bytes.Equal(decoded.Identity, f.Token.Identity) ||
				!reflect.DeepEqual(decoded.ClientName, f.Token.ClientName) || !reflect.DeepEqual(decoded.ClientTime, f.Token.ClientTime) {
				return nil, fmt.Errorf("fold changed operation value or metadata")
			}
			if i >= len(f.Parents) {
				return nil, fmt.Errorf("extra stage")
			}
			got := []string{}
			for _, p := range parents {
				got = append(got, hex.EncodeToString(p))
			}
			if !reflect.DeepEqual(got, f.Parents[i]) {
				return nil, fmt.Errorf("fold parents")
			}
			id, e := unhex(f.StageIDs[i])
			i++
			return id, e
		})
		if e != nil {
			return e
		}
		if i != len(f.StageIDs) {
			return fmt.Errorf("missing stage")
		}
		return nil
	case "storage":
		root, e := unhex(c.Root)
		if e != nil {
			return e
		}
		k, e := cryptov1.Derive(root)
		if e != nil {
			return e
		}
		entries := []storage.Entry{}
		for _, entry := range c.Storage.Entries {
			b, e := unhex(entry.Object)
			if e != nil {
				return e
			}
			entries = append(entries, storage.Entry{Path: entry.Path, Kind: entry.Kind, Read: func() ([]byte, error) {
				if entry.Unreadable {
					return nil, fmt.Errorf("unreadable")
				}
				return b, nil
			}})
		}
		obs, err := storage.Scan(c.Storage.NamespaceKind, entries, k)
		classes := []string{}
		for _, o := range obs {
			classes = append(classes, o.Class)
		}
		if (err != nil) != c.Storage.Diagnostics || !reflect.DeepEqual(classes, c.Storage.Classes) {
			return fmt.Errorf("storage diagnostics/classes mismatch")
		}
		return nil
	case "workflow":
		w := c.Workflow
		existing, e := unhex(w.Existing)
		if e != nil {
			return e
		}
		intended, e := unhex(w.Intended)
		if e != nil {
			return e
		}
		base, e := unhex(w.Base)
		if e != nil {
			return e
		}
		var result string
		switch w.Action {
		case "publish":
			var after []byte
			after, result = storage.Install(existing, intended, w.Kind, w.Durable)
			if result != "PUBLISHED_NEW" && !bytes.Equal(after, existing) {
				return fmt.Errorf("existing mutated")
			}
		case "create":
			result = storage.Create(w.Kind, w.Complete, w.Durable)
		case "replace":
			result = storage.Replace(base, existing, w.Kind, w.Readable, w.Complete, w.Durable)
		default:
			return fmt.Errorf("unknown workflow")
		}
		if result != w.Result {
			return fmt.Errorf("workflow: %s != %s", result, w.Result)
		}
		return nil
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
	x := c.Crypto
	id, b, e := k.Seal(p)
	if e != nil {
		return e
	}
	if id != x.ObjectID {
		return fmt.Errorf("object ID mismatch")
	}
	rawID, _ := unhex(id)
	padded, _ := cryptov1.Padded(p)
	for _, check := range []struct {
		name, want string
		got        []byte
	}{
		{"object", x.Object, b}, {"id key", x.IDKey, k.ID}, {"object root", x.ObjectRootKey, k.ObjectRoot},
		{"object key", x.ObjectKey, k.ObjectKey(rawID)}, {"nonce", x.Nonce, rawID[:12]}, {"AAD", x.AAD, cryptov1.AAD(rawID)},
		{"padding", x.Padded, padded}, {"ciphertext", x.Ciphertext, b[:1008]}, {"tag", x.Tag, b[1008:]},
	} {
		if e := equalHex(check.name, check.want, check.got); e != nil {
			return e
		}
	}
	if len(p) != x.SemanticLength {
		return fmt.Errorf("semantic length")
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
		return fmt.Errorf("roundtrip")
	}
	class, o := object.Dispatch(opened)
	if class != c.Expected {
		return fmt.Errorf("grammar: %s != %s", class, c.Expected)
	}
	if c.Input != nil {
		encoded, e := c.Input.Encode()
		if e != nil {
			return e
		}
		if !bytes.Equal(encoded, p) {
			return fmt.Errorf("field encoding")
		}
		again, e := o.Encode()
		if e != nil || !bytes.Equal(again, p) {
			return fmt.Errorf("canonical roundtrip")
		}
	}
	return nil
}
func runBootstrap(c Case) error {
	x := c.Bootstrap
	fields := []string{c.Root, x.Password, x.Salt, x.Nonce, x.Record}
	b := [][]byte{}
	for _, s := range fields {
		v, e := unhex(s)
		if e != nil {
			return e
		}
		b = append(b, v)
	}
	root, password, salt, nonce, record := b[0], b[1], b[2], b[3], b[4]
	wrapped, e := cryptov1.Wrap(password, root, salt, nonce)
	if e != nil {
		return e
	}
	if !bytes.Equal(wrapped, record) {
		return fmt.Errorf("wrap")
	}
	recovered, e := cryptov1.Unwrap(password, record)
	if e != nil {
		return e
	}
	if !bytes.Equal(root, recovered) {
		return fmt.Errorf("unwrap")
	}
	key, e := cryptov1.WrapKey(password, salt)
	if e != nil {
		return e
	}
	fp, _ := cryptov1.Fingerprint(root)
	for _, check := range []struct {
		name, want string
		got        []byte
	}{{"wrap key", x.WrapKey, key}, {"header", x.Header, record[:39]}, {"fingerprint", x.Fingerprint, fp}} {
		if e := equalHex(check.name, check.want, check.got); e != nil {
			return e
		}
	}
	return nil
}
func runTOTP(c Case) error {
	x := c.TOTP
	secret, e := unhex(x.Secret)
	if e != nil {
		return e
	}
	for _, row := range x.Rows {
		code, e := totp.Code(x.Algorithm, secret, x.Digits, x.Period, row.UnixSeconds)
		if e != nil {
			return e
		}
		if code != row.Code || x.Period == 0 || row.Counter != row.UnixSeconds/uint64(x.Period) {
			return fmt.Errorf("TOTP mismatch")
		}
		b := make([]byte, 8)
		binary.BigEndian.PutUint64(b, row.Counter)
		if e := equalHex("counter", row.CounterHex, b); e != nil {
			return e
		}
	}
	return nil
}
