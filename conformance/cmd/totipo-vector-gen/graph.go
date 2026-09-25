package main

import (
	"totipo/conformance/internal/graph"
	"totipo/conformance/internal/vectors"
)

// Expected outcomes below are authored from sections 22–38, not calculated by
// the graph evaluator. Symbolic IDs stand for authenticated immutable objects.
func node(id string, parents ...string) *graph.Node {
	return &graph.Node{ID: id, Version: 1, Type: "TOKEN", Identity: "T", Class: "SUPPORTED_VALID", Parents: parents, Digest: "bytes-" + id}
}
func value(account string) *graph.Value {
	return &graph.Value{Status: "LIVE", Issuer: "Example", Account: account, Algorithm: 1, Digits: 6, Period: 30, Secret: "3132333435"}
}
func learn(n *graph.Node, v *graph.Value) vectors.Step {
	return vectors.Step{Action: "learn", Node: n, Value: v}
}
func gone(id string) vectors.Step { return vectors.Step{Action: "disappear", ID: id} }
func result(heads []string, state string, ordinary, author, candidate bool) graph.Result {
	return graph.Result{Heads: heads, ValueState: state, Ordinary: ordinary, Author: author, Candidate: candidate, CandidateWarning: candidate}
}
func query(r graph.Result) vectors.Step {
	return vectors.Step{Action: "query", Query: &vectors.Query{Identity: "T", Candidate: "a"}, Want: &r}
}
func opaque(id string, parents ...string) *graph.Node {
	n := node(id, parents...)
	n.Version = 2
	n.Class = "OPAQUE_ROUTABLE"
	return n
}
func unscoped() *graph.Node {
	return &graph.Node{ID: "u", Version: 3, Class: "OPAQUE_UNSCOPED", Digest: "bytes-u"}
}
func gc(id string, steps ...vectors.Step) {
	add(vectors.Case{ID: id, Operation: "graph", Expected: "PASS", Graph: &vectors.GraphCase{Steps: steps}}, "semantic", "22–25", "31–38")
}
func graphCases() {
	a, b := value("alice"), value("bob")
	root := func() vectors.Step { return learn(node("a"), a) }
	ordinary := result([]string{"b"}, "UNAMBIGUOUS", true, true, true)
	conflict := result([]string{"a", "b"}, "CONFLICT", false, true, true)
	missing := result([]string{"b"}, "VALUE_INCOMPLETE_UNAVAILABLE", false, true, true)
	opaqueB := result([]string{"b"}, "VALUE_INCOMPLETE_OPAQUE", false, false, true)
	gc("v1.graph.sequential.001", root(), learn(node("b", "a"), b), query(ordinary))
	gc("v1.graph.equal-concurrent.001", root(), learn(node("b"), a), query(result([]string{"a", "b"}, "UNAMBIGUOUS", true, true, true)))
	gc("v1.graph.conflicting-concurrent.001", root(), learn(node("b"), b), query(conflict))
	gc("v1.graph.missing-current-value.001", root(), learn(node("b", "a"), b), gone("b"), query(missing))
	gc("v1.graph.intermediate-disappears.001", root(), learn(node("b", "a"), a), learn(node("c", "b"), b), gone("b"), query(result([]string{"c"}, "UNAMBIGUOUS", true, true, true)))
	gc("v1.graph.late-parent.001", learn(node("b", "a"), b), query(result([]string{"b"}, "UNAMBIGUOUS", true, true, false)), root(), query(ordinary))
	bad := learn(node("b", "a"), b)
	bad.IntegrityError = true
	r := result([]string{}, "EMPTY", false, false, false)
	r.IntegrityFailure = true
	gc("v1.graph.cycle-integrity-failure.001", learn(node("a", "b"), a), bad, query(r))
	other := node("a")
	other.Type = "DEVICE"
	other.Identity = "D"
	other.Digest = "different-bytes"
	bad = learn(other, nil)
	bad.IntegrityError = true
	r = result([]string{"a"}, "UNAMBIGUOUS", false, false, false)
	r.IntegrityFailure = true
	gc("v1.graph.global-object-id-conflict.001", root(), bad, query(r))
	gc("v1.future.scoped-token.001", root(), learn(opaque("b", "a"), nil), query(opaqueB))
	gc("v1.future.concurrent-supported-opaque.001", root(), learn(opaque("b"), nil), query(result([]string{"a", "b"}, "VALUE_INCOMPLETE_OPAQUE", false, false, true)))
	unrelated := opaque("z")
	unrelated.Identity = "OTHER"
	gc("v1.future.unrelated-token.001", root(), learn(unrelated, nil), query(result([]string{"a"}, "UNAMBIGUOUS", true, true, true)))
	gc("v1.future.supported-descendant.001", root(), learn(opaque("b", "a"), nil), query(opaqueB), learn(node("c", "b"), b), query(result([]string{"c"}, "UNAMBIGUOUS", true, true, true)))
	future := opaque("a")
	future.Version = 255
	future.AuthorTime = ^uint64(0)
	low := node("b", "a")
	low.AuthorTime = 0
	gc("v1.future.version-does-not-order.001", learn(future, nil), learn(low, b), query(result([]string{"b"}, "UNAMBIGUOUS", true, true, false)))
	d := opaque("d")
	d.Type = "DEVICE"
	d.Identity = "D"
	r = result([]string{"a"}, "UNAMBIGUOUS", true, true, true)
	r.Presentation = "OPAQUE"
	q := query(r)
	q.Query.Device = "D"
	gc("v1.future.device-presentation.001", root(), learn(d, nil), q)
	for _, id := range []string{"v1.future.unscoped-authoritative-block.001", "v1.future.unscoped-candidate-use.001"} {
		gc(id, root(), learn(unscoped(), nil), gone("u"), query(result([]string{"a"}, "UNAMBIGUOUS", false, false, true)))
	}
	gc("v1.future.disappearance-retains-routing.001", root(), learn(opaque("b", "a"), nil), gone("b"), query(opaqueB))
	gc("v1.candidate.conflict-a.001", root(), learn(node("b"), b), query(conflict))
	gc("v1.candidate.current-peer-missing.001", root(), learn(node("b"), b), gone("b"), query(result([]string{"a", "b"}, "VALUE_INCOMPLETE_UNAVAILABLE", false, true, true)))
	gc("v1.candidate.historical-current-missing.001", root(), learn(node("b", "a"), b), gone("b"), query(missing))
	gc("v1.candidate.opaque-current.001", root(), learn(opaque("b", "a"), nil), query(opaqueB))
	gc("v1.candidate.discovery-incomplete.001", root(), vectors.Step{Action: "discovery-incomplete", Flag: true}, query(result([]string{"a"}, "UNAMBIGUOUS", false, false, true)))
	gc("v1.candidate.persistence-block.001", root(), vectors.Step{Action: "persist-fails", Node: node("b", "a"), Value: b}, query(result([]string{"a"}, "UNAMBIGUOUS", false, false, false)))
	r = result([]string{"a"}, "UNAMBIGUOUS", false, false, false)
	r.IntegrityFailure = true
	gc("v1.candidate.continuity-block.001", root(), vectors.Step{Action: "continuity-unknown", Flag: true}, query(r))
	old, new := node("a"), node("b")
	old.AuthorTime = 1
	new.AuthorTime = ^uint64(0)
	gc("v1.timestamp.equal-value-different-times.001", learn(old, a), learn(new, a), query(result([]string{"a", "b"}, "UNAMBIGUOUS", true, true, true)))
	first, last := node("b", "a"), node("c", "b")
	first.AuthorTime = 1700000000
	last.AuthorTime = 1700000000
	gc("v1.timestamp.fold-common-time.001", root(), learn(first, b), learn(last, b), query(result([]string{"c"}, "UNAMBIGUOUS", true, true, true)))
	// Additional edge, provenance-presentation, and durable reappearance checks.
	wrong := node("p")
	wrong.Identity = "OTHER"
	gc("v1.graph.wrong-identity-parent.001", learn(wrong, a), learn(node("a", "p"), a), query(result([]string{"a"}, "UNAMBIGUOUS", true, true, true)))
	gc("v1.graph.reappearance.001", root(), gone("a"), root(), query(result([]string{"a"}, "UNAMBIGUOUS", true, true, true)))
	changed := node("a")
	changed.AuthorTime = 7
	bad = learn(changed, a)
	bad.IntegrityError = true
	r = result([]string{"a"}, "UNAMBIGUOUS", false, false, false)
	r.IntegrityFailure = true
	gc("v1.graph.reappearance-mismatch.001", root(), bad, query(r))
}
