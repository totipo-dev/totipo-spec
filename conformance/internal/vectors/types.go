// Package vectors consumes the language-neutral r18 corpus.
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
type Case struct {
	Format    string         `json:"format"`
	ID        string         `json:"id"`
	Operation string         `json:"operation"`
	Expected  string         `json:"expected"`
	Notes     string         `json:"notes,omitempty"`
	Input     *object.Object `json:"input,omitempty"`
	Semantic  string         `json:"semantic_hex,omitempty"`
	Root      string         `json:"root_hex,omitempty"`
	PostAEAD  *PostAEAD      `json:"post_aead,omitempty"`
	Crypto    *Crypto        `json:"crypto,omitempty"`
	Bootstrap *Bootstrap     `json:"bootstrap,omitempty"`
	TOTP      *TOTP          `json:"totp,omitempty"`
	Graph     *GraphCase     `json:"graph,omitempty"`
	Fold      *FoldCase      `json:"fold,omitempty"`
	Storage   *StorageCase   `json:"storage,omitempty"`
	Workflow  *Workflow      `json:"workflow,omitempty"`
}
type Crypto struct {
	ObjectID       string `json:"object_id"`
	IDKey          string `json:"id_key_hex"`
	ObjectRootKey  string `json:"object_root_key_hex"`
	ObjectKey      string `json:"object_key_hex"`
	Nonce          string `json:"nonce_hex"`
	AAD            string `json:"aad_hex"`
	SemanticLength int    `json:"semantic_length"`
	Padded         string `json:"padded_plaintext_hex"`
	Ciphertext     string `json:"ciphertext_hex"`
	Tag            string `json:"gcm_tag_hex"`
	Object         string `json:"object_hex"`
}
type Bootstrap struct {
	Password    string `json:"password_hex"`
	Salt        string `json:"salt_hex"`
	Nonce       string `json:"nonce_hex"`
	WrapKey     string `json:"wrap_key_hex"`
	Header      string `json:"header_hex"`
	Record      string `json:"record_hex"`
	Fingerprint string `json:"fingerprint_hex"`
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
type GraphCase struct {
	Steps []Step `json:"steps"`
}
type Step struct {
	Action         string        `json:"action"`
	Node           *graph.Node   `json:"node,omitempty"`
	ID             string        `json:"id,omitempty"`
	Identity       string        `json:"identity,omitempty"`
	Want           *graph.Result `json:"expect,omitempty"`
	IntegrityError bool          `json:"integrity_error,omitempty"`
}
type FoldCase struct {
	Token    object.Object `json:"token"`
	Frontier []string      `json:"frontier"`
	StageIDs []string      `json:"stage_ids"`
	Parents  [][]string    `json:"parents"`
}
type StorageEntry struct {
	Path       string `json:"path"`
	Kind       string `json:"kind"`
	Object     string `json:"object_hex,omitempty"`
	Unreadable bool   `json:"unreadable,omitempty"`
}
type StorageCase struct {
	NamespaceKind string         `json:"namespace_kind"`
	Entries       []StorageEntry `json:"entries"`
	Classes       []string       `json:"classes"`
	Diagnostics   bool           `json:"diagnostics"`
}
type Workflow struct {
	Action           string `json:"action"`
	Kind             string `json:"kind"`
	Existing         string `json:"existing_hex"`
	Intended         string `json:"intended_hex"`
	Base             string `json:"base_hex"`
	Readable         bool   `json:"readable"`
	Complete         bool   `json:"complete"`
	Durable          bool   `json:"durable"`
	OrphanObjects    bool   `json:"orphan_objects"`
	ParentsAvailable bool   `json:"parents_available"`
	Result           string `json:"result"`
}

// PostAEAD contains an intentionally malformed but genuinely authenticated
// encryption plaintext. Defect identifies the single failed envelope boundary.
type PostAEAD struct {
	ObjectID  string `json:"object_id"`
	Plaintext string `json:"encryption_plaintext_hex"`
	Object    string `json:"object_hex"`
	Defect    string `json:"defect"`
}
