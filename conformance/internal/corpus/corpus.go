// Package corpus validates test artifacts only; it has no protocol dependencies.
package corpus

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

type Provenance struct {
	Source  string `json:"source"`
	Status  string `json:"status"`
	Locator string `json:"locator"`
}
type Case struct {
	ID           string                `json:"id"`
	Kind         string                `json:"kind"`
	Description  string                `json:"description"`
	Provenance   Provenance            `json:"provenance"`
	Input        map[string]string     `json:"input"`
	Expected     map[string]string     `json:"expected"`
	Scenario     *Scenario             `json:"scenario,omitempty"`
	Trace        []TraceEvent          `json:"trace,omitempty"`
	Presentation *PresentationScenario `json:"presentation,omitempty"`
}
type DeviceUpdate struct {
	ID      string   `json:"id"`
	Signer  string   `json:"signer"`
	Parents []string `json:"parents"`
	Name    string   `json:"name"`
	Kind    string   `json:"kind"`
}
type PresentationView struct {
	Validation map[string]string   `json:"validation"`
	Heads      map[string][]string `json:"heads"`
	Names      map[string][]string `json:"names"`
}
type PresentationScenario struct {
	Updates  []DeviceUpdate   `json:"updates"`
	Expected PresentationView `json:"expected"`
}
type TraceEvent struct {
	Op       string            `json:"op"`
	Args     map[string]string `json:"args"`
	Expected map[string]string `json:"expected"`
}

// Scenario is a symbolic post-authentication artifact, never a wire object.
type Scenario struct {
	Updates     []SymbolicUpdate    `json:"updates"`
	Expected    SymbolicView        `json:"expected"`
	Unavailable []string            `json:"unavailable,omitempty"`
	Observed    []string            `json:"observed,omitempty"`
	Coverage    map[string][]string `json:"coverage,omitempty"`
}
type SymbolicUpdate struct {
	ID      string            `json:"id"`
	Token   string            `json:"token"`
	Parents []string          `json:"parents"`
	Fields  map[string]string `json:"fields"`
	Signer  string            `json:"signer,omitempty"`
}
type SymbolicWitness struct {
	Status      string   `json:"status"`
	Credential  string   `json:"credential"`
	Transitions []string `json:"transitions"`
}
type SymbolicToken struct {
	Heads     []string            `json:"heads"`
	Fields    map[string][]string `json:"fields"`
	Conflicts []SymbolicWitness   `json:"conflicts"`
}
type SymbolicView struct {
	Validation  map[string]string        `json:"validation"`
	Tokens      map[string]SymbolicToken `json:"tokens"`
	Resolutions []string                 `json:"resolutions"`
}
type Bundle struct {
	Schema int    `json:"schema"`
	Cases  []Case `json:"cases"`
}

var idPattern = regexp.MustCompile(`^v0/[a-z0-9-]+/[a-z0-9-]+$`)
var hexPattern = regexp.MustCompile(`^(?:[0-9a-f]{2})*$`)
var decimalPattern = regexp.MustCompile(`^(0|[1-9][0-9]*)$`)
var timePattern = regexp.MustCompile(`^(0|-?[1-9][0-9]*)$`)
var codePattern = regexp.MustCompile(`^[0-9]{6,8}$`)

func Hex(s string) ([]byte, error) {
	if !hexPattern.MatchString(s) {
		return nil, errors.New("hex must be even-length lowercase without whitespace")
	}
	return hex.DecodeString(s)
}

