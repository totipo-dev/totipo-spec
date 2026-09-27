package vectors

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"reflect"
	"totipo/conformance/internal/cryptov1"
	"totipo/conformance/internal/graph"
	"totipo/conformance/internal/object"
)

type RetentionCase struct {
	Fixture string           `json:"fixture_case"`
	Notes   string           `json:"notes"`
	Trials  []RetentionTrial `json:"trials"`
}
type RetentionTrial struct {
	Persist           bool              `json:"persist"`
	Reclassify        string            `json:"reclassify,omitempty"`
	ReclassifyPersist bool              `json:"reclassify_persist,omitempty"`
	Want              RetentionExpected `json:"expect"`
}
type RetentionExpected struct {
	RetainedExact      bool   `json:"retained_exact"`
	PersistenceBlocked bool   `json:"persistence_blocked"`
	Authoritative      bool   `json:"authoritative"`
	Class              string `json:"class"`
}
type LateProvenanceCase struct {
	Fixture string            `json:"fixture_case"`
	Notes   string            `json:"notes"`
	Trials  []ProvenanceTrial `json:"trials"`
}
type ProvenanceTrial struct {
	CorruptSignature  bool   `json:"corrupt_signature"`
	Before            string `json:"before"`
	After             string `json:"after"`
	SemanticUnchanged bool   `json:"semantic_unchanged"`
}

func resolveRecovery(cases []Case) error {
	index := map[string]Case{}
	for _, c := range cases {
		index[c.ID] = c
	}
	for i := range cases {
		c := &cases[i]
		var id string
		if c.Retention != nil {
			id = c.Retention.Fixture
		}
		if c.LateProvenance != nil {
			id = c.LateProvenance.Fixture
		}
		if id == "" {
			continue
		}
		f, ok := index[id]
		if !ok {
			return fmt.Errorf("%s: missing recovery fixture %s", c.ID, id)
		}
		if c.Retention != nil && (f.Crypto == nil || f.Root == "" || f.Expected != object.Unscoped || (f.Operation != "dispatch" && f.Operation != "crypto")) {
			return fmt.Errorf("invalid opaque envelope fixture")
		}
		if c.LateProvenance != nil && (f.Operation != "crypto" || f.Input == nil || f.Root == "" || f.Crypto == nil || f.Crypto.PublicKey == "" || f.Expected != object.Supported) {
			return fmt.Errorf("invalid provenance fixture")
		}
		c.fixtures = map[string]Case{id: f}
	}
	return nil
}

func runRetention(c Case) error {
	x := c.Retention
	f, ok := c.fixtures[x.Fixture]
	if !ok {
		return fmt.Errorf("unresolved retention fixture")
	}
	root, err := unhex(f.Root)
	if err != nil {
		return err
	}
	keys, err := cryptov1.Derive(root)
	if err != nil {
		return err
	}
	fixture, err := unhex(f.Crypto.Object)
	if err != nil {
		return err
	}
	id := f.Crypto.ObjectID
	for i, trial := range x.Trials {
		s := graph.New()
		candidate := graph.Node{ID: "candidate", Type: "TOKEN", Identity: "T", Class: object.Supported, Digest: "candidate-bytes"}
		if err := s.Learn(candidate, &graph.Value{Status: "LIVE"}, true); err != nil {
			return err
		}
		syncCopy := bytes.Clone(fixture)
		if err := s.LearnOpaque(id, syncCopy, keys, trial.Persist); err != nil {
			return err
		}
		// Replace then remove the synchronized copy; this must not alias the record.
		clear(syncCopy)
		syncCopy = nil
		s.RemoteUnavailable(id, false)
		if trial.Reclassify != "" {
			called := false
			err := s.ReprocessOpaque(id, keys, func(semantic []byte) (graph.Node, *graph.Value, error) {
				called = true
				if err := equalHex("retained semantic", f.Semantic, semantic); err != nil {
					return graph.Node{}, nil, err
				}
				// Controlled future interpretation of this fixture only. This is
				// not a new published grammar or an OBJECT_VERSION allocation.
				return graph.Node{Version: 1, Type: "DEVICE", Identity: "synthetic-future", Class: trial.Reclassify}, nil, nil
			}, trial.ReclassifyPersist)
			if err != nil {
				return err
			}
			if !called {
				return fmt.Errorf("retained classifier not called")
			}
		}
		record, retained := s.OpaqueRecords[id]
		got := RetentionExpected{RetainedExact: retained && bytes.Equal(record.ExactObjectBytes[:], fixture), PersistenceBlocked: s.PersistenceBlocked, Authoritative: s.Authoritative(), Class: s.Nodes[id].Class}
		if got != trial.Want {
			return fmt.Errorf("retention trial %d: got %+v want %+v", i, got, trial.Want)
		}
		if s.Evaluate("T", "candidate", "").Candidate != !s.PersistenceBlocked || (s.PersistenceBlocked && s.BaseSafe()) {
			return fmt.Errorf("persistence safety gate open")
		}
	}
	return nil
}

