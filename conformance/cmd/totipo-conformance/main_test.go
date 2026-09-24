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

const profilePath = "../../../requirements/v0-rc1.json"
const vectorRoot = "../../../vectors/v0"

func TestCLI(t *testing.T) {
	cases, err := corpus.Load(vectorRoot)
	if err != nil {
		t.Fatal(err)
	}
	totpCount := 0
	for _, c := range cases {
		if strings.HasPrefix(c.ID, "v0/totp/") {
			totpCount++
		}
	}
	for _, test := range []struct {
		name string
		args []string
		code int
		want string
	}{
		{"current-corpus", []string{vectorRoot}, 0, fmt.Sprintf("PASS %d / FAIL 0 / BLOCKED 0", len(cases))},
		{"full-profile", []string{"--requirements", profilePath}, 0, "PASS 388 / FAIL 0 / BLOCKED 0\nCONFORMANT v0-rc1"},
		{"partial-profile", []string{"--requirements", profilePath, "--filter", "v0/totp/"}, 0, "PASS 82 / FAIL 0 / BLOCKED 0"},
		{"empty-filter-still-partial", []string{"--requirements", profilePath, "--filter="}, 0, "PARTIAL v0-rc1"},
		{"verify-only", []string{"--requirements", profilePath, "--verify-only"}, 0, "cases not executed"},
		{"positional-and-equals", []string{vectorRoot, "--requirements=" + profilePath, "--verify-only"}, 0, "VALID v0-rc1"},
		{"legacy-filter-after-root", []string{vectorRoot, "--filter=v0/totp/"}, 0, fmt.Sprintf("PASS %d / FAIL 0 / BLOCKED 0", totpCount)},
		{"legacy-filter-before-root", []string{"--filter", "v0/totp/", vectorRoot}, 0, fmt.Sprintf("PASS %d / FAIL 0 / BLOCKED 0", totpCount)},
		{"no-matches", []string{"--requirements", profilePath, "--filter", "v0/absent/"}, 1, "filter matched no cases"},
		{"missing-profile", []string{"--requirements", "missing.json"}, 1, "missing.json"},
		{"verify-filter", []string{"--requirements", profilePath, "--filter=", "--verify-only"}, 1, "usage:"},
		{"verify-without-profile", []string{vectorRoot, "--verify-only"}, 1, "usage:"},
		{"empty-requirements", []string{vectorRoot, "--requirements="}, 1, "requires a value"},
		{"no-args", nil, 1, "usage:"},
		{"missing-filter-value", []string{vectorRoot, "--filter"}, 1, "requires a value"},
		{"missing-profile-value", []string{"--requirements"}, 1, "requires a value"},
		{"unknown-option", []string{vectorRoot, "--unknown"}, 1, "usage:"},
	} {
		t.Run(test.name, func(t *testing.T) {
			var out, errOut bytes.Buffer
			code := run(test.args, &out, &errOut)
			combined := out.String() + errOut.String()
			if code != test.code || !strings.Contains(combined, test.want) {
				t.Fatalf("exit=%d; want=%d, %q; output:\n%s", code, test.code, test.want, combined)
			}
			partial := strings.Contains(strings.Join(test.args, " "), "--filter") && strings.Contains(strings.Join(test.args, " "), "--requirements")
			if partial && code == 0 && (!strings.Contains(combined, "PARTIAL") || strings.Contains(combined, "CONFORMANT")) {
				t.Fatal("filtered run claimed complete conformance")
			}
			if strings.Contains(strings.Join(test.args, " "), "--verify-only") && strings.Contains(combined, "CONFORMANT") {
				t.Fatal("verification without execution claimed conformance")
			}
		})
	}
}

func TestProfileFailureAndBlocked(t *testing.T) {
	for _, blocked := range []bool{false, true} {
		t.Run(fmt.Sprintf("blocked=%v", blocked), func(t *testing.T) {
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
			c := corpus.Case{ID: "v0/envelope/test", Kind: "file_length", Description: "temporary CLI failure fixture",
				Provenance: corpus.Provenance{Source: "test", Status: "spec-derived", Locator: "test"},
				Input:      map[string]string{"length": "0"}, Expected: map[string]string{"disposition": "FILE_LENGTH_VALID"}}
			wantCode, wantResult := 1, "PASS 0 / FAIL 1 / BLOCKED 0"
			if blocked {
				c.Kind = "blocked"
				c.Input = map[string]string{"requirement": "test"}
				c.Expected = map[string]string{"disposition": "BLOCKED", "reason": "temporary blocked fixture"}
				wantCode, wantResult = 2, "PASS 0 / FAIL 0 / BLOCKED 1"
			}
			b, err := json.Marshal(corpus.Bundle{Schema: 1, Cases: []corpus.Case{c}})
			if err != nil {
				t.Fatal(err)
			}
			write("vectors/v0/cases.json", b)
			write("vectors/v0/manifest.sha256", []byte(fmt.Sprintf("%x  cases.json\n", sha256.Sum256(b))))
			write("spec/totipo-vault-format-v0.md", []byte("**Status:** test, revision 37\n"))
			p, manifest, err := requirements.Candidate(repo, "v0-test")
			if err != nil {
				t.Fatal(err)
			}
			if err := requirements.WriteCandidate(repo, p, manifest); err != nil {
				t.Fatal(err)
			}
			args := []string{"--requirements", filepath.Join(repo, "requirements/v0-test.json")}
			var out, errOut bytes.Buffer
			if code := run(args, &out, &errOut); code != wantCode || !strings.Contains(out.String(), wantResult) || strings.Contains(out.String(), "CONFORMANT") {
				t.Fatalf("exit=%d, stdout=%s, stderr=%s", code, &out, &errOut)
			}
			// A filtered diagnostic still validates every pinned artifact first.
			write("spec/totipo-vault-format-v0.md", []byte("mutated spec"))
			out.Reset()
			errOut.Reset()
			if code := run(append(args, "--filter", "v0/envelope/"), &out, &errOut); code != 1 || !strings.Contains(errOut.String(), "SHA-256 mismatch") {
				t.Fatalf("integrity failure exit=%d: %s", code, &errOut)
			}
		})
	}
}