// JSON objects cannot contain duplicate keys: encoding/json otherwise silently accepts them.
func uniqueJSON(d *json.Decoder) error {
	t, e := d.Token()
	if e != nil {
		return e
	}
	if t == nil {
		return errors.New("null is not part of the artifact schema")
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
				return errors.New("duplicate/invalid JSON key")
			}
			seen[s] = true
			if e := uniqueJSON(d); e != nil {
				return e
			}
		}
	case '[':
		for d.More() {
			if e := uniqueJSON(d); e != nil {
				return e
			}
		}
	default:
		return errors.New("unexpected delimiter")
	}
	_, e = d.Token()
	return e
}
func fields(m map[string]string, required, optional string) error {
	allowed := map[string]bool{}
	for _, k := range strings.Fields(required) {
		allowed[k] = true
		if _, ok := m[k]; !ok {
			return fmt.Errorf("missing field %s", k)
		}
	}
	for _, k := range strings.Fields(optional) {
		allowed[k] = true
	}
	for k, v := range m {
		if !allowed[k] {
			return fmt.Errorf("unknown field %s", k)
		}
		if strings.HasSuffix(k, "_hex") {
			if _, e := Hex(v); e != nil {
				return fmt.Errorf("%s: %w", k, e)
			}
		}
		switch k {
		case "algorithm", "digits", "period", "length", "kdf_calls":
			if !decimalPattern.MatchString(v) {
				return fmt.Errorf("%s must be a canonical unsigned decimal string", k)
			}
		case "seconds":
			if !timePattern.MatchString(v) {
				return errors.New("seconds must be a canonical decimal string")
			}
		case "code":
			if !codePattern.MatchString(v) {
				return errors.New("invalid expected OTP code")
			}
		case "degraded", "incomplete", "lifecycle_conflict", "credential_complete", "issuer_conflict", "account_conflict", "active", "recovery_visible", "metadata_conflict":
			if v != "true" && v != "false" {
				return fmt.Errorf("%s must be true or false", k)
			}
		}
	}
	return nil
}
func Validate(c Case) error {
	if c.Input == nil || c.Expected == nil {
		return errors.New("missing input/expected maps")
	}
	if (c.Kind == "model") != (c.Scenario != nil) {
		return errors.New("scenario required only for model cases")
	}
	if (c.Kind == "memory") != (c.Trace != nil) {
		return errors.New("trace required only for memory cases")
	}
	if (c.Kind == "presentation") != (c.Presentation != nil) {
		return errors.New("presentation scenario required only for presentation cases")
	}
	if !idPattern.MatchString(c.ID) || c.Description == "" || c.Provenance.Source == "" || c.Provenance.Locator == "" {
		return errors.New("missing/invalid identity or provenance")
	}
	switch c.Provenance.Status {
	case "reviewed-pinned-vector", "reviewed-pinned", "spec-derived", "spec-derived-reviewed", "published-standard":
	default:
		return errors.New("unknown provenance status")
	}
	var in, ex, opt, dispositions string
	switch c.Kind {
	case "blocked":
		in = "requirement"
		ex = "reason"
		dispositions = "BLOCKED"
	case "use_profile":
		in = "degraded incomplete lifecycle_conflict status_values credential_values credential_complete issuer_conflict account_conflict"
		ex = "active recovery_visible metadata_conflict"
		dispositions = "ALLOW BLOCK"
	case "presentation":
		dispositions = "VERIFIED"
		if len(c.Presentation.Updates) == 0 || c.Presentation.Expected.Validation == nil || c.Presentation.Expected.Heads == nil || c.Presentation.Expected.Names == nil {
			return errors.New("incomplete presentation scenario")
		}
		for _, u := range c.Presentation.Updates {
			if u.ID == "" || u.Signer == "" || u.Parents == nil || u.Kind == "" {
				return errors.New("incomplete device update")
			}
		}
	case "memory":
		in = "binding"
		dispositions = "VERIFIED"
		if len(c.Trace) == 0 {
			return errors.New("empty trace")
		}
		for _, e := range c.Trace {
			if e.Op == "" || e.Args == nil || e.Expected == nil || e.Expected["result"] == "" {
				return errors.New("incomplete trace event")
			}
			if err := validateTrace(e); err != nil {
				return err
			}
		}
	case "ed25519":
		in = "public_key_hex message_hex signature_hex"
		ex = "reason"
		dispositions = "ACCEPT REJECT"
	case "model":
		dispositions = "VERIFIED"
		if len(c.Scenario.Updates) == 0 || c.Scenario.Expected.Validation == nil || c.Scenario.Expected.Tokens == nil || c.Scenario.Expected.Resolutions == nil {
			return errors.New("incomplete scenario")
		}
		for _, u := range c.Scenario.Updates {
			if u.ID == "" || u.Token == "" || u.Parents == nil || u.Fields == nil {
				return errors.New("incomplete symbolic update")
			}
		}
		for _, v := range c.Scenario.Expected.Validation {
			if v != "FULLY_VALID" && v != "PENDING" && v != "INVALID" {
				return errors.New("unknown model disposition")
			}
		}
	case "tlv":
		in = "bytes_hex form"
		opt = "unsigned_hex"
		dispositions = "STRUCTURALLY_VALID STRUCTURALLY_INVALID"
		if c.Input["form"] != "signed" && c.Input["form"] != "unsigned" {
			return errors.New("unknown form")
		}
	case "envelope":
		in = "bytes_hex"
		opt = "semantic_hex semantic_disposition"
		dispositions = "ENVELOPE_VALID ENVELOPE_REJECTED"
	case "file_length":
		in = "length"
		dispositions = "FILE_LENGTH_VALID FILE_LENGTH_REJECTED"
	case "dispatch":
		in = "bytes_hex"
		dispositions = "INVALID_PREFIX AUTHENTICATED_UNSUPPORTED_FUTURE STRUCTURALLY_VALID STRUCTURALLY_INVALID"
	case "object_crypto":
		in = "root_hex signing_seed_hex unsigned_hex object_type"
		ex = "public_key_hex prk_hex k_id_hex k_object_root_hex k_signature_context_hex signature_input_hex signature_hex signed_plaintext_hex object_id_hex object_key_hex nonce_hex aad_hex envelope_plaintext_hex ciphertext_hex tag_hex file_hex semantic_length_hex signature_domain_hex"
		dispositions = "VERIFIED"
		if c.Input["object_type"] != "TOKEN_UPDATE" && c.Input["object_type"] != "DEVICE_UPDATE" {
			return errors.New("object_type")
		}
	case "bootstrap":
		in = "root_hex password_utf8_hex argon2_salt_hex wrap_nonce_hex"
		ex = "header_aad_hex k_wrap_hex wrapped_root_hex wrap_tag_hex vault_file_hex"
		dispositions = "VERIFIED"
	case "bootstrap_check":
		in = "file_hex password_utf8_hex"
		ex = "kdf_calls"
		dispositions = "FILE_LENGTH_REJECTED MAGIC_REJECTED BOOTSTRAP_VERSION_REJECTED PASSWORD_REJECTED AEAD_REJECTED VERIFIED"
		opt = "root_hex"
	case "password":
		in = "encoding value"
		dispositions = "PASSWORD_VALID PASSWORD_REJECTED"
		if c.Input["encoding"] != "utf8-hex" && c.Input["encoding"] != "codepoints" {
			return errors.New("password encoding")
		}
		if c.Input["encoding"] == "utf8-hex" {
			if _, e := Hex(c.Input["value"]); e != nil {
				return e
			}
		}
	case "totp":
		in = "secret_hex algorithm digits period seconds"
		opt = "code"
		dispositions = "VERIFIED INVALID"
	case "pipeline":
		in = "root_hex object_id_hex file_hex"
		ex = "stages"
		dispositions = "FILE_LENGTH_REJECTED AEAD_REJECTED ENVELOPE_REJECTED OBJECT_ID_REJECTED INVALID_PREFIX AUTHENTICATED_UNSUPPORTED_FUTURE STRUCTURALLY_VALID STRUCTURALLY_INVALID"
	default:
		return fmt.Errorf("unknown kind %q", c.Kind)
	}
	if !strings.Contains(" "+dispositions+" ", " "+c.Expected["disposition"]+" ") || c.Expected["disposition"] == "" {
		return errors.New("invalid disposition for kind")
	}
	if e := fields(c.Input, in, ""); e != nil {
		return e
	}
	if e := fields(c.Expected, "disposition "+ex, opt); e != nil {
		return e
	}
	if c.Kind == "totp" && c.Expected["disposition"] == "VERIFIED" {
		if _, ok := c.Expected["code"]; !ok {
			return errors.New("missing TOTP code")
		}
	}
	if c.Kind == "tlv" && c.Expected["disposition"] == "STRUCTURALLY_VALID" && c.Input["form"] == "signed" {
		if _, ok := c.Expected["unsigned_hex"]; !ok {
			return errors.New("missing unsigned expectation")
		}
	}
	if c.Kind == "envelope" && c.Expected["disposition"] == "ENVELOPE_VALID" {
		if _, ok := c.Expected["semantic_hex"]; !ok {
			return errors.New("missing semantic expectation")
		}
		s := c.Expected["semantic_disposition"]
		if s != "STRUCTURALLY_VALID" && s != "STRUCTURALLY_INVALID" {
			return errors.New("semantic disposition")
		}
	}
	return nil
}
func Decode(b []byte) ([]Case, error) {
	d := json.NewDecoder(bytes.NewReader(b))
	if e := uniqueJSON(d); e != nil {
		return nil, e
	}
	if _, e := d.Token(); e != io.EOF {
		return nil, errors.New("trailing JSON")
	}
	d = json.NewDecoder(bytes.NewReader(b))
	d.DisallowUnknownFields()
	var bundle Bundle
	if e := d.Decode(&bundle); e != nil {
		return nil, e
	}
	if bundle.Schema != 1 || len(bundle.Cases) == 0 {
		return nil, errors.New("invalid schema or empty cases")
	}
	for _, c := range bundle.Cases {
		if e := Validate(c); e != nil {
			return nil, fmt.Errorf("%s: %w", c.ID, e)
		}
	}
	return bundle.Cases, nil
}
func Load(root string) ([]Case, error) {
	files := map[string][]byte{}
	err := filepath.WalkDir(root, func(p string, d fs.DirEntry, e error) error {
		if e != nil {
			return e
		}
		if d.Type()&os.ModeSymlink != 0 {
			return fmt.Errorf("symlink not permitted: %s", p)
		}
		if d.IsDir() {
			return nil
		}
		if !d.Type().IsRegular() {
			return fmt.Errorf("nonregular artifact: %s", p)
		}
		if !strings.HasSuffix(p, ".json") {
			return nil
		}
		b, e := os.ReadFile(p)
		if e != nil {
			return e
		}
		rel, e := filepath.Rel(root, p)
		if e != nil {
			return e
		}
		files[filepath.ToSlash(rel)] = b
		return nil
	})
	if err != nil {
		return nil, err
	}
	if len(files) == 0 {
		return nil, errors.New("no vectors")
	}
	manifest, e := os.ReadFile(filepath.Join(root, "manifest.sha256"))
	if e != nil {
		return nil, e
	}
	var names []string
	for n := range files {
		names = append(names, n)
	}
	sort.Strings(names)
	var want strings.Builder
	for _, n := range names {
		h := sha256.Sum256(files[n])
		fmt.Fprintf(&want, "%x  %s\n", h, n)
	}
	if string(manifest) != want.String() {
		return nil, errors.New("stale, incomplete, or noncanonical manifest")
	}
	seen := map[string]bool{}
	var out []Case
	for _, n := range names {
		cases, e := Decode(files[n])
		if e != nil {
			return nil, fmt.Errorf("%s: %w", n, e)
		}
		for _, c := range cases {
			if seen[c.ID] {
				return nil, fmt.Errorf("duplicate ID %s", c.ID)
			}
			seen[c.ID] = true
			out = append(out, c)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out, nil
}
