package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"totipo/conformance/internal/corpus"
	"totipo/conformance/internal/requirements"
)

func TestExplicitWriteAndNoOverwrite(t *testing.T) {
	repo := t.TempDir()
	write := func(path string, b []byte) {
		t.Helper()
		path = filepath.Join(repo, path)
		if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, b, 0600); err != nil {
			t.Fatal(err)
		}
	}
	cases, err := corpus.Load("../../../vectors/v0")
	if err != nil {
		t.Fatal(err)
	}
	b, err := json.Marshal(corpus.Bundle{Schema: 1, Cases: cases[:1]})
	if err != nil {
		t.Fatal(err)
	}
	write("vectors/v0/cases.json", b)
	manifest := []byte(fmt.Sprintf("%x  cases.json\n", sha256.Sum256(b)))
	write("vectors/v0/manifest.sha256", manifest)
	write("spec/totipo-vault-format-v0.md", []byte("# Example spec\n\n**Status:** v0 release candidate, revision 37  \n"))
	args := []string{"--profile", "v0-test", "--repo-root", repo}
	var out, errOut bytes.Buffer
	if code := run(args, &out, &errOut); code != 0 {
		t.Fatalf("preview exit %d: %s", code, errOut.String())
	}
	if _, err := os.Stat(filepath.Join(repo, "requirements")); !os.IsNotExist(err) {
		t.Fatal("preview wrote artifacts")
	}
	var p requirements.Profile
	if err := json.Unmarshal(out.Bytes(), &p); err != nil {
		t.Fatal(err)
	}
	if p.Spec.Revision != 37 || p.Requirements.Expected.Pass != 1 || p.Requirements.CaseIDs[0] != cases[0].ID {
		t.Fatal("candidate was not derived from the actual corpus/spec")
	}
	out.Reset()
	if code := run(append(args, "--write"), &out, &errOut); code != 0 {
		t.Fatalf("write exit %d: %s", code, errOut.String())
	}
	profilePath := filepath.Join(repo, "requirements/v0-test.json")
	pinned, err := requirements.Load(profilePath)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := pinned.LoadCases(repo, ""); err != nil {
		t.Fatal(err)
	}
	before, err := requirements.HashFile(profilePath)
	if err != nil {
		t.Fatal(err)
	}
	if code := run(append(args, "--write"), &out, &errOut); code == 0 || !strings.Contains(errOut.String(), "refusing to overwrite") {
		t.Fatal("write silently replaced a profile")
	}
	after, err := requirements.HashFile(profilePath)
	if err != nil || before != after {
		t.Fatal("existing profile changed")
	}
	frozen, err := os.ReadFile(filepath.Join(repo, pinned.Vectors.ManifestPath))
	if err != nil || !bytes.Equal(frozen, manifest) {
		t.Fatal("frozen manifest differs from original")
	}
}
