// Package model is a literal, symbolic, post-authentication token model for r36.
// Inputs stand for identity-matching, structurally valid, signature-verified objects.
// It does not perform crypto, publication, durable recovery, or human confirmation.
package model

import (
	"fmt"
	"sort"
)

var Fields = []string{"STATUS", "ISSUER", "ACCOUNT", "CREDENTIAL"}

type Update struct {
	ID      string            `json:"id"`
	Token   string            `json:"token"`
	Parents []string          `json:"parents"`
	Fields  map[string]string `json:"fields"`
}
type State map[string][]string
type Witness struct {
	Status      string   `json:"status"`
	Credential  string   `json:"credential"`
	Transitions []string `json:"transitions"`
}
type TokenView struct {
	Heads     []string  `json:"heads"`
	Fields    State     `json:"fields"`
	Conflicts []Witness `json:"conflicts"`
}
type View struct {
	Validation  map[string]string    `json:"validation"`
	Tokens      map[string]TokenView `json:"tokens"`
	Resolutions []string             `json:"resolutions"`
}
type node struct {
	Update
	State        State
	FieldParents State
	Resolution   bool
}
type evaluator struct {
	input     map[string]Update
	nodes     map[string]*node
	status    map[string]string
	visiting  map[string]bool
	postCheck func([]Witness) []Witness
}

