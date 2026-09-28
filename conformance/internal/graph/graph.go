// Package graph interprets explicit accepted snapshots. Advisory history only supplies warnings.
package graph

import (
	"encoding/json"
	"errors"
	"reflect"
	"sort"
)

var ErrIntegrity = errors.New("accepted graph integrity failure")

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

// State contains accepted evidence for one observation context. It is not a journal.
type State struct {
	Nodes               map[string]Node
	Available           map[string]Value
	IntegrityFailure    bool
	DiscoveryIncomplete bool
	HistoryLost         bool
	HistoryExpected     bool
	CacheWriteFailed    bool
	Remembered          map[string]Node
}

func New() *State {
	return &State{Nodes: map[string]Node{}, Available: map[string]Value{}, Remembered: map[string]Node{}}
}

// Snapshot makes a detached planning snapshot; subsequent observations cannot mutate it.
func (s *State) Snapshot() *State {
	c := New()
	for id, n := range s.Nodes {
		c.Nodes[id] = clone(n)
	}
	for id, v := range s.Available {
		c.Available[id] = v
	}
	c.IntegrityFailure = s.IntegrityFailure
	return c
}
func (s *State) Remember() {
	for id, n := range s.Nodes {
		s.Remembered[id] = clone(n)
	}
	s.HistoryLost = false
	s.HistoryExpected = true
}

// LoseHistory requires meaningful evidence of previously retained/expected state.
func (s *State) LoseHistory() { s.Remembered = map[string]Node{}; s.HistoryLost = s.HistoryExpected }
func (s *State) ClearHistory() {
	s.Remembered = map[string]Node{}
	s.HistoryLost = false
	s.HistoryExpected = false
	s.CacheWriteFailed = false
}

func (s *State) Warnings() []string {
	var w []string
	if s.DiscoveryIncomplete {
		w = append(w, "PROCESSING_INCOMPLETE")
	}
	if s.HistoryLost {
		w = append(w, "HISTORY_MEMORY_LOST")
	}
	if s.CacheWriteFailed {
		w = append(w, "HISTORY_CACHE_WRITE_FAILED")
	}
	for id, old := range s.Remembered {
		incorporated := false
		if _, ok := s.Nodes[id]; ok {
			incorporated = true
		}
		// A direct signed claim incorporates this exact remembered ID even if its
		// bytes are unavailable. Do not traverse remembered/cache-only intermediates.
		for _, n := range s.Nodes {
			for _, p := range n.Parents {
				if p == id && n.Class != "OPAQUE_UNSCOPED" && old.Class != "OPAQUE_UNSCOPED" && n.Type == old.Type && n.Identity == old.Identity {
					incorporated = true
				}
			}
		}
		if !incorporated {
			w = append(w, "HISTORY_REGRESSION")
			break
		}
	}
	for _, n := range s.Nodes {
		if n.Class == "OPAQUE_UNSCOPED" {
			w = append(w, "UNKNOWN_FUTURE_EVIDENCE")
			break
		}
	}
	sort.Strings(w)
	return w
}
func clone(n Node) Node {
	n.Parents = append([]string(nil), n.Parents...)
	sort.Strings(n.Parents)
	return n
}
func (s *State) Learn(n Node, v *Value) error {
	n = clone(n)
	if old, ok := s.Nodes[n.ID]; ok && !reflect.DeepEqual(old, n) {
		s.IntegrityFailure = true
		return ErrIntegrity
	}
	if n.Class == "SUPPORTED_VALID" && v == nil {
		return errors.New("supported accepted observation requires complete value")
	}
	if old, ok := s.Available[n.ID]; ok && v != nil {
		previous, next := old, *v
		previous.Verified, next.Verified = false, false
		previous.Provenance, next.Provenance = "", ""
		if previous != next {
			s.IntegrityFailure = true
			return ErrIntegrity
		}
	}
	s.Nodes[n.ID] = n
	if v != nil && n.Class == "SUPPORTED_VALID" {
		s.Available[n.ID] = *v
	}
	if s.cyclic() {
		s.IntegrityFailure = true
		return ErrIntegrity
	}
	return nil
}

// Disappear applies a subsequent observation in which this object is absent.
func (s *State) Disappear(id string)         { delete(s.Available, id); delete(s.Nodes, id) }
func (s *State) RemoteUnavailable(id string) { s.Disappear(id) }

func (s *State) Rename(n Node, v Value) bool {
	if s.IntegrityFailure || n.Type != "DEVICE" || n.Class != "SUPPORTED_VALID" || !verified(v) {
		return false
	}
	heads := s.Heads("DEVICE", n.Identity)
	for _, id := range heads {
		if s.Nodes[id].Class != "SUPPORTED_VALID" {
			return false
		}
	}
	n.Parents = heads
	return s.Learn(n, &v) == nil
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

type Result struct {
	Warnings             []string `json:"warnings,omitempty"`
	RequiresConfirmation bool     `json:"requires_confirmation,omitempty"`
	Heads                []string `json:"heads"`
	ValueState           string   `json:"value_state"`
	Ordinary             bool     `json:"ordinary"`
	Author               bool     `json:"author"`
	Candidate            bool     `json:"candidate"`
	CandidateWarning     bool     `json:"candidate_warning"`
	IntegrityFailure     bool     `json:"integrity_failure"`
	Presentation         string   `json:"presentation,omitempty"`
}

func (s *State) Evaluate(identity, candidate, device string) Result {
	r := Result{Warnings: s.Warnings(), Heads: s.Heads("TOKEN", identity), ValueState: "EMPTY", IntegrityFailure: s.IntegrityFailure}
	opaque := false
	values := map[string]Value{}
	for _, id := range r.Heads {
		n := s.Nodes[id]
		if n.Class == "OPAQUE_ROUTABLE" {
			opaque = true
			continue
		}
		v := s.Available[id]
		v.DisplayName = ""
		v.Verified = false
		v.Provenance = ""
		b, _ := json.Marshal(v)
		values[string(b)] = v
	}
	switch {
	case opaque:
		r.ValueState = "VALUE_INCOMPLETE_OPAQUE"
	case len(values) > 1:
		r.ValueState = "CONFLICT"
	case len(values) == 1:
		r.ValueState = "UNAMBIGUOUS"
	}
	r.Author = !s.IntegrityFailure && !opaque && (r.ValueState == "UNAMBIGUOUS" || r.ValueState == "EMPTY")
	r.RequiresConfirmation = !s.IntegrityFailure && r.ValueState == "CONFLICT"
	if r.ValueState == "UNAMBIGUOUS" && !s.IntegrityFailure {
		for _, v := range values {
			r.Ordinary = v.Status == "LIVE"
		}
	}
	n, known := s.Nodes[candidate]
	v, available := s.Available[candidate]
	r.Candidate = !s.IntegrityFailure && known && available && n.Type == "TOKEN" && n.Identity == identity && n.Class == "SUPPORTED_VALID" && v.Status == "LIVE"
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
	if s.IntegrityFailure {
		r.Ordinary = false
		r.Author = false
		r.Candidate = false
		r.CandidateWarning = false
	}
	return r
}
