// Package graph is a deterministic in-memory model of durable security records.
// It models successful persistence, loss of synchronized bytes, and safety gates;
// it is not a filesystem/database implementation.
package graph

import (
	"encoding/json"
	"errors"
	"reflect"
	"sort"
)

var ErrIntegrity = errors.New("local graph integrity failure")

type Node struct {
	ID         string   `json:"id"`
	Version    uint8    `json:"version"`
	Type       string   `json:"type"`
	Identity   string   `json:"identity"`
	Class      string   `json:"class"`
	Parents    []string `json:"parents"`
	Author     string   `json:"author,omitempty"`
	AuthorTime uint64   `json:"author_time"`
	PublicKey  string   `json:"public_key,omitempty"`
	// Digest represents the exact authenticated semantic bytes in abstract cases.
	Digest string `json:"semantic_digest"`
}
type Value struct {
	Status      string `json:"status"`
	Issuer      string `json:"issuer"`
	Account     string `json:"account"`
	Algorithm   uint8  `json:"algorithm"`
	Digits      uint8  `json:"digits"`
	Period      uint32 `json:"period"`
	Secret      string `json:"secret_hex"`
	DisplayName string `json:"display_name,omitempty"`
	Verified    bool   `json:"verified,omitempty"`
	Provenance  string `json:"provenance,omitempty"`
}
type State struct {
	Nodes               map[string]Node
	Available           map[string]Value
	ContinuityUnknown   bool
	PersistenceBlocked  bool
	DiscoveryIncomplete bool
	replacement         *State
}

func New() *State { return &State{Nodes: map[string]Node{}, Available: map[string]Value{}} }
func clone(n Node) Node {
	n.Parents = append([]string(nil), n.Parents...)
	sort.Strings(n.Parents)
	return n
}
func (s *State) Learn(n Node, v *Value, persist bool) error {
	if !persist {
		s.PersistenceBlocked = true
		return nil
	}
	n = clone(n)
	if old, ok := s.Nodes[n.ID]; ok && !reflect.DeepEqual(old, n) {
		s.ContinuityUnknown = true
		return ErrIntegrity
	}
	s.Nodes[n.ID] = n
	if v != nil && n.Class == "SUPPORTED_VALID" {
		s.Available[n.ID] = *v
	}
	if s.cyclic() {
		s.ContinuityUnknown = true
		return ErrIntegrity
	}
	return nil
}
func (s *State) Disappear(id string) { delete(s.Available, id) }

// RemoteUnavailable never changes durable security knowledge. A trusted exact
// local copy may continue supplying the value despite hostile synchronized bytes.
func (s *State) RemoteUnavailable(id string, trustedCopy bool) {
	if !trustedCopy {
		s.Disappear(id)
	}
}

func (s *State) CorruptSecurityMemory() { s.ContinuityUnknown = true }

// BeginReset represents explicit user confirmation of lost continuity guarantees.
// The old epoch remains intact until a complete, successfully persisted scan.
func (s *State) BeginReset() {
	s.ContinuityUnknown = true
	s.replacement = New()
}
func (s *State) BaselineLearn(n Node, v *Value, persist bool) error {
	if s.replacement == nil {
		return errors.New("reset not started")
	}
	return s.replacement.Learn(n, v, persist)
}
func (s *State) FinishReset(complete bool) bool {
	if s.replacement == nil || !complete || !s.replacement.BaseSafe() {
		return false
	}
	*s = *s.replacement
	return true
}

// Reclassify models compatible reprocessing of the exact authenticated object;
// Digest represents its immutable semantic bytes, not its old interpretation.
func (s *State) Reclassify(n Node, v *Value, persist bool) error {
	old, ok := s.Nodes[n.ID]
	if !ok || old.Class != "OPAQUE_UNSCOPED" || old.Digest == "" || old.Digest != n.Digest ||
		(n.Class != "SUPPORTED_VALID" && n.Class != "OPAQUE_ROUTABLE") {
		s.ContinuityUnknown = true
		return ErrIntegrity
	}
	if !persist {
		s.PersistenceBlocked = true
		return nil
	}
	delete(s.Nodes, n.ID)
	return s.Learn(n, v, true)
}

func (s *State) Rename(n Node, v Value) bool {
	if !s.Authoritative() || n.Type != "DEVICE" || n.Class != "SUPPORTED_VALID" || !verified(v) {
		return false
	}
	heads := s.Heads("DEVICE", n.Identity)
	for _, id := range heads {
		if s.Nodes[id].Class != "SUPPORTED_VALID" {
			return false
		}
	}
	n.Parents = heads
	return s.Learn(n, &v, true) == nil
}

