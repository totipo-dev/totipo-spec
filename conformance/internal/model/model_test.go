package model

import (
	"fmt"
	"math/rand"
	"reflect"
	"testing"
)

func history() []Update {
	return []Update{
		{"r", "t", []string{}, map[string]string{"STATUS": "LIVE", "ISSUER": "i", "ACCOUNT": "a", "CREDENTIAL": "c"}},
		{"d", "t", []string{"r"}, map[string]string{"STATUS": "TOMBSTONE"}},
		{"w", "t", []string{"r"}, map[string]string{"CREDENTIAL": "c2"}},
		{"x", "t", []string{"d", "w"}, map[string]string{"STATUS": "LIVE"}},
	}
}
func TestArrivalOrder(t *testing.T) {
	h := history()
	want, e := Evaluate(h)
	if e != nil {
		t.Fatal(e)
	}
	var permute func(int)
	permute = func(i int) {
		if i == len(h) {
			got, e := Evaluate(h)
			if e != nil || !reflect.DeepEqual(got, want) {
				t.Fatalf("order-dependent: %v %v", got, e)
			}
			return
		}
		for j := i; j < len(h); j++ {
			h[i], h[j] = h[j], h[i]
			permute(i + 1)
			h[i], h[j] = h[j], h[i]
		}
	}
	permute(0)
}
func TestFailedCandidateContributesNoCoverage(t *testing.T) {
	got, e := evaluate(history(), func(post []Witness) []Witness {
		return append(post, Witness{"injected", "invariant-failure", []string{}})
	})
	if e != nil {
		t.Fatal(e)
	}
	if got.Validation["x"] != "INVALID" || len(got.Resolutions) != 0 || len(got.Tokens["t"].Conflicts) != 1 {
		t.Fatalf("candidate coverage escaped: %+v", got)
	}
}
func TestPendingAndInvalidDependencies(t *testing.T) {
	h := history()[:1]
	h = append(h, Update{"bad", "t", []string{}, map[string]string{"STATUS": "TOMBSTONE"}}, Update{"p", "t", []string{"missing", "bad"}, map[string]string{"ISSUER": "i"}})
	got, e := Evaluate(h)
	if e != nil {
		t.Fatal(e)
	}
	if got.Validation["p"] != "INVALID" {
		t.Fatal("missing dependency concealed invalid dependency")
	}
	h = []Update{{"a", "t", []string{"b"}, map[string]string{"ISSUER": "i"}}, {"b", "t", []string{"a"}, map[string]string{"ISSUER": "i"}}}
	got, e = Evaluate(h)
	if e != nil || got.Validation["a"] != "INVALID" || got.Validation["b"] != "INVALID" {
		t.Fatalf("cycle: %+v %v", got, e)
	}
}
func TestJoinLawsAndTypedAncestry(t *testing.T) {
	h := history()[:3]
	e := &evaluator{input: map[string]Update{}, nodes: map[string]*node{}, status: map[string]string{}, visiting: map[string]bool{}}
	for _, u := range h {
		e.input[u.ID] = u
	}
	for _, u := range h {
		e.validate(u.ID)
	}
	a, b, c := e.nodes["r"].State, e.nodes["d"].State, e.nodes["w"].State
	if !reflect.DeepEqual(e.join(a, a), a) || !reflect.DeepEqual(e.join(b, c), e.join(c, b)) || !reflect.DeepEqual(e.join(e.join(a, b), c), e.join(a, e.join(b, c))) {
		t.Fatal("JOIN laws")
	}
	for id, n := range e.nodes {
		for f := range n.Fields {
			for _, a := range e.fieldHistory(id, f) {
				if !e.causal(a, id) {
					t.Fatal("field ancestry without causal ancestry")
				}
			}
		}
	}
}

func TestRandomizedJoinAndAncestry(t *testing.T) {
	triples, edges := 0, 0
	for seed := int64(0); seed < 12; seed++ {
		rng := rand.New(rand.NewSource(36000 + seed))
		e := &evaluator{input: map[string]Update{}, nodes: map[string]*node{}, status: map[string]string{}, visiting: map[string]bool{}}
		root := history()[0]
		e.input[root.ID] = root
		e.validate(root.ID)
		ids := []string{root.ID}
		for step := 0; step < 18; step++ {
			id := fmt.Sprintf("u%02d", step)
			field := []string{"ISSUER", "ACCOUNT", "CREDENTIAL"}[rng.Intn(3)]
			parent := ids[rng.Intn(len(ids))]
			e.input[id] = Update{ID: id, Token: "t", Parents: []string{parent}, Fields: map[string]string{field: fmt.Sprintf("value%d", rng.Intn(4))}}
			if e.validate(id) != "FULLY_VALID" {
				t.Fatal("valid field edit rejected")
			}
			ids = append(ids, id)
			a := e.nodes[ids[rng.Intn(len(ids))]].State
			b := e.nodes[ids[rng.Intn(len(ids))]].State
			c := e.nodes[ids[rng.Intn(len(ids))]].State
			if !reflect.DeepEqual(e.join(a, a), a) || !reflect.DeepEqual(e.join(a, b), e.join(b, a)) || !reflect.DeepEqual(e.join(e.join(a, b), c), e.join(a, e.join(b, c))) {
				t.Fatalf("JOIN law seed %d step %d", seed, step)
			}
			triples++
			for f := range e.nodes[id].Fields {
				for _, ancestor := range e.fieldHistory(id, f) {
					edges++
					if !e.causal(ancestor, id) {
						t.Fatal("field ancestry without causal ancestry")
					}
				}
			}
		}
	}
	t.Logf("JOIN triples=%d laws=%d typed-ancestry-checks=%d", triples, triples*3, edges)
}