func runLateProvenance(c Case) error {
	x := c.LateProvenance
	f, ok := c.fixtures[x.Fixture]
	if !ok {
		return fmt.Errorf("unresolved provenance fixture")
	}
	root, err := unhex(f.Root)
	if err != nil {
		return err
	}
	keys, err := cryptov1.Derive(root)
	if err != nil {
		return err
	}
	public, err := unhex(f.Crypto.PublicKey)
	if err != nil {
		return err
	}
	for i, trial := range x.Trials {
		o := *f.Input
		o.Signature = bytes.Clone(o.Signature)
		if trial.CorruptSignature {
			o.Signature[len(o.Signature)-1] ^= 1
		}
		semantic, err := o.Encode()
		if err != nil {
			return err
		}
		s := graph.New()
		n := graph.Node{ID: "token", Version: o.Version, Type: "TOKEN", Identity: hex.EncodeToString(o.Identity), Author: hex.EncodeToString(o.Author), Class: object.Supported, AuthorTime: o.AuthorTime, Digest: Hash(semantic)}
		for _, parent := range o.Parents {
			n.Parents = append(n.Parents, hex.EncodeToString(parent))
		}
		if len(n.Parents) != 1 {
			return fmt.Errorf("expected child fixture")
		}
		ancestor := n
		ancestor.ID = n.Parents[0]
		ancestor.Parents = nil
		v := graph.Value{Status: "LIVE", Issuer: o.Issuer, Account: o.Account, Algorithm: o.Algorithm, Digits: o.Digits, Period: o.Period, Secret: hex.EncodeToString(o.Secret)}
		if err := s.Learn(ancestor, &v, true); err != nil {
			return err
		}
		if err := s.Learn(n, &v, true); err != nil {
			return err
		}
		p := graph.NewProvenance(s, keys)
		if err := p.Track(n.ID, semantic); err != nil {
			return err
		}
		if s.Available[n.ID].Provenance != trial.Before {
			return fmt.Errorf("trial %d initial provenance mismatch", i)
		}
		before := s.Evaluate(n.Identity, n.ID, "")
		beforeNodes := map[string]graph.Node{}
		for id, node := range s.Nodes {
			beforeNodes[id] = node
		}
		beforeValue := s.Available[n.ID]
		p.KeyAvailable(public)
		afterValue := s.Available[n.ID]
		if afterValue.Provenance != trial.After {
			return fmt.Errorf("trial %d late provenance: %s", i, afterValue.Provenance)
		}
		beforeValue.Provenance, afterValue.Provenance = "", ""
		beforeValue.Verified, afterValue.Verified = false, false
		unchanged := reflect.DeepEqual(beforeNodes, s.Nodes) && beforeValue == afterValue && reflect.DeepEqual(before, s.Evaluate(n.Identity, n.ID, ""))
		if !trial.SemanticUnchanged || !unchanged {
			return fmt.Errorf("provenance changed TOKEN value/validity/ancestry/heads/authority")
		}
	}
	return nil
}