func verified(v Value) bool { return v.Provenance == "VERIFIED" || (v.Provenance == "" && v.Verified) }
func (s *State) Edge(child, parent string) string {
	c, ok := s.Nodes[child]
	if !ok {
		return "UNRESOLVED"
	}
	p, ok := s.Nodes[parent]
	if !ok {
		return "UNRESOLVED"
	}
	if c.Class == "OPAQUE_UNSCOPED" || p.Class == "OPAQUE_UNSCOPED" || c.Type != p.Type || c.Identity != p.Identity {
		return "REJECTED"
	}
	return "RESOLVED"
}
func (s *State) cyclic() bool {
	colors := map[string]uint8{}
	var visit func(string) bool
	visit = func(id string) bool {
		if colors[id] == 1 {
			return true
		}
		if colors[id] == 2 {
			return false
		}
		colors[id] = 1
		for _, p := range s.Nodes[id].Parents {
			if s.Edge(id, p) == "RESOLVED" && visit(p) {
				return true
			}
		}
		colors[id] = 2
		return false
	}
	for id := range s.Nodes {
		if visit(id) {
			return true
		}
	}
	return false
}
func (s *State) Heads(typ, identity string) []string {
	heads := map[string]bool{}
	for id, n := range s.Nodes {
		if n.Type == typ && n.Identity == identity && n.Class != "OPAQUE_UNSCOPED" {
			heads[id] = true
		}
	}
	for _, n := range s.Nodes {
		if n.Type != typ || n.Identity != identity {
			continue
		}
		for _, p := range n.Parents {
			if s.Edge(n.ID, p) == "RESOLVED" {
				delete(heads, p)
			}
		}
	}
	out := []string{}
	for id := range heads {
		out = append(out, id)
	}
	sort.Strings(out)
	return out
}
func (s *State) BaseSafe() bool { return !s.ContinuityUnknown && !s.PersistenceBlocked }
func (s *State) Authoritative() bool {
	if !s.BaseSafe() || s.DiscoveryIncomplete {
		return false
	}
	for _, n := range s.Nodes {
		if n.Class == "OPAQUE_UNSCOPED" {
			return false
		}
	}
	return true
}

type Result struct {
	Heads            []string `json:"heads"`
	ValueState       string   `json:"value_state"`
	Ordinary         bool     `json:"ordinary"`
	Author           bool     `json:"author"`
	Candidate        bool     `json:"candidate"`
	CandidateWarning bool     `json:"candidate_warning"`
	IntegrityFailure bool     `json:"integrity_failure"`
	Presentation     string   `json:"presentation,omitempty"`
}

func (s *State) Evaluate(identity, candidate, device string) Result {
	r := Result{Heads: s.Heads("TOKEN", identity), ValueState: "EMPTY", IntegrityFailure: s.ContinuityUnknown}
	opaque, missing := false, false
	values := map[string]Value{}
	for _, id := range r.Heads {
		n := s.Nodes[id]
		if n.Class == "OPAQUE_ROUTABLE" {
			opaque = true
			continue
		}
		v, ok := s.Available[id]
		if !ok {
			missing = true
			continue
		}
		v.DisplayName = ""
		v.Verified = false
		v.Provenance = ""
		b, _ := json.Marshal(v)
		values[string(b)] = v
	}
	switch {
	case opaque:
		r.ValueState = "VALUE_INCOMPLETE_OPAQUE"
	case missing:
		r.ValueState = "VALUE_INCOMPLETE_UNAVAILABLE"
	case len(values) > 1:
		r.ValueState = "CONFLICT"
	case len(values) == 1:
		r.ValueState = "UNAMBIGUOUS"
	}
	r.Author = s.Authoritative() && !opaque
	if r.ValueState == "UNAMBIGUOUS" && s.Authoritative() {
		for _, v := range values {
			r.Ordinary = v.Status == "LIVE"
		}
	}
	n, known := s.Nodes[candidate]
	v, available := s.Available[candidate]
	r.Candidate = s.BaseSafe() && known && available && n.Type == "TOKEN" && n.Identity == identity && n.Class == "SUPPORTED_VALID" && v.Status == "LIVE"
	r.CandidateWarning = r.Candidate
	if device != "" {
		names := map[string]bool{}
		h := s.Heads("DEVICE", device)
		r.Presentation = "UNAVAILABLE"
		complete := len(h) > 0
		for _, id := range h {
			n := s.Nodes[id]
			v, ok := s.Available[id]
			if n.Class == "OPAQUE_ROUTABLE" {
				r.Presentation = "OPAQUE"
				complete = false
				break
			}
			if !ok {
				complete = false
			} else if verified(v) {
				names[v.DisplayName] = true
			}
		}
		if complete && len(names) > 0 {
			r.Presentation = "VERIFIED"
			if len(names) > 1 {
				r.Presentation = "CONFLICT"
			}
		}
	}
	if !s.BaseSafe() {
		r.Ordinary = false
		r.Author = false
		r.Candidate = false
		r.CandidateWarning = false
	}
	return r
}
