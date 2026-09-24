// Package referenceoracle reconstructs state by causal-closure projection.
// It shares only public DTO types with model, never its traversal or JOIN logic.
package referenceoracle

import (
	"fmt"
	"sort"
	"totipo/conformance/internal/model"
)

type oracle struct {
	objects     map[string]model.Update
	valid       map[string]string
	resolutions map[string]bool
	active      map[string]bool
	variant     string
}

func keys(m map[string]bool) []string {
	r := []string{}
	for k, v := range m {
		if v {
			r = append(r, k)
		}
	}
	sort.Strings(r)
	return r
}
func (o *oracle) closure(roots []string) map[string]bool {
	set := map[string]bool{}
	todo := append([]string{}, roots...)
	for len(todo) > 0 {
		id := todo[0]
		todo = todo[1:]
		if set[id] {
			continue
		}
		set[id] = true
		if u, ok := o.objects[id]; ok {
			todo = append(todo, u.Parents...)
		}
	}
	return set
}
func (o *oracle) before(a, b string) bool {
	if a == b {
		return false
	}
	return o.closure(o.objects[b].Parents)[a]
}
func (o *oracle) maximal(set map[string]bool) []string {
	out := []string{}
	for _, a := range keys(set) {
		hidden := false
		for b := range set {
			if o.before(a, b) {
				hidden = true
				break
			}
		}
		if !hidden {
			out = append(out, a)
		}
	}
	return out
}
func (o *oracle) heads(history map[string]bool, field string) []string {
	asserted := map[string]bool{}
	for id := range history {
		if _, ok := o.objects[id].Fields[field]; ok {
			asserted[id] = true
		}
	}
	return o.maximal(asserted)
}
func (o *oracle) state(history map[string]bool) model.State {
	out := model.State{}
	for _, f := range []string{"STATUS", "ISSUER", "ACCOUNT", "CREDENTIAL"} {
		out[f] = o.heads(history, f)
	}
	return out
}
func (o *oracle) transition(id string) bool {
	parents := o.heads(o.closure(o.objects[id].Parents), "STATUS")
	for _, p := range parents {
		if o.objects[p].Fields["STATUS"] != o.objects[id].Fields["STATUS"] {
			return true
		}
	}
	return false
}
func (o *oracle) conflicts(history map[string]bool, candidate string) []model.Witness {
	state := o.state(history)
	out := []model.Witness{}
	for _, h := range state["STATUS"] {
		statusPast := o.closure([]string{h})
		unresolved := map[string]bool{}
		transitions := map[string][]string{}
		for _, w := range keys(history) {
			if _, ok := o.objects[w].Fields["CREDENTIAL"]; !ok {
				continue
			}
			if o.variant == "current-heads-only" {
				current := false
				for _, ch := range state["CREDENTIAL"] {
					current = current || ch == w
				}
				if !current {
					continue
				}
			}
			covered := false
			for r := range statusPast {
				if !(o.resolutions[r] || r == candidate) {
					continue
				}
				if o.variant == "no-inherited-coverage" && r != h {
					continue
				}
				incorporated := o.before(w, r)
				if o.variant == "reversed-coverage" {
					incorporated = o.before(r, w)
				}
				if incorporated {
					covered = true
					break
				}
			}
			if covered {
				continue
			}
			for _, l := range keys(statusPast) {
				if _, ok := o.objects[l].Fields["STATUS"]; !ok {
					continue
				}
				if o.transition(l) && l != w && !o.before(l, w) && !o.before(w, l) {
					unresolved[w] = true
					transitions[w] = append(transitions[w], l)
				}
			}
		}
		for _, w := range o.maximal(unresolved) {
			out = append(out, model.Witness{Status: h, Credential: w, Transitions: transitions[w]})
		}
	}
	return out
}
func (o *oracle) check(id string) string {
	if s := o.valid[id]; s != "" {
		return s
	}
	u, ok := o.objects[id]
	if !ok {
		return "PENDING"
	}
	if o.active[id] {
		return "INVALID"
	}
	o.active[id] = true
	defer delete(o.active, id)
	finish := func(s string) string { o.valid[id] = s; return s }
	bad := len(u.Fields) == 0 || len(u.Parents) > 32
	for f, v := range u.Fields {
		switch f {
		case "STATUS":
			bad = bad || (v != "LIVE" && v != "TOMBSTONE")
		case "ISSUER", "ACCOUNT", "CREDENTIAL":
		default:
			bad = true
		}
	}
	pending := false
	seen := map[string]bool{}
	for _, p := range u.Parents {
		bad = bad || seen[p] || p == id
		seen[p] = true
		s := o.check(p)
		bad = bad || s == "INVALID"
		pending = pending || s == "PENDING"
		if parent, ok := o.objects[p]; ok && parent.Token != u.Token {
			bad = true
		}
	}
	if bad {
		return finish("INVALID")
	}
	if pending {
		return finish("PENDING")
	}
	for _, a := range u.Parents {
		for _, b := range u.Parents {
			if o.before(a, b) {
				return finish("INVALID")
			}
		}
	}
	past := o.closure(u.Parents)
	pre := o.conflicts(past, "")
	status, hasStatus := u.Fields["STATUS"]
	_, credential := u.Fields["CREDENTIAL"]
	if len(past) == 0 && (len(u.Fields) != 4 || status != "LIVE") {
		return finish("INVALID")
	}
	candidate := len(pre) > 0 && hasStatus
	if candidate && len(u.Fields) != 1 {
		return finish("INVALID")
	}
	if credential {
		if status == "TOMBSTONE" {
			return finish("INVALID")
		}
		for _, h := range o.heads(past, "STATUS") {
			if o.objects[h].Fields["STATUS"] != "LIVE" && status != "LIVE" {
				return finish("INVALID")
			}
		}
	}
	if candidate {
		post := o.closure([]string{id})
		if len(o.conflicts(post, id)) != 0 {
			return finish("INVALID")
		}
		o.resolutions[id] = true
	}
	return finish("FULLY_VALID")
}
func Evaluate(updates []model.Update) (model.View, error) { return EvaluateVariant(updates, "") }