func sortedSet(xs []string) []string {
	m := map[string]bool{}
	for _, x := range xs {
		m[x] = true
	}
	out := []string{}
	for x := range m {
		out = append(out, x)
	}
	sort.Strings(out)
	return out
}
func emptyState() State {
	s := State{}
	for _, f := range Fields {
		s[f] = []string{}
	}
	return s
}
func (e *evaluator) causal(a, b string) bool {
	if a == b {
		return true
	}
	seen := map[string]bool{}
	var visit func(string) bool
	visit = func(id string) bool {
		if seen[id] {
			return false
		}
		seen[id] = true
		u, ok := e.input[id]
		if !ok {
			return false
		}
		for _, p := range u.Parents {
			if p == a || visit(p) {
				return true
			}
		}
		return false
	}
	return visit(b)
}
func (e *evaluator) fieldHistory(id, f string) []string {
	seen := map[string]bool{}
	var visit func(string)
	visit = func(x string) {
		if seen[x] {
			return
		}
		seen[x] = true
		n := e.nodes[x]
		if n == nil {
			return
		}
		for _, p := range n.FieldParents[f] {
			visit(p)
		}
	}
	visit(id)
	out := []string{}
	for x := range seen {
		out = append(out, x)
	}
	return sortedSet(out)
}
func (e *evaluator) fieldBefore(a, b, f string) bool {
	for _, x := range e.fieldHistory(b, f) {
		if x == a {
			return true
		}
	}
	return false
}
func (e *evaluator) maximal(ids []string, before func(string, string) bool) []string {
	out := []string{}
	for _, a := range sortedSet(ids) {
		keep := true
		for _, b := range ids {
			if a != b && before(a, b) {
				keep = false
				break
			}
		}
		if keep {
			out = append(out, a)
		}
	}
	return out
}
func (e *evaluator) join(states ...State) State {
	s := emptyState()
	for _, f := range Fields {
		var ids []string
		for _, state := range states {
			ids = append(ids, state[f]...)
		}
		s[f] = e.maximal(ids, func(a, b string) bool { return e.fieldBefore(a, b, f) })
	}
	return s
}
func (e *evaluator) transition(id string) bool {
	n := e.nodes[id]
	for _, p := range n.FieldParents["STATUS"] {
		if e.nodes[p].Fields["STATUS"] != n.Fields["STATUS"] {
			return true
		}
	}
	return false
}
func (e *evaluator) conflicts(s State, candidate string) []Witness {
	history := []string{}
	for _, h := range s["CREDENTIAL"] {
		history = append(history, e.fieldHistory(h, "CREDENTIAL")...)
	}
	history = sortedSet(history)
	out := []Witness{}
	for _, h := range s["STATUS"] {
		statusHistory := e.fieldHistory(h, "STATUS")
		raw := map[string][]string{}
		for _, w := range history {
			confirmed := false
			for _, r := range statusHistory {
				if (e.nodes[r].Resolution || r == candidate) && w != r && e.causal(w, r) {
					confirmed = true
					break
				}
			}
			if confirmed {
				continue
			}
			for _, l := range statusHistory {
				if e.transition(l) && !e.causal(l, w) && !e.causal(w, l) {
					raw[w] = append(raw[w], l)
				}
			}
		}
		ids := []string{}
		for w := range raw {
			ids = append(ids, w)
		}
		for _, w := range e.maximal(ids, func(a, b string) bool { return e.fieldBefore(a, b, "CREDENTIAL") }) {
			out = append(out, Witness{h, w, sortedSet(raw[w])})
		}
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Status == out[j].Status {
			return out[i].Credential < out[j].Credential
		}
		return out[i].Status < out[j].Status
	})
	return out
}
func (e *evaluator) validate(id string) string {
	if s := e.status[id]; s != "" {
		return s
	}
	u, exists := e.input[id]
	if !exists {
		return "PENDING"
	}
	if e.visiting[id] {
		return "INVALID"
	}
	e.visiting[id] = true
	defer delete(e.visiting, id)
	finish := func(s string) string { e.status[id] = s; return s }
	if len(u.Fields) == 0 || len(u.Parents) > 32 {
		return finish("INVALID")
	}
	for f, v := range u.Fields {
		if f != "STATUS" && f != "ISSUER" && f != "ACCOUNT" && f != "CREDENTIAL" {
			return finish("INVALID")
		}
		if f == "STATUS" && v != "LIVE" && v != "TOMBSTONE" {
			return finish("INVALID")
		}
	}
	seen := map[string]bool{}
	pending, invalid := false, false
	for _, p := range u.Parents {
		if seen[p] || p == id {
			invalid = true
		}
		seen[p] = true
		s := e.validate(p)
		if s == "INVALID" {
			invalid = true
		}
		if s == "PENDING" {
			pending = true
		}
		if parent, ok := e.input[p]; ok && parent.Token != u.Token {
			invalid = true
		}
	}
	if invalid {
		return finish("INVALID")
	}
	if pending {
		return finish("PENDING")
	}
	for _, a := range u.Parents {
		for _, b := range u.Parents {
			if a != b && e.causal(a, b) {
				return finish("INVALID")
			}
		}
	}
	bases := []State{}
	for _, p := range u.Parents {
		bases = append(bases, e.nodes[p].State)
	}
	base := e.join(bases...)
	pre := e.conflicts(base, "")
	status, assertsStatus := u.Fields["STATUS"]
	_, credential := u.Fields["CREDENTIAL"]
	if len(u.Parents) == 0 {
		if len(u.Fields) != 4 || status != "LIVE" {
			return finish("INVALID")
		}
	}
	candidate := len(pre) > 0 && assertsStatus
	if candidate && len(u.Fields) != 1 {
		return finish("INVALID")
	}
	if credential {
		if status == "TOMBSTONE" {
			return finish("INVALID")
		}
		for _, h := range base["STATUS"] {
			if e.nodes[h].Fields["STATUS"] == "TOMBSTONE" && status != "LIVE" {
				return finish("INVALID")
			}
		}
	}
	n := &node{Update: u, State: e.join(base), FieldParents: emptyState()}
	for f := range u.Fields {
		n.FieldParents[f] = append([]string{}, base[f]...)
		n.State[f] = []string{id}
	}
	e.nodes[id] = n
	if candidate {
		post := e.conflicts(n.State, id)
		if e.postCheck != nil {
			post = e.postCheck(post)
		}
		if len(post) > 0 {
			delete(e.nodes, id)
			return finish("INVALID")
		}
		n.Resolution = true
	}
	return finish("FULLY_VALID")
}
func Evaluate(updates []Update) (View, error) { return evaluate(updates, nil) }
func evaluate(updates []Update, postCheck func([]Witness) []Witness) (View, error) {
	e := &evaluator{map[string]Update{}, map[string]*node{}, map[string]string{}, map[string]bool{}, postCheck}
	ids := []string{}
	for _, u := range updates {
		if u.ID == "" || u.Token == "" {
			return View{}, fmt.Errorf("missing symbolic identity")
		}
		if _, ok := e.input[u.ID]; ok {
			return View{}, fmt.Errorf("duplicate symbolic ID %s", u.ID)
		}
		e.input[u.ID] = u
		ids = append(ids, u.ID)
	}
	sort.Strings(ids)
	for _, id := range ids {
		e.validate(id)
	}
	view := View{e.status, map[string]TokenView{}, []string{}}
	tokens := map[string][]string{}
	for _, id := range ids {
		if n := e.nodes[id]; n != nil && e.status[id] == "FULLY_VALID" {
			tokens[n.Token] = append(tokens[n.Token], id)
			if n.Resolution {
				view.Resolutions = append(view.Resolutions, id)
			}
		}
	}
	for token, ids := range tokens {
		heads := e.maximal(ids, e.causal)
		states := []State{}
		for _, h := range heads {
			states = append(states, e.nodes[h].State)
		}
		s := e.join(states...)
		view.Tokens[token] = TokenView{heads, s, e.conflicts(s, "")}
	}
	return view, nil
}

// ResolutionCoverage exposes the complete historical credential revisions
// incorporated by each accepted resolution, including non-head revisions.
func ResolutionCoverage(updates []Update) (map[string][]string, error) {
	view, err := Evaluate(updates)
	if err != nil {
		return nil, err
	}
	objects := map[string]Update{}
	for _, u := range updates {
		objects[u.ID] = u
	}
	out := map[string][]string{}
	for _, r := range view.Resolutions {
		seen := map[string]bool{}
		todo := append([]string{}, objects[r].Parents...)
		credentials := []string{}
		for len(todo) > 0 {
			id := todo[len(todo)-1]
			todo = todo[:len(todo)-1]
			if seen[id] {
				continue
			}
			seen[id] = true
			u := objects[id]
			if _, ok := u.Fields["CREDENTIAL"]; ok {
				credentials = append(credentials, id)
			}
			todo = append(todo, u.Parents...)
		}
		out[r] = sortedSet(credentials)
	}
	return out, nil
}
