package corpus

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func artifact(t *testing.T) []byte {
	t.Helper()
	b, e := os.ReadFile("../../../vectors/v0/tlv/cases.json")
	if e != nil {
		t.Fatal(e)
	}
	return b
}
func fixture(t *testing.T, b []byte) string {
	t.Helper()
	root := t.TempDir()
	if e := os.WriteFile(filepath.Join(root, "cases.json"), b, 0600); e != nil {
		t.Fatal(e)
	}
	h := sha256.Sum256(b)
	if e := os.WriteFile(filepath.Join(root, "manifest.sha256"), []byte(fmt.Sprintf("%x  cases.json\n", h)), 0600); e != nil {
		t.Fatal(e)
	}
	return root
}
func TestStrictHex(t *testing.T) {
	for _, s := range []string{"0", "AA", "0x01", "00 11", "00\n", "gg"} {
		if _, e := Hex(s); e == nil {
			t.Errorf("accepted %q", s)
		}
	}
	for _, s := range []string{"", "0011abcdef"} {
		if _, e := Hex(s); e != nil {
			t.Fatal(e)
		}
	}
}
func TestSchema(t *testing.T) {
	b := artifact(t)
	for name, mutate := range map[string]func(string) string{
		"unknown-key":   func(s string) string { return strings.Replace(s, `"schema": 1`, `"schema": 1, "extra": true`, 1) },
		"duplicate-key": func(s string) string { return strings.Replace(s, `"schema": 1`, `"schema": 1, "schema": 1`, 1) },
		"version":       func(s string) string { return strings.Replace(s, `"schema": 1`, `"schema": 2`, 1) },
		"hex":           func(s string) string { return strings.Replace(s, `"bytes_hex": "`, `"bytes_hex": "FF`, 1) },
		"null":          func(s string) string { return strings.Replace(s, `"form": "signed"`, `"form": null`, 1) },
		"unknown-kind":  func(s string) string { return strings.Replace(s, `"kind": "tlv"`, `"kind": "future"`, 1) },
		"trailing":      func(s string) string { return s + `{}` },
		"truncated":     func(s string) string { return s[:len(s)/2] },
	} {
		t.Run(name, func(t *testing.T) {
			if _, e := Decode([]byte(mutate(string(b)))); e == nil {
				t.Fatal("accepted malformed artifact")
			}
		})
	}
}
func TestDuplicateIDs(t *testing.T) {
	var b Bundle
	if e := json.Unmarshal(artifact(t), &b); e != nil {
		t.Fatal(e)
	}
	b.Cases = append(b.Cases, b.Cases[0])
	raw, e := json.Marshal(b)
	if e != nil {
		t.Fatal(e)
	}
	if _, e := Load(fixture(t, raw)); e == nil || !strings.Contains(e.Error(), "duplicate ID") {
		t.Fatalf("got %v", e)
	}
}
func TestManifest(t *testing.T) {
	b := artifact(t)
	root := fixture(t, b)
	if _, e := Load(root); e != nil {
		t.Fatal(e)
	}
	if e := os.WriteFile(filepath.Join(root, "cases.json"), append(b, ' '), 0600); e != nil {
		t.Fatal(e)
	}
	if _, e := Load(root); e == nil {
		t.Fatal("stale manifest accepted")
	}
	root = fixture(t, b)
	if e := os.WriteFile(filepath.Join(root, "unlisted.json"), b, 0600); e != nil {
		t.Fatal(e)
	}
	if _, e := Load(root); e == nil {
		t.Fatal("unlisted artifact accepted")
	}
}
func TestUnreadable(t *testing.T) {
	root := fixture(t, artifact(t))
	p := filepath.Join(root, "cases.json")
	if e := os.Remove(p); e != nil {
		t.Fatal(e)
	}
	if _, e := Load(root); e == nil {
		t.Fatal("missing vector accepted")
	}
	if _, e := Load(filepath.Join(root, "missing")); e == nil {
		t.Fatal("missing corpus accepted")
	}
}
