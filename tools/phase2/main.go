// Phase 2 authored fixture maintenance. Expected heads/witnesses are explicit
// below; this command imports neither primary model nor reference oracle.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"sort"
)

type M = map[string]string
type U struct {
	ID      string   `json:"id"`
	Token   string   `json:"token"`
	Parents []string `json:"parents"`
	Fields  M        `json:"fields"`
	Signer  string   `json:"signer,omitempty"`
}
type W struct {
	Status      string   `json:"status"`
	Credential  string   `json:"credential"`
	Transitions []string `json:"transitions"`
}
type TV struct {
	Heads     []string            `json:"heads"`
	Fields    map[string][]string `json:"fields"`
	Conflicts []W                 `json:"conflicts"`
}
type V struct {
	Validation  M             `json:"validation"`
	Tokens      map[string]TV `json:"tokens"`
	Resolutions []string      `json:"resolutions"`
}
type S struct {
	Updates     []U                 `json:"updates"`
	Expected    V                   `json:"expected"`
	Unavailable []string            `json:"unavailable,omitempty"`
	Observed    []string            `json:"observed,omitempty"`
	Coverage    map[string][]string `json:"coverage,omitempty"`
}
type P struct {
	Source  string `json:"source"`
	Status  string `json:"status"`
	Locator string `json:"locator"`
}
type C struct {
	ID          string `json:"id"`
	Kind        string `json:"kind"`
	Description string `json:"description"`
	Provenance  P      `json:"provenance"`
	Input       M      `json:"input"`
	Expected    M      `json:"expected"`
	Scenario    *S     `json:"scenario,omitempty"`
}

