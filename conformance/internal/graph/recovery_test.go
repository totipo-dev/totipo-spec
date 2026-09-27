package graph

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"os"
	"testing"
	"totipo/conformance/internal/cryptov1"
	"totipo/conformance/internal/object"
)

func opaqueFixture(t *testing.T) (string, []byte, cryptov1.Keys) {
	t.Helper()
	b, err := os.ReadFile("../../../vectors/cases/routing/v1.routing.unknown-type-unscoped.001.json")
	if err != nil {
		t.Fatal(err)
	}
	var f struct {
		Root   string `json:"root_hex"`
		Crypto struct {
			ID    string `json:"object_id"`
			Bytes string `json:"object_hex"`
		} `json:"crypto"`
	}
	if err := json.Unmarshal(b, &f); err != nil {
		t.Fatal(err)
	}
	root, err := hex.DecodeString(f.Root)
	if err != nil {
		t.Fatal(err)
	}
	data, err := hex.DecodeString(f.Crypto.Bytes)
	if err != nil {
		t.Fatal(err)
	}
	keys, err := cryptov1.Derive(root)
	if err != nil {
		t.Fatal(err)
	}
	return f.Crypto.ID, data, keys
}

func TestRetainedEvidenceFailureBoundaries(t *testing.T) {
	id, encrypted, keys := opaqueFixture(t)
	s := New()
	bad := append([]byte(nil), encrypted...)
	bad[0] ^= 1
	if s.LearnOpaque(id, bad, keys, true) == nil || len(s.Nodes) != 0 || len(s.OpaqueRecords) != 0 || s.ContinuityUnknown {
		t.Fatal("unauthenticated remote bytes became local evidence")
	}
	if err := s.LearnOpaque(id, encrypted, keys, true); err != nil {
		t.Fatal(err)
	}
	old := s.Nodes[id]
	if s.Reclassify(Node{ID: id, Class: object.Opaque, Digest: old.Digest}, nil, true) == nil {
		t.Fatal("digest-only concrete reclassification accepted")
	}
	failure := errors.New("unsupported by compatible classifier")
	if err := s.ReprocessOpaque(id, keys, func([]byte) (Node, *Value, error) { return Node{}, nil, failure }, true); err != failure || s.Nodes[id].Class != object.Unscoped || s.Authoritative() {
		t.Fatal("failed classifier cleared evidence")
	}
	s.BeginReset()
	if s.FinishReset(false) || len(s.OpaqueRecords) != 1 {
		t.Fatal("incomplete reset discarded exact evidence")
	}
	if !s.FinishReset(true) || len(s.OpaqueRecords) != 0 {
		t.Fatal("complete absent reset kept old evidence")
	}

	s = New()
	if err := s.LearnOpaque(id, encrypted, keys, true); err != nil {
		t.Fatal(err)
	}
	record := s.OpaqueRecords[id]
	record.ExactObjectBytes[0] ^= 1
	s.OpaqueRecords[id] = record
	called := false
	if err := s.ReprocessOpaque(id, keys, func([]byte) (Node, *Value, error) { called = true; return Node{}, nil, nil }, true); err != ErrIntegrity || !s.ContinuityUnknown || called {
		t.Fatal("corrupt local retained bytes reached classifier")
	}
}

func TestResetRejectsIncompleteReplacementDiscovery(t *testing.T) {
	s := New()
	s.BeginReset()
	s.replacement.DiscoveryIncomplete = true
	if s.FinishReset(true) || !s.ContinuityUnknown {
		t.Fatal("caller completion flag bypassed incomplete discovery")
	}
}

func TestBaselineRequiresExactOpaqueRetention(t *testing.T) {
	id, encrypted, keys := opaqueFixture(t)
	s := New()
	s.BeginReset()
	if err := s.BaselineLearnOpaque(id, encrypted, keys, false); err != nil {
		t.Fatal(err)
	}
	if s.FinishReset(true) || !s.ContinuityUnknown {
		t.Fatal("baseline completed despite failed opaque retention")
	}
	s.BeginReset()
	if err := s.BaselineLearnOpaque(id, encrypted, keys, true); err != nil {
		t.Fatal(err)
	}
	if !s.FinishReset(true) || s.Authoritative() || len(s.OpaqueRecords) != 1 {
		t.Fatal("rediscovered opaque bytes not retained/blocking")
	}
}
