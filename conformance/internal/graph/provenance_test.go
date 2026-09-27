package graph

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"reflect"
	"testing"
	"totipo/conformance/internal/cryptov1"
	"totipo/conformance/internal/object"
)

func TestLateKeyForRetainedUnavailableToken(t *testing.T) {
	b, err := os.ReadFile("../../../vectors/cases/crypto/v1.crypto.token-child.001.json")
	if err != nil {
		t.Fatal(err)
	}
	var f struct {
		Root   string        `json:"root_hex"`
		Input  object.Object `json:"input"`
		Crypto struct {
			Public string `json:"fixture_public_key_hex"`
		} `json:"crypto"`
	}
	if err := json.Unmarshal(b, &f); err != nil {
		t.Fatal(err)
	}
	decode := func(h string) []byte {
		b, err := hex.DecodeString(h)
		if err != nil {
			t.Fatal(err)
		}
		return b
	}
	keys, err := cryptov1.Derive(decode(f.Root))
	if err != nil {
		t.Fatal(err)
	}
	semantic, err := f.Input.Encode()
	if err != nil {
		t.Fatal(err)
	}
	digest := sha256.Sum256(semantic)
	n := Node{ID: "token", Type: "TOKEN", Class: object.Supported, Version: f.Input.Version, Identity: hex.EncodeToString(f.Input.Identity), Author: hex.EncodeToString(f.Input.Author), AuthorTime: f.Input.AuthorTime, Digest: hex.EncodeToString(digest[:])}
	for _, parent := range f.Input.Parents {
		n.Parents = append(n.Parents, hex.EncodeToString(parent))
	}
	s := New()
	if err := s.Learn(n, &Value{Status: "LIVE"}, true); err != nil {
		t.Fatal(err)
	}
	p := NewProvenance(s, keys)
	if err := p.Track(n.ID, semantic); err != nil {
		t.Fatal(err)
	}
	public := decode(f.Crypto.Public)
	if p.Status[n.ID] != "UNRESOLVED" {
		t.Fatal("unexpected initial status")
	}
	p.KeyAvailable([]byte("unrelated key material"))
	if p.Status[n.ID] != "UNRESOLVED" {
		t.Fatal("unrelated key resolved attribution")
	}
	clear(semantic) // Tracker must own immutable evidence.
	s.Disappear(n.ID)
	before := s.Evaluate(n.Identity, n.ID, "")
	p.KeyAvailable(public)
	if p.Status[n.ID] != "VERIFIED" || len(s.Available) != 0 || !reflect.DeepEqual(before, s.Evaluate(n.Identity, n.ID, "")) {
		t.Fatal("late provenance changed unavailable TOKEN authority")
	}
	if err := p.Track(n.ID, semantic); err == nil {
		t.Fatal("invalid replacement evidence accepted")
	}
	if p.Status[n.ID] != "VERIFIED" {
		t.Fatal("failed replacement changed provenance")
	}
}
