package graph

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"totipo/conformance/internal/cryptov1"
	"totipo/conformance/internal/object"
)

// LearnOpaque authenticates current external bytes; no durable retention is required.
func (s *State) LearnOpaque(id string, encrypted []byte, keys cryptov1.Keys) error {
	semantic, err := keys.Open(id, encrypted)
	if err != nil {
		return err
	}
	class, _ := object.Dispatch(semantic)
	if class != object.Unscoped {
		return errors.New("expected unscoped evidence")
	}
	digest := sha256.Sum256(semantic)
	return s.Learn(Node{ID: id, Class: class, Digest: hex.EncodeToString(digest[:])}, nil)
}