func write(path string, v any) {
	b, e := json.MarshalIndent(v, "", "  ")
	if e != nil {
		panic(e)
	}
	if e = os.WriteFile(path, append(b, '\n'), 0644); e != nil {
		panic(e)
	}
}
func one(s string) []string { return []string{s} }
func state(s, i, a, c []string) map[string][]string {
	return map[string][]string{"STATUS": s, "ISSUER": i, "ACCOUNT": a, "CREDENTIAL": c}
}
func update(id string, parents []string, f M) U { return U{id, "t", parents, f, "device-a"} }
func main() {
	yes := flag.Bool("write-vectors", false, "explicit maintenance")
	flag.Parse()
	if !*yes {
		panic("requires --write-vectors")
	}
	tokenFixtures()
	memoryFixtures()
	deviceFixtures()
	safetyFixtures()
}
func tokenFixtures() {
	var cases []C
	empty := []W{}
	nores := []string{}
	root := update("r", []string{}, M{"STATUS": "LIVE", "ISSUER": "i", "ACCOUNT": "a", "CREDENTIAL": "k0"})
	add := func(id, section string, us []U, heads []string, fields map[string][]string, ws []W, rs []string, invalid M) {
		statuses := M{}
		for _, u := range us {
			statuses[u.ID] = "FULLY_VALID"
		}
		for k, v := range invalid {
			statuses[k] = v
		}
		tokens := map[string]TV{}
		if len(heads) > 0 {
			tokens["t"] = TV{heads, fields, ws}
		}
		cases = append(cases, C{"v0/transitions/" + id, "model", id, P{"review/phase2/REVIEW.md", "spec-derived-reviewed", "r36 " + section + "; " + id}, M{}, M{"disposition": "VERIFIED"}, &S{Updates: us, Expected: V{statuses, tokens, rs}}})
	}
	for _, missing := range []string{"STATUS", "ISSUER", "ACCOUNT", "CREDENTIAL"} {
		f := M{}
		for k, v := range root.Fields {
			if k != missing {
				f[k] = v
			}
		}
		bad := update("bad", []string{}, f)
		add("creation-missing-"+missing, "28", []U{bad}, []string{}, nil, empty, nores, M{"bad": "INVALID"})
	}
	for _, field := range []string{"ISSUER", "ACCOUNT", "CREDENTIAL"} {
		u := update("e", one("r"), M{field: "new"})
		s := state(one("r"), one("r"), one("r"), one("r"))
		s[field] = one("e")
		add("edit-"+field, "23-31", []U{root, u}, one("e"), s, empty, nores, nil)
	}
	both := update("e", one("r"), M{"ISSUER": "new", "ACCOUNT": "new"})
	add("combined-fields", "29", []U{root, both}, one("e"), state(one("r"), one("e"), one("e"), one("r")), empty, nores, nil)
	i := update("i", one("r"), M{"ISSUER": "i2"})
	w := update("w", one("r"), M{"CREDENTIAL": "k1"})
	add("issuer-credential-compose", "27", []U{root, i, w}, []string{"i", "w"}, state(one("r"), one("i"), one("r"), one("w")), empty, nores, nil)
	a := update("a", one("r"), M{"ISSUER": "left"})
	b := update("b", one("r"), M{"ISSUER": "right"})
	add("different-values-conflict", "30", []U{root, a, b}, []string{"a", "b"}, state(one("r"), []string{"a", "b"}, one("r"), one("r")), empty, nores, nil)
	m := update("m", []string{"a", "b"}, M{"ISSUER": "chosen"})
	add("complete-visible-field-merge", "24,31", []U{root, a, b, m}, one("m"), state(one("r"), one("m"), one("r"), one("r")), empty, nores, nil)
	hidden := update("h", one("a"), M{"ISSUER": "chosen"})
	add("omitted-branch-not-retroactive-invalidity", "22,31,50", []U{root, a, b, hidden}, []string{"b", "h"}, state(one("r"), []string{"b", "h"}, one("r"), one("r")), empty, nores, nil)
	d := update("d", one("r"), M{"STATUS": "TOMBSTONE"})
	s := update("s", one("d"), M{"STATUS": "LIVE"})
	restoreWitness := update("w", one("d"), M{"STATUS": "LIVE", "CREDENTIAL": "k1"})
	add("restoration-credential-race", "32-36,41", []U{root, d, s, restoreWitness}, []string{"s", "w"}, state([]string{"s", "w"}, one("r"), one("r"), one("w")), []W{{"s", "w", one("s")}}, nores, nil)
	ws := []W{{"d", "w", one("d")}}
	for _, f := range []string{"ISSUER", "ACCOUNT"} {
		u := update("e", []string{"d", "w"}, M{f: "edit"})
		st := state(one("d"), one("r"), one("r"), one("w"))
		st[f] = one("e")
		add("conflict-edit-"+f, "37", []U{root, d, w, u}, one("e"), st, ws, nores, nil)
	}
	w2 := update("w2", one("r"), M{"CREDENTIAL": "k2"})
	add("multiple-witness-branches", "36", []U{root, d, w, w2}, []string{"d", "w", "w2"}, state(one("d"), one("r"), one("r"), []string{"w", "w2"}), []W{{"d", "w", one("d")}, {"d", "w2", one("d")}}, nores, nil)
	d2 := update("d2", one("r"), M{"STATUS": "TOMBSTONE"})
	add("multiple-equal-status-heads", "30,32,36", []U{root, d, d2, w}, []string{"d", "d2", "w"}, state([]string{"d", "d2"}, one("r"), one("r"), one("w")), []W{{"d", "w", one("d")}, {"d2", "w", one("d2")}}, nores, nil)
	live := update("live", one("r"), M{"STATUS": "LIVE"})
	add("ordinary-status-conflict-no-witness", "32", []U{root, d, live}, []string{"d", "live"}, state([]string{"d", "live"}, one("r"), one("r"), one("r")), empty, nores, nil)
	many := update("many", []string{"d", "live"}, M{"STATUS": "LIVE"})
	add("multi-parent-status-transition", "32", []U{root, d, live, many, w}, []string{"many", "w"}, state(one("many"), one("r"), one("r"), one("w")), []W{{"many", "w", []string{"d", "many"}}}, nores, nil)
	resolve := update("x", []string{"d", "w"}, M{"STATUS": "TOMBSTONE"})
	add("same-scalar-resolution", "34,38,51", []U{root, d, w, resolve}, one("x"), state(one("x"), one("r"), one("r"), one("w")), empty, one("x"), nil)
	descendant := update("y", one("x"), M{"STATUS": "LIVE"})
	add("coverage-inherited-by-status-descendant", "34", []U{root, d, w, resolve, descendant}, one("y"), state(one("y"), one("r"), one("r"), one("w")), empty, one("x"), nil)
	manyRes := update("x", []string{"d", "d2", "w", "w2"}, M{"STATUS": "LIVE"})
	add("resolve-all-status-and-witness-branches", "38,51", []U{root, d, d2, w, w2, manyRes}, one("x"), state(one("x"), one("r"), one("r"), []string{"w", "w2"}), empty, one("x"), nil)
	for _, f := range []string{"ISSUER", "ACCOUNT", "CREDENTIAL"} {
		bad := update("x", []string{"d", "w"}, M{"STATUS": "LIVE", f: "bad"})
		add("pre-nonempty-status-plus-"+f, "37,51", []U{root, d, w, bad}, []string{"d", "w"}, state(one("d"), one("r"), one("r"), one("w")), ws, nores, M{"x": "INVALID"})
	}
	bad := update("bad", one("r"), M{"STATUS": "TOMBSTONE", "CREDENTIAL": "bad"})
	add("tombstone-plus-credential", "41", []U{root, bad}, one("r"), state(one("r"), one("r"), one("r"), one("r")), empty, nores, M{"bad": "INVALID"})
	credential := update("c", []string{"d", "live"}, M{"CREDENTIAL": "bad"})
	add("any-tombstone-blocks-credential-only", "41", []U{root, d, live, credential}, []string{"d", "live"}, state([]string{"d", "live"}, one("r"), one("r"), one("r")), empty, nores, M{"c": "INVALID"})
	ordinary := update("o", one("r"), M{"STATUS": "LIVE", "ACCOUNT": "new"})
	add("pre-empty-status-plus-account", "50", []U{root, ordinary}, one("o"), state(one("o"), one("r"), one("o"), one("r")), empty, nores, nil)
	informed := update("c", []string{"s", "w"}, M{"CREDENTIAL": "k2"})
	add("historical-witness-current-head-counterexample", "33-36,41", []U{root, d, s, w, informed}, one("c"), state(one("s"), one("r"), one("r"), one("c")), []W{{"s", "w", []string{"d", "s"}}}, nores, nil)
	add("hidden-branch-selected-view", "22,49,56", []U{root, d, w}, one("d"), state(one("d"), one("r"), one("r"), one("r")), empty, nores, nil)
	cases[len(cases)-1].Scenario.Observed = one("d")
	delete(cases[len(cases)-1].Scenario.Expected.Validation, "w")
	add("dependency-unavailable", "56,57", []U{root, w}, []string{}, nil, empty, nores, M{"w": "PENDING"})
	cases[len(cases)-1].Scenario.Unavailable = one("r")
	delete(cases[len(cases)-1].Scenario.Expected.Validation, "r")
	differentSigner := b
	differentSigner.Signer = "device-b"
	add("signer-does-not-create-causality", "46", []U{root, a, differentSigner}, []string{"a", "b"}, state(one("r"), []string{"a", "b"}, one("r"), one("r")), empty, nores, nil)
	for n := range cases {
		switch cases[n].ID {
		case "v0/transitions/same-scalar-resolution", "v0/transitions/coverage-inherited-by-status-descendant":
			cases[n].Scenario.Coverage = map[string][]string{"x": {"r", "w"}}
		case "v0/transitions/resolve-all-status-and-witness-branches":
			cases[n].Scenario.Coverage = map[string][]string{"x": {"r", "w", "w2"}}
		}
	}
	// Lowercase IDs are stable independent of display enum spelling.
	for n := range cases {
		for i, c := range cases[n].ID {
			if c >= 'A' && c <= 'Z' {
				b := []byte(cases[n].ID)
				b[i] = byte(c + 32)
				cases[n].ID = string(b)
			}
		}
	}
	sort.Slice(cases, func(i, j int) bool { return cases[i].ID < cases[j].ID })
	write("vectors/v0/transitions/phase2.json", struct {
		Schema int `json:"schema"`
		Cases  []C `json:"cases"`
	}{1, cases})
	fmt.Println("authored", len(cases), "transition cases")
}
