package requirements

import (
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"

	"totipo/conformance/internal/corpus"
)

// Candidate derives reviewed metadata from the existing loader. It never writes
// artifacts or runs a vector generator. The returned manifest is an exact copy
// of the current corpus manifest, intended to be frozen alongside the profile.
func Candidate(repoRoot, name string) (*Profile, []byte, error) {
	if !namePattern.MatchString(name) {
		return nil, nil, errors.New("invalid profile name")
	}
	const specPath = "spec/totipo-vault-format-v0.md"
	const vectorRoot = "vectors/v0"
	cases, err := corpus.Load(filepath.Join(repoRoot, vectorRoot))
	if err != nil {
		return nil, nil, err
	}
	spec, err := os.ReadFile(filepath.Join(repoRoot, specPath))
	if err != nil {
		return nil, nil, err
	}
	m := regexp.MustCompile(`(?m)^\*\*Status:\*\* [^\r\n]*revision ([1-9][0-9]*)[ \t]*$`).FindSubmatch(spec)
	if m == nil {
		return nil, nil, errors.New("cannot derive canonical spec revision from Status header")
	}
	revision, err := strconv.Atoi(string(m[1]))
	if err != nil {
		return nil, nil, err
	}
	manifest, err := os.ReadFile(filepath.Join(repoRoot, vectorRoot, "manifest.sha256"))
	if err != nil {
		return nil, nil, err
	}
	ids := make([]string, len(cases))
	for i, c := range cases {
		ids[i] = c.ID
	}
	p := &Profile{
		Profile: name, Protocol: "Totipo Vault Format v0",
		Spec: Spec{Path: specPath, Revision: revision, SHA256: fmt.Sprintf("%x", sha256.Sum256(spec))},
		Vectors: Vectors{
			Root: vectorRoot, ManifestPath: "requirements/" + name + ".manifest.sha256",
			ManifestSHA256: fmt.Sprintf("%x", sha256.Sum256(manifest)),
		},
		Requirements: Requirements{CaseIDs: ids, Expected: Counts{Pass: len(ids)}, CategoryCounts: CategoryCounts(ids)},
	}
	return p, manifest, p.Validate()
}

func (p *Profile) JSON() ([]byte, error) {
	if err := p.Validate(); err != nil {
		return nil, err
	}
	b, err := json.MarshalIndent(p, "", "  ")
	if err != nil {
		return nil, err
	}
	return append(b, '\n'), nil
}

// WriteCandidate refuses to overwrite either release artifact, even before a
// release is tagged. Corrections require deliberate maintainer handling.
func WriteCandidate(repoRoot string, p *Profile, manifest []byte) error {
	b, err := p.JSON()
	if err != nil {
		return err
	}
	manifestRel := "requirements/" + p.Profile + ".manifest.sha256"
	if p.Vectors.ManifestPath != manifestRel || fmt.Sprintf("%x", sha256.Sum256(manifest)) != p.Vectors.ManifestSHA256 {
		return errors.New("candidate manifest path/hash mismatch")
	}
	paths := []string{filepath.Join(repoRoot, manifestRel), filepath.Join(repoRoot, "requirements", p.Profile+".json")}
	for _, path := range paths {
		if _, err := os.Lstat(path); !errors.Is(err, os.ErrNotExist) {
			if err != nil {
				return err
			}
			return fmt.Errorf("refusing to overwrite %s", path)
		}
	}
	if err := os.MkdirAll(filepath.Join(repoRoot, "requirements"), 0755); err != nil {
		return err
	}
	for i, data := range [][]byte{manifest, b} {
		f, err := os.OpenFile(paths[i], os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0644)
		if err != nil {
			return err
		}
		_, writeErr := f.Write(data)
		closeErr := f.Close()
		if err := errors.Join(writeErr, closeErr); err != nil {
			return err
		}
	}
	return nil
}
