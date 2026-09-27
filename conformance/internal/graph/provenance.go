package graph

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"reflect"
	"totipo/conformance/internal/cryptov1"
	"totipo/conformance/internal/object"
)

// Provenance tracks known TOKEN semantic evidence separately from graph/value
// authority. It models retained readable assertions, not production persistence.
type Provenance struct {
	Status     map[string]string
	state      *State
	keys       cryptov1.Keys
	tokens     map[string][]byte
	publicKeys map[string][]byte
}

func NewProvenance(s *State, keys cryptov1.Keys) *Provenance {
	return &Provenance{state: s, keys: cryptov1.Keys{ID: bytes.Clone(keys.ID), ObjectRoot: bytes.Clone(keys.ObjectRoot), SignatureContext: bytes.Clone(keys.SignatureContext)}, Status: map[string]string{}, tokens: map[string][]byte{}, publicKeys: map[string][]byte{}}
}

// Track attaches already authenticated semantic bytes to an existing TOKEN.
func (p *Provenance) Track(id string, semantic []byte) error {
	class, o := object.Dispatch(semantic)
	n, ok := p.state.Nodes[id]
	if !ok || n.Type != "TOKEN" || n.Class != object.Supported || class != object.Supported || o.Type != object.Token || n.Author != hex.EncodeToString(o.Author) {
		return errors.New("invalid known TOKEN provenance evidence")
	}
	digest := sha256.Sum256(semantic)
	var parents []string
	for _, parent := range o.Parents {
		parents = append(parents, hex.EncodeToString(parent))
	}
	if n.Digest != hex.EncodeToString(digest[:]) || n.Identity != hex.EncodeToString(o.Identity) || n.Version != o.Version || n.AuthorTime != o.AuthorTime || !reflect.DeepEqual(clone(n).Parents, clone(Node{Parents: parents}).Parents) {
		return errors.New("TOKEN provenance evidence differs from known assertion")
	}
	p.tokens[id] = bytes.Clone(semantic)
	p.recompute(id, o)
	return nil
}

// KeyAvailable covers DEVICE arrival, restored records, and locally bound keys.
// DEVICE_ID is derived from the supplied canonical key, never caller-selected.
func (p *Provenance) KeyAvailable(public []byte) {
	author := hex.EncodeToString(object.DeviceID(public))
	p.publicKeys[author] = bytes.Clone(public)
	for id, semantic := range p.tokens {
		_, o := object.Dispatch(semantic)
		if hex.EncodeToString(o.Author) == author {
			p.recompute(id, o)
		}
	}
}

func (p *Provenance) recompute(id string, o *object.Object) {
	p.Status[id] = p.keys.Provenance(*o, p.publicKeys[hex.EncodeToString(o.Author)])
	if v, ok := p.state.Available[id]; ok {
		v.Provenance = p.Status[id]
		v.Verified = v.Provenance == "VERIFIED"
		p.state.Available[id] = v
	}
}
