package requirements

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"testing"

	"totipo/conformance/internal/corpus"
)

func copyTree(t *testing.T, src, dst string) {
	t.Helper()
	err := filepath.WalkDir(src, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		write(t, filepath.Join(dst, rel), read(t, path))
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}

func profileFixture(t *testing.T) (*Profile, string) {
	t.Helper()
	repo := t.TempDir()
	copyTree(t, "../../../vectors/v0", filepath.Join(repo, "vectors/v0"))
	copyTree(t, "../../../requirements", filepath.Join(repo, "requirements"))
	write(t, filepath.Join(repo, "spec/totipo-vault-format-v0.md"), read(t, "../../../spec/totipo-vault-format-v0.md"))
	p, err := Load(filepath.Join(repo, "requirements/v0-rc1.json"))
	if err != nil {
		t.Fatal(err)
	}
	return p, repo
}

// Only temporary fixtures are rewritten; committed vectors are never generated.
func updateManifest(t *testing.T, root string) {
	t.Helper()
	var lines []string
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || !strings.HasSuffix(path, ".json") {
			return nil
		}
		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		lines = append(lines, fmt.Sprintf("%x  %s\n", sha256.Sum256(read(t, path)), filepath.ToSlash(rel)))
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	sort.Slice(lines, func(i, j int) bool { return lines[i][66:] < lines[j][66:] })
	write(t, filepath.Join(root, "manifest.sha256"), []byte(strings.Join(lines, "")))
}

func TestReleaseArtifacts(t *testing.T) {
	p, err := Load("../../../requirements/v0-rc1.json")
	if err != nil {
		t.Fatal(err)
	}
	cases, err := p.LoadCases("../../..", "")
	if err != nil {
		t.Fatal(err)
	}
	var ids []string
	for _, c := range cases {
		ids = append(ids, c.ID)
	}
	if len(ids) != 388 || !reflect.DeepEqual(ids, p.Requirements.CaseIDs) {
		t.Fatal("release selection differs from the exact frozen 388 IDs")
	}
	// Independently recorded Phase 2 category totals guard initial derivation.
	want := map[string]int{"tlv": 65, "envelope": 13, "object-crypto": 2, "bootstrap": 37, "dispatch": 56, "totp": 82, "ed25519": 33, "lifecycle": 15, "transitions": 46, "recovery": 28, "presentation": 11}
	if !reflect.DeepEqual(CategoryCounts(ids), want) {
		t.Fatalf("category totals: %v", CategoryCounts(ids))
	}
}

func TestArtifactMutations(t *testing.T) {
	for _, name := range []string{"spec", "pinned-manifest", "current-manifest", "required-file-missing", "required-case-missing", "changed-expectation-stale-manifest", "changed-expectation-updated-manifest"} {
		t.Run(name, func(t *testing.T) {
			p, repo := profileFixture(t)
			root := filepath.Join(repo, "vectors/v0")
			switch name {
			case "spec", "pinned-manifest", "current-manifest":
				path := filepath.Join(repo, p.Spec.Path)
				if name == "pinned-manifest" {
					path = filepath.Join(repo, p.Vectors.ManifestPath)
				} else if name == "current-manifest" {
					path = filepath.Join(root, "manifest.sha256")
				}
				write(t, path, append(read(t, path), '\n'))
			case "required-file-missing":
				if err := os.Remove(filepath.Join(root, "totp/cases.json")); err != nil {
					t.Fatal(err)
				}
			default:
				path := filepath.Join(root, "totp/cases.json")
				var bundle corpus.Bundle
				if err := json.Unmarshal(read(t, path), &bundle); err != nil {
					t.Fatal(err)
				}
				if name == "required-case-missing" {
					bundle.Cases = bundle.Cases[1:]
				} else {
					bundle.Cases[0].Expected["code"] = "000000"
				}
				b, err := json.Marshal(bundle)
				if err != nil {
					t.Fatal(err)
				}
				write(t, path, b)
				if name != "changed-expectation-stale-manifest" {
					updateManifest(t, root)
				}
			}
			if _, err := p.LoadCases(repo, ""); err == nil {
				t.Fatal("accepted changed pinned artifacts")
			}
		})
	}
}

func TestUnknownReleaseID(t *testing.T) {
	p, repo := profileFixture(t)
	p.Requirements.CaseIDs[0] = "v0/bootstrap/aaa-unknown"
	sort.Strings(p.Requirements.CaseIDs)
	if _, err := p.LoadCases(repo, ""); err == nil || !strings.Contains(err.Error(), "missing required case") {
		t.Fatalf("unknown ID: %v", err)
	}
}

func TestExtraCorpusCaseDoesNotExpandProfile(t *testing.T) {
	p, repo := profileFixture(t)
	root := filepath.Join(repo, "vectors/v0")
	before, err := p.LoadCases(repo, "")
	if err != nil {
		t.Fatal(err)
	}
	var bundle corpus.Bundle
	if err := json.Unmarshal(read(t, filepath.Join(root, "totp/cases.json")), &bundle); err != nil {
		t.Fatal(err)
	}
	extra := bundle.Cases[0]
	extra.ID = "v0/totp/future-extra"
	b, err := json.Marshal(corpus.Bundle{Schema: 1, Cases: []corpus.Case{extra}})
	if err != nil {
		t.Fatal(err)
	}
	write(t, filepath.Join(root, "totp/future.json"), b)
	updateManifest(t, root)
	current, err := corpus.Load(root)
	if err != nil {
		t.Fatal(err)
	}
	selected, err := p.LoadCases(repo, "")
	if err != nil {
		t.Fatal(err)
	}
	if len(current) != len(before)+1 || !reflect.DeepEqual(before, selected) {
		t.Fatal("extra vector did not grow current corpus independently of frozen profile")
	}
	found := false
	for _, c := range current {
		found = found || c.ID == extra.ID
	}
	if !found {
		t.Fatal("current corpus did not include added case")
	}
}
