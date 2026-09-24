package requirements

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"totipo/conformance/internal/corpus"
)

func exampleProfile() *Profile {
	return &Profile{
		Profile: "v0-example", Protocol: "Totipo Vault Format v0",
		Spec:         Spec{Path: "spec/example.md", Revision: 37, SHA256: strings.Repeat("a", 64)},
		Vectors:      Vectors{Root: "vectors/v0", ManifestPath: "vectors/v0/manifest.sha256", ManifestSHA256: strings.Repeat("b", 64)},
		Requirements: Requirements{CaseIDs: []string{"v0/totp/example"}, Expected: Counts{Pass: 1}, CategoryCounts: map[string]int{"totp": 1}},
	}
}

func write(t *testing.T, path string, b []byte) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, b, 0600); err != nil {
		t.Fatal(err)
	}
}

func read(t *testing.T, path string) []byte {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return b
}

func TestProfileSchema(t *testing.T) {
	b, err := json.Marshal(exampleProfile())
	if err != nil {
		t.Fatal(err)
	}
	for name, mutate := range map[string]func(string) string{
		"valid":         func(s string) string { return s },
		"unknown-field": func(s string) string { return strings.Replace(s, `"profile":`, `"extra":true,"profile":`, 1) },
		"duplicate-key": func(s string) string { return strings.Replace(s, `"revision":37`, `"revision":36,"revision":37`, 1) },
		"null":          func(s string) string { return strings.Replace(s, `"blocked":0`, `"blocked":null`, 1) },
		"missing-zero":  func(s string) string { return strings.Replace(s, `,"blocked":0`, ``, 1) },
		"missing-spec":  func(s string) string { return strings.Replace(s, `"revision":37,`, ``, 1) },
		"trailing":      func(s string) string { return s + `{}` },
		"truncated":     func(s string) string { return s[:len(s)/2] },
	} {
		t.Run(name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "profile.json")
			write(t, path, []byte(mutate(string(b))))
			_, err := Load(path)
			if name == "valid" && err != nil || name != "valid" && err == nil {
				t.Fatalf("Load returned %v", err)
			}
		})
	}
}

func TestProfileRequirements(t *testing.T) {
	for name, mutate := range map[string]func(*Profile){
		"duplicate-required-ID": func(p *Profile) {
			p.Requirements.CaseIDs = append(p.Requirements.CaseIDs, p.Requirements.CaseIDs[0])
			p.Requirements.Expected.Pass = 2
		},
		"unsorted": func(p *Profile) {
			p.Requirements.CaseIDs = append(p.Requirements.CaseIDs, "v0/totp/aaa")
			p.Requirements.Expected.Pass = 2
		},
		"invalid-ID":       func(p *Profile) { p.Requirements.CaseIDs[0] = "invalid" },
		"empty":            func(p *Profile) { p.Requirements.CaseIDs = nil; p.Requirements.Expected.Pass = 0 },
		"total":            func(p *Profile) { p.Requirements.Expected.Pass++ },
		"expected-fail":    func(p *Profile) { p.Requirements.Expected.Fail = 1 },
		"expected-blocked": func(p *Profile) { p.Requirements.Expected.Blocked = 1 },
		"category-count":   func(p *Profile) { p.Requirements.CategoryCounts["totp"]++ },
		"missing-category": func(p *Profile) { delete(p.Requirements.CategoryCounts, "totp") },
		"extra-category":   func(p *Profile) { p.Requirements.CategoryCounts["unused"] = 1 },
		"hash":             func(p *Profile) { p.Spec.SHA256 = "unknown" },
		"absolute-path":    func(p *Profile) { p.Spec.Path = "/tmp/spec.md" },
		"path-traversal":   func(p *Profile) { p.Spec.Path = "../spec.md" },
		"windows-path":     func(p *Profile) { p.Spec.Path = `C:\spec.md` },
	} {
		t.Run(name, func(t *testing.T) {
			p := exampleProfile()
			mutate(p)
			if err := p.Validate(); err == nil {
				t.Fatal("accepted invalid profile")
			}
		})
	}
}

func TestSelect(t *testing.T) {
	p := exampleProfile()
	cases := []corpus.Case{{ID: "v0/totp/extra"}, {ID: "v0/totp/example"}}
	got, err := p.Select(cases)
	if err != nil || len(got) != 1 || got[0].ID != "v0/totp/example" {
		t.Fatalf("selection: %v, %v", got, err)
	}
	if _, err := p.Select(cases[:1]); err == nil || !strings.Contains(err.Error(), "missing required case") {
		t.Fatalf("missing required case: %v", err)
	}
	if _, err := p.Select(append(cases, cases[1])); err == nil || !strings.Contains(err.Error(), "duplicate corpus ID") {
		t.Fatalf("duplicate corpus ID: %v", err)
	}
	p.Requirements.CaseIDs[0] = "v0/totp/unknown"
	if _, err := p.Select(cases); err == nil || !strings.Contains(err.Error(), "missing required case") {
		t.Fatalf("unknown required ID: %v", err)
	}
}