// EvaluateVariant retains deliberate errors solely to demonstrate corpus sensitivity.
func EvaluateVariant(updates []model.Update, variant string) (model.View, error) {
	o := &oracle{map[string]model.Update{}, map[string]string{}, map[string]bool{}, map[string]bool{}, variant}
	for _, u := range updates {
		if u.ID == "" || u.Token == "" {
			return model.View{}, fmt.Errorf("identity")
		}
		if _, ok := o.objects[u.ID]; ok {
			return model.View{}, fmt.Errorf("duplicate ID")
		}
		o.objects[u.ID] = u
	}
	ids := map[string]bool{}
	for id := range o.objects {
		ids[id] = true
	}
	for _, id := range keys(ids) {
		o.check(id)
	}
	out := model.View{Validation: o.valid, Tokens: map[string]model.TokenView{}, Resolutions: keys(o.resolutions)}
	histories := map[string]map[string]bool{}
	for id, s := range o.valid {
		if s != "FULLY_VALID" {
			continue
		}
		token := o.objects[id].Token
		if histories[token] == nil {
			histories[token] = map[string]bool{}
		}
		histories[token][id] = true
	}
	for token, history := range histories {
		out.Tokens[token] = model.TokenView{Heads: o.maximal(history), Fields: o.state(history), Conflicts: o.conflicts(history, "")}
	}
	return out, nil
}

func ResolutionCoverage(updates []model.Update) (map[string][]string, error) {
	view, e := Evaluate(updates)
	if e != nil {
		return nil, e
	}
	o := &oracle{objects: map[string]model.Update{}}
	for _, u := range updates {
		o.objects[u.ID] = u
	}
	out := map[string][]string{}
	for _, r := range view.Resolutions {
		set := map[string]bool{}
		for id := range o.closure(o.objects[r].Parents) {
			if _, ok := o.objects[id].Fields["CREDENTIAL"]; ok {
				set[id] = true
			}
		}
		out[r] = keys(set)
	}
	return out, nil
}
