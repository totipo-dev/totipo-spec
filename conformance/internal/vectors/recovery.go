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
		if c.LateProvenance != nil && (f.Operation != "crypto" || f.Input == nil || f.Root == "" || f.Crypto == nil || f.Crypto.PublicKey == "" || f.Expected != object.Supported) {
			return fmt.Errorf("invalid provenance fixture")
		}
		c.fixtures = map[string]Case{id: f}
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
		if err := s.Learn(ancestor, &v); err != nil {
			return err
		}
		if err := s.Learn(n, &v); err != nil {
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
