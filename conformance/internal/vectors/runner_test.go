package vectors

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCorpus(t *testing.T) {
	root := filepath.Join("..", "..", "..")
	m, cases, e := Read(root)
	if e != nil {
		t.Fatal(e)
	}
	if e = VerifyProfile(root, m); e != nil {
		t.Fatal(e)
	}
	selected, e := Select(m, cases, ReferenceCapabilities)
	if e != nil {
		t.Fatal(e)
	}
	for _, c := range selected {
		t.Run(c.ID, func(t *testing.T) {
			if e := RunWithCapabilities(c, ReferenceCapabilities); e != nil {
				t.Fatal(e)
			}
		})
	}
}
func TestManifestTamper(t *testing.T) {
	root := filepath.Join("..", "..", "..")
	m, _, e := Read(root)
	if e != nil {
		t.Fatal(e)
	}
	tmp := t.TempDir()
	if e = os.MkdirAll(filepath.Join(tmp, "vectors"), 0755); e != nil {
		t.Fatal(e)
	}
	b, e := os.ReadFile(filepath.Join(root, "vectors/manifest.json"))
	if e != nil {
		t.Fatal(e)
	}
	if e = os.WriteFile(filepath.Join(tmp, "vectors/manifest.json"), b, 0644); e != nil {
		t.Fatal(e)
	}
	if _, _, e = Read(tmp); e == nil {
		t.Fatal("missing case accepted")
	}
	m.Cases[0].SHA256 = "00"
	if e = VerifyProfile(root, m); e == nil {
		t.Fatal("profile mismatch accepted")
	}
}

func TestStorageReferencesAndExpectations(t *testing.T) {
	root := filepath.Join("..", "..", "..")
	_, cases, e := Read(root)
	if e != nil {
		t.Fatal(e)
	}
	for i, c := range cases {
		if c.ID != "v1.future.family-compat-shadow.001" {
			continue
		}
		// A missing fixture must fail verification even before case execution.
		altered := append([]Case(nil), cases...)
		x := *c.Storage
		x.Entries = append([]StorageEntry(nil), x.Entries...)
		x.Entries[0].Fixture = "v1.crypto.missing.001"
		altered[i].Storage = &x
		if e := resolveStorage(altered); e == nil {
			t.Fatal("missing fixture accepted")
		}
		// Ensure graph and observation expectations are actively checked.
		x = *c.Storage
		x.Want = c.Storage.Want
		x.Want.View.Ordinary = true
		c.Storage = &x
		if e := Run(c); e == nil {
			t.Fatal("incorrect opaque-state expectation accepted")
		}
		return
	}
	t.Fatal("missing compatibility case")
}

func TestRecoveryReferencesAndExpectations(t *testing.T) {
	_, cases, err := Read(filepath.Join("..", "..", ".."))
	if err != nil {
		t.Fatal(err)
	}
	checks := 0
	for _, c := range cases {
		if c.LateProvenance != nil {
			x := *c.LateProvenance
			x.Trials = append([]ProvenanceTrial(nil), x.Trials...)
			x.Trials[0].After = "REJECTED"
			c.LateProvenance = &x
			if Run(c) == nil {
				t.Fatal("incorrect late provenance expectation accepted")
			}
			checks++
		}
	}
	if checks != 1 {
		t.Fatalf("missing recovery coverage: %d", checks)
	}
}

func TestBaselineWithoutAdvisoryHistory(t *testing.T) {
	m, cases, err := Read(filepath.Join("..", "..", ".."))
	if err != nil {
		t.Fatal(err)
	}
	selected, err := Select(m, cases, nil)
	if err != nil {
		t.Fatal(err)
	}
	baseline := 0
	for _, e := range m.Cases {
		if e.Applicability.Kind == "baseline" {
			baseline++
		}
	}
	if len(selected) != baseline || len(selected) >= len(cases) {
		t.Fatal("baseline did not exclude optional cases")
	}
	for _, c := range selected {
		if err := Run(c); err != nil {
			t.Fatalf("baseline without cache %s: %v", c.ID, err)
		}
	}
	full, err := Select(m, cases, ReferenceCapabilities)
	if err != nil || len(full) != len(cases) {
		t.Fatal("reference suite omitted cases", err)
	}
	for _, c := range cases {
		if c.ID == "v1.history.memory-lost.001" && Run(c) == nil {
			t.Fatal("conditional operation silently ran without capability")
		}
	}
	for _, caps := range [][]string{{"unknown"}, {AdvisoryHistory, AdvisoryHistory}} {
		if _, err := Select(m, cases, caps); err == nil {
			t.Fatal("invalid capabilities accepted")
		}
	}
}

func TestProfileRequiresExactlyBaseline(t *testing.T) {
	root := filepath.Join("..", "..", "..")
	m, _, err := Read(root)
	if err != nil {
		t.Fatal(err)
	}
	for _, kind := range []string{"baseline", "conditional"} {
		altered := m
		altered.Cases = append([]Entry(nil), m.Cases...)
		for i, e := range altered.Cases {
			if e.Applicability.Kind != kind {
				continue
			}
			if kind == "baseline" {
				c := AdvisoryHistory
				altered.Cases[i].Applicability = Applicability{Kind: "conditional", Capability: &c}
			} else {
				altered.Cases[i].Applicability = Applicability{Kind: "baseline"}
			}
			break
		}
		if VerifyProfile(root, altered) == nil {
			t.Fatal("profile accepted changed applicability", kind)
		}
	}
}

func TestUnlistedPhysicalCaseRejected(t *testing.T) {
	root := t.TempDir()
	// Copy current contract into an isolated directory, including all case files.
	src := filepath.Join("..", "..", "..", "vectors")
	err := filepath.WalkDir(src, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		dst := filepath.Join(root, "vectors", rel)
		if d.IsDir() {
			return os.MkdirAll(dst, 0755)
		}
		b, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		return os.WriteFile(dst, b, 0644)
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "vectors/cases/unlisted.json"), []byte("{}"), 0644); err != nil {
		t.Fatal(err)
	}
	if _, _, err := Read(root); err == nil {
		t.Fatal("unlisted physical case accepted")
	}
}
