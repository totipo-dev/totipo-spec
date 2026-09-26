// Package vectors defines and consumes the language-neutral JSON case format.
package vectors

import (
	"totipo/conformance/internal/graph"
	"totipo/conformance/internal/object"
)

type Entry struct {
	ID        string   `json:"id"`
	Category  string   `json:"category"`
	Kind      string   `json:"kind"`
	Normative bool     `json:"normative"`
	Path      string   `json:"path"`
	Expected  string   `json:"expected"`
	Sections  []string `json:"spec_sections"`
	SHA256    string   `json:"sha256"`
}
type Manifest struct {
	Format   string  `json:"format"`
	Protocol string  `json:"protocol"`
	Revision string  `json:"spec_revision"`
	Cases    []Entry `json:"cases"`
}
type Profile struct {
	Format           string `json:"format"`
	Status           string `json:"status"`
	Protocol         string `json:"protocol"`
	Revision         string `json:"spec_revision"`
	ManifestSHA256   string `json:"manifest_sha256"`
	SpecSHA256       string `json:"spec_sha256"`
	SchemaSHA256     string `json:"schema_sha256"`
	CaseSchemaSHA256 string `json:"case_schema_sha256"`
	Required         []Pin  `json:"required_cases"`
}
type Pin struct {
	ID     string `json:"id"`
	SHA256 string `json:"sha256"`
}
type TOTP struct {
	Source    string    `json:"source"`
	Notes     string    `json:"notes"`
	Algorithm byte      `json:"algorithm"`
	Digits    byte      `json:"digits"`
	Period    uint32    `json:"period"`
	T0        uint64    `json:"t0"`
	Secret    string    `json:"secret_hex"`
	Rows      []TOTPRow `json:"rows"`
}
type TOTPRow struct {
	UnixSeconds uint64 `json:"unix_time_seconds"`
	Counter     uint64 `json:"counter"`
	CounterHex  string `json:"counter_hex"`
	Code        string `json:"code"`
}
type Case struct {
	TOTP      *TOTP          `json:"totp,omitempty"`
	Format    string         `json:"format"`
	ID        string         `json:"id"`
	Operation string         `json:"operation"`
	Expected  string         `json:"expected"`
	Input     *object.Object `json:"input,omitempty"`
	Semantic  string         `json:"semantic_hex,omitempty"`
	Root      string         `json:"root_hex,omitempty"`
	PublicKey string         `json:"public_key_hex,omitempty"`
	Crypto    *Crypto        `json:"crypto,omitempty"`
	Future    *Future        `json:"future,omitempty"`
	Size      *Size          `json:"size,omitempty"`
	Bootstrap *Bootstrap     `json:"bootstrap,omitempty"`
	Graph     *GraphCase     `json:"graph,omitempty"`
}
type Future struct {
	Routing object.Routing `json:"routing"`
	Tail    string         `json:"opaque_tail_hex"`
}

type Crypto struct {
	PrivateKey       string `json:"fixture_private_key_hex,omitempty"`
	PublicKey        string `json:"fixture_public_key_hex,omitempty"`
	Unsigned         string `json:"unsigned_semantic_hex,omitempty"`
	SignatureInput   string `json:"signature_input_hex,omitempty"`
	Signature        string `json:"signature_der_hex,omitempty"`
	ObjectID         string `json:"object_id"`
	IDKey            string `json:"id_key_hex"`
	ObjectRootKey    string `json:"object_root_key_hex"`
	SignatureContext string `json:"signature_context_hex"`
	ObjectKey        string `json:"object_key_hex"`
	Nonce            string `json:"nonce_hex"`
	AAD              string `json:"aad_hex"`
	SemanticLength   int    `json:"semantic_length"`
	Padded           string `json:"padded_plaintext_hex"`
	Ciphertext       string `json:"ciphertext_hex"`
	Tag              string `json:"gcm_tag_hex"`
	Object           string `json:"object_hex"`
}
type Size struct {
	Reserved int  `json:"reserved_bytes"`
	FanIn    int  `json:"fan_in"`
	Fits     bool `json:"fits"`
}
type Bootstrap struct {
	Password string `json:"password_hex"`
	Salt     string `json:"salt_hex"`
	Nonce    string `json:"nonce_hex"`
	WrapKey  string `json:"wrap_key_hex"`
	Header   string `json:"header_hex"`
	Record   string `json:"record_hex"`
	Binding  string `json:"binding_hex"`
}
type GraphCase struct {
	Steps []Step `json:"steps"`
}
type Query struct {
	Identity  string `json:"identity"`
	Candidate string `json:"candidate"`
	Device    string `json:"device,omitempty"`
}
type Step struct {
	Action         string        `json:"action"`
	Node           *graph.Node   `json:"node,omitempty"`
	Value          *graph.Value  `json:"value,omitempty"`
	ID             string        `json:"id,omitempty"`
	Flag           bool          `json:"flag,omitempty"`
	Query          *Query        `json:"query,omitempty"`
	Want           *graph.Result `json:"expect,omitempty"`
	IntegrityError bool          `json:"integrity_error,omitempty"`
}
