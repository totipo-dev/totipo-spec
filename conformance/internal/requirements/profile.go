// Package requirements validates reviewed, versioned conformance targets.
package requirements

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"strings"

	"totipo/conformance/internal/corpus"
)

type Profile struct {
	Profile      string       `json:"profile"`
	Protocol     string       `json:"protocol"`
	Spec         Spec         `json:"spec"`
	Vectors      Vectors      `json:"vectors"`
	Requirements Requirements `json:"requirements"`
}

type Spec struct {
	Path     string `json:"path"`
	Revision int    `json:"revision"`
	SHA256   string `json:"sha256"`
}

type Vectors struct {
	Root           string `json:"root"`
	ManifestPath   string `json:"manifest_path"`
	ManifestSHA256 string `json:"manifest_sha256"`
}

type Requirements struct {
	CaseIDs        []string       `json:"case_ids"`
	Expected       Counts         `json:"expected"`
	CategoryCounts map[string]int `json:"category_counts,omitempty"`
}

type Counts struct {
	Pass    int `json:"pass"`
	Fail    int `json:"fail"`
	Blocked int `json:"blocked"`
}

var namePattern = regexp.MustCompile(`^[a-z0-9][a-z0-9-]*$`)
var hashPattern = regexp.MustCompile(`^[0-9a-f]{64}$`)
var idPattern = regexp.MustCompile(`^v0/[a-z0-9-]+/[a-z0-9-]+$`)

func relativePath(s string) bool {
	return s != "." && fs.ValidPath(s) && !strings.ContainsAny(s, `\:`)
}

// Load rejects unknown fields, duplicate JSON keys, nulls and trailing data.
func Load(path string) (*Profile, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	d := json.NewDecoder(bytes.NewReader(b))
	if err := uniqueJSON(d); err != nil {
		return nil, err
	}
	if _, err := d.Token(); err != io.EOF {
		return nil, errors.New("trailing profile JSON")
	}
	d = json.NewDecoder(bytes.NewReader(b))
	d.DisallowUnknownFields()
	var p Profile
	if err := d.Decode(&p); err != nil {
		return nil, err
	}
	// Integers default to zero when omitted. Require all result fields explicitly.
	var fields struct {
		Requirements struct {
			Expected map[string]int `json:"expected"`
		} `json:"requirements"`
	}
	if err := json.Unmarshal(b, &fields); err != nil {
		return nil, err
	}
	if len(fields.Requirements.Expected) != 3 {
		return nil, errors.New("expected requires pass, fail and blocked")
	}
	if err := p.Validate(); err != nil {
		return nil, err
	}
	return &p, nil
}

func uniqueJSON(d *json.Decoder) error {
	t, err := d.Token()
	if err != nil {
		return err
	}
	if t == nil {
		return errors.New("null is not part of the profile schema")
	}
	switch t {
	case json.Delim('{'):
		seen := map[string]bool{}
		for d.More() {
			key, err := d.Token()
			if err != nil {
				return err
			}
			s, ok := key.(string)
			if !ok || seen[s] {
				return errors.New("duplicate/invalid profile JSON key")
			}
			seen[s] = true
			if err := uniqueJSON(d); err != nil {
				return err
			}
		}
		_, err = d.Token()
	case json.Delim('['):
		for d.More() {
			if err := uniqueJSON(d); err != nil {
				return err
			}
		}
		_, err = d.Token()
	}
	return err
}

func (p *Profile) Validate() error {
	if !namePattern.MatchString(p.Profile) || p.Protocol != "Totipo Vault Format v0" {
		return errors.New("invalid profile name or protocol")
	}
	if !relativePath(p.Spec.Path) || !relativePath(p.Vectors.Root) || !relativePath(p.Vectors.ManifestPath) {
		return errors.New("profile paths must be repository-relative slash paths")
	}
	if p.Spec.Revision <= 0 || !hashPattern.MatchString(p.Spec.SHA256) || !hashPattern.MatchString(p.Vectors.ManifestSHA256) {
		return errors.New("invalid revision or SHA-256")
	}
	r := p.Requirements
	if len(r.CaseIDs) == 0 || r.Expected != (Counts{Pass: len(r.CaseIDs)}) {
		return errors.New("expected counts must require every ID to pass with zero fail/blocked")
	}
	for i, id := range r.CaseIDs {
		if !idPattern.MatchString(id) {
			return fmt.Errorf("invalid required ID %q", id)
		}
		if i > 0 && r.CaseIDs[i-1] >= id {
			return errors.New("required IDs must be sorted and unique")
		}
	}
	if r.CategoryCounts != nil && !reflect.DeepEqual(r.CategoryCounts, CategoryCounts(r.CaseIDs)) {
		return errors.New("category counts do not match required IDs")
	}
	return nil
}

func CategoryCounts(ids []string) map[string]int {
	counts := map[string]int{}
	for _, id := range ids {
		parts := strings.Split(id, "/")
		if len(parts) == 3 {
			counts[parts[1]]++
		}
	}
	return counts
}

// Select preserves profile order and ignores extra IDs, but rejects duplicates
// anywhere in the supplied corpus, consistent with corpus.Load.
func (p *Profile) Select(cases []corpus.Case) ([]corpus.Case, error) {
	if err := p.Validate(); err != nil {
		return nil, err
	}
	byID := make(map[string]corpus.Case, len(cases))
	for _, c := range cases {
		if _, exists := byID[c.ID]; exists {
			return nil, fmt.Errorf("duplicate corpus ID %s", c.ID)
		}
		byID[c.ID] = c
	}
	selected := make([]corpus.Case, 0, len(p.Requirements.CaseIDs))
	for _, id := range p.Requirements.CaseIDs {
		c, exists := byID[id]
		if !exists {
			return nil, fmt.Errorf("missing required case %s", id)
		}
		selected = append(selected, c)
	}
	return selected, nil
}

func HashFile(path string) (string, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%x", sha256.Sum256(b)), nil
}

func verifyHash(path, want string) error {
	got, err := HashFile(path)
	if err != nil {
		return err
	}
	if got != want {
		return fmt.Errorf("SHA-256 mismatch for %s: got %s, want %s", path, got, want)
	}
	return nil
}

// RepositoryRoot resolves profiles kept directly under the requirements directory.
// Artifact paths are relative to that repository, never the caller's directory.
func RepositoryRoot(profilePath string) (string, error) {
	abs, err := filepath.Abs(profilePath)
	if err != nil {
		return "", err
	}
	if filepath.Base(filepath.Dir(abs)) != "requirements" {
		return "", errors.New("profile must reside directly under requirements/")
	}
	return filepath.Dir(filepath.Dir(abs)), nil
}
