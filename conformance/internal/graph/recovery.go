package graph

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"totipo/conformance/internal/cryptov1"
	"totipo/conformance/internal/object"
)

// OpaqueUnscopedRecord retains encrypted evidence, not a parsed reconstruction.
// The map key is OBJECT_ID. Array values give callers copies rather than aliases.
type OpaqueUnscopedRecord struct {
	ExactObjectBytes [1024]byte
}

// LearnOpaque authenticates the envelope and keyed ID before recording evidence.
// Learn remains the symbolic topology API used by the pre-r13 graph fixtures.
func (s *State) LearnOpaque(id string, encrypted []byte, keys cryptov1.Keys, persist bool) error {
	semantic, err := keys.Open(id, encrypted)
	if err != nil {
		return err
	}
	class, _ := object.Dispatch(semantic)
	if class != object.Unscoped {
		return errors.New("expected authenticated opaque-unscoped object")
	}
	if !persist {
		s.PersistenceBlocked = true
		return nil
	}
	digest := sha256.Sum256(semantic)
	n := Node{ID: id, Class: class, Digest: hex.EncodeToString(digest[:])}
	if err := s.Learn(n, nil, true); err != nil {
		return err
	}
	s.OpaqueRecords[id] = OpaqueUnscopedRecord{ExactObjectBytes: [1024]byte(encrypted)}
	return nil
}

// BaselineLearnOpaque applies the same retention boundary to a replacement epoch.
func (s *State) BaselineLearnOpaque(id string, encrypted []byte, keys cryptov1.Keys, persist bool) error {
	if s.replacement == nil {
		return errors.New("reset not started")
	}
	return s.replacement.LearnOpaque(id, encrypted, keys, persist)
}

// ReprocessOpaque hands a private copy of the retained, reauthenticated semantic
// bytes to a compatible classifier. It does not consult synchronized storage.
// A classifier must implement its own supported/future grammar; no new grammar
// is allocated by this reference model.
func (s *State) ReprocessOpaque(id string, keys cryptov1.Keys, classify func([]byte) (Node, *Value, error), persist bool) error {
	record, ok := s.OpaqueRecords[id]
	if !ok {
		return errors.New("missing retained opaque object")
	}
	semantic, err := keys.Open(id, record.ExactObjectBytes[:])
	if err != nil {
		s.ContinuityUnknown = true
		return ErrIntegrity
	}
	n, v, err := classify(bytes.Clone(semantic))
	if err != nil {
		return err
	}
	digest := sha256.Sum256(semantic)
	n.ID, n.Digest = id, hex.EncodeToString(digest[:])
	return s.reclassify(n, v, persist)
}

// ScanCandidate is one observation in a fixed baseline scan snapshot. Missing
// classifications, resource exhaustion, and unavailable bytes are nonterminal.
type ScanCandidate struct {
	ID             string `json:"id"`
	Classification string `json:"classification"`
}

func ResourceComplete(snapshot []ScanCandidate) bool {
	seen := map[string]bool{}
	for _, c := range snapshot {
		if c.ID == "" || seen[c.ID] {
			return false
		}
		seen[c.ID] = true
		switch c.Classification {
		case object.Supported, object.Opaque, object.Unscoped, object.Invalid, "INVALID_STORAGE":
		default:
			return false
		}
	}
	return true
}
