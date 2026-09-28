package graph

import (
	"encoding/json"
	"errors"
)

// Plan is a decision against a detached snapshot, not a publication transaction.
// Context includes only the selected TOKEN's heads and complete semantic values.
type Plan struct {
	Node      Node
	Value     Value
	Intent    string
	confirmed bool
	context   string
	changed   bool
}

func semantic(v Value) Value { v.DisplayName = ""; v.Verified = false; v.Provenance = ""; return v }
func (s *State) context(identity string, desired Value, intent string) string {
	type alternative struct {
		ID    string
		Class string
		Value Value
	}
	a := []alternative{}
	for _, id := range s.Heads("TOKEN", identity) {
		a = append(a, alternative{id, s.Nodes[id].Class, semantic(s.Available[id])})
	}
	b, _ := json.Marshal(struct {
		Identity     string
		Alternatives []alternative
		Desired      Value
		Intent       string
	}{identity, a, semantic(desired), intent})
	return string(b)
}
func (s *State) Plan(n Node, v Value, intent string, confirmed bool) (*Plan, error) {
	snapshot := s.Snapshot()
	r := snapshot.Evaluate(n.Identity, "", "")
	if n.Type != "TOKEN" || n.Class != "SUPPORTED_VALID" || (!r.Author && !(r.RequiresConfirmation && confirmed)) {
		return nil, errors.New("decision unavailable or requires confirmation")
	}
	n = clone(n)
	n.Parents = snapshot.Heads("TOKEN", n.Identity)
	return &Plan{Node: n, Value: v, Intent: intent, confirmed: r.RequiresConfirmation, context: snapshot.context(n.Identity, v, intent)}, nil
}

// Observe records materially relevant changes the client actually learned. A
// restored-looking context does not erase information learned after confirmation.
func (p *Plan) Observe(s *State) {
	if p.confirmed && s.context(p.Node.Identity, p.Value, p.Intent) != p.context {
		p.changed = true
	}
}
func (p *Plan) Publish(s *State, acknowledged bool) bool {
	p.Observe(s)
	if s.IntegrityFailure || p.changed || !acknowledged {
		return false
	}
	return s.Learn(p.Node, &p.Value) == nil
}
