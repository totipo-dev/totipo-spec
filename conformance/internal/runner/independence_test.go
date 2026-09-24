package runner

import (
	"encoding/json"
	"os"
	"strings"
	"testing"
	"totipo/conformance/internal/corpus"
)

func TestExternalNormalization(t *testing.T) {
	// Directly re-read the historical corpus to detect converter mistakes without rewriting it.
	b, e := os.ReadFile("../../../review/source-vectors/totp-vault-v0-r31-tlv-corpus.json")
	if e != nil {
		t.Fatal(e)
	}
	var source struct {
		Cases []struct {
			ID, Hex, Expect, Kind, Form string
			SemanticParseExpect         string `json:"semantic_parse_expect"`
		}
	}
	if e := json.Unmarshal(b, &source); e != nil {
		t.Fatal(e)
	}
	cases, e := corpus.Load("../../../vectors/v0")
	if e != nil {
		t.Fatal(e)
	}
	normalized := map[string]corpus.Case{}
	for _, c := range cases {
		if c.Provenance.Source == "review/source-vectors/totp-vault-v0-r31-tlv-corpus.json" {
			normalized[c.Provenance.Locator] = c
		}
	}
	for _, old := range source.Cases {
		c, ok := normalized[old.ID]
		if !ok {
			t.Fatalf("missing %s", old.ID)
		}
		if old.Kind != "object_file_length_gate" && c.Input["bytes_hex"] != old.Hex {
			t.Fatalf("altered bytes %s", old.ID)
		}
		valid := c.Expected["disposition"] == "STRUCTURALLY_VALID" || c.Expected["disposition"] == "ENVELOPE_VALID" || c.Expected["disposition"] == "FILE_LENGTH_VALID"
		if valid != (old.Expect == "valid") {
			t.Fatalf("altered disposition %s", old.ID)
		}
	}
}

func TestPinnedCryptoNormalization(t *testing.T) {
	cases, e := corpus.Load("../../../vectors/v0")
	if e != nil {
		t.Fatal(e)
	}
	for _, c := range cases {
		if c.Kind != "object_crypto" && c.Kind != "bootstrap" {
			continue
		}
		raw, e := os.ReadFile("../../../" + c.Provenance.Source)
		if e != nil {
			t.Fatal(e)
		}
		section := ""
		for _, line := range strings.Split(string(raw), "\n") {
			if strings.HasPrefix(line, "[") {
				section = strings.Trim(line, "[]")
				continue
			}
			key, value, ok := strings.Cut(line, "=")
			if !ok {
				continue
			}
			if c.Kind == "bootstrap" && section == "" && key == "root" {
				if c.Input["root_hex"] != value {
					t.Fatal("bootstrap root changed")
				}
			}
			if section != c.Provenance.Locator {
				continue
			}
			got, ok := c.Input[key+"_hex"]
			if !ok {
				got, ok = c.Expected[key+"_hex"]
			}
			if !ok || got != value {
				t.Fatalf("%s: reviewed %s altered or missing", c.ID, key)
			}
		}
	}
}
