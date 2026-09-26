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
	for _, c := range cases {
		t.Run(c.ID, func(t *testing.T) {
			if e := Run(c); e != nil {
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
