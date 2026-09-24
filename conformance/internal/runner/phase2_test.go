package runner

import (
	"crypto/ed25519"
	"fmt"
	"math/rand"
	"reflect"
	"testing"
	"totipo/conformance/internal/corpus"
	"totipo/conformance/internal/model"
	"totipo/conformance/internal/referenceoracle"
)

func TestEd25519StandardComparison(t *testing.T) {
	cs, e := corpus.Load("../../../vectors/v0")
	if e != nil {
		t.Fatal(e)
	}
	positive, negative, different := 0, 0, 0
	for _, c := range cs {
		if c.Kind != "ed25519" {
			continue
		}
		a, m, s := decoded(c.Input, "public_key_hex"), decoded(c.Input, "message_hex"), decoded(c.Input, "signature_hex")
		standard := len(a) == 32 && ed25519.Verify(a, m, s)
		strict := c.Expected["disposition"] == "ACCEPT"
		if strict {
			positive++
		} else {
			negative++
		}
		if standard != strict {
			different++
			t.Logf("DIFFERENCE %s: standard=%v strict=%v (%s)", c.ID, standard, strict, c.Expected["reason"])
		}
	}
	t.Logf("positive=%d negative=%d differing=%d", positive, negative, different)
	if positive < 4 || negative < 20 || different < 1 {
		t.Fatal("corpus does not exercise the intended profile distinctions")
	}
}
func TestBrokenOracleVariants(t *testing.T) {
	cs, e := corpus.Load("../../../vectors/v0")
	if e != nil {
		t.Fatal(e)
	}
	for _, variant := range []string{"current-heads-only", "reversed-coverage", "no-inherited-coverage"} {
		t.Run(variant, func(t *testing.T) {
			distinguished := false
			for _, c := range cs {
				if c.Kind != "model" {
					continue
				}
				us := []model.Update{}
				for _, u := range c.Scenario.Updates {
					us = append(us, model.Update{ID: u.ID, Token: u.Token, Parents: u.Parents, Fields: u.Fields})
				}
				correct, e := referenceoracle.Evaluate(us)
				if e != nil {
					t.Fatal(e)
				}
				broken, e := referenceoracle.EvaluateVariant(us, variant)
				if e != nil {
					t.Fatal(e)
				}
				if !reflect.DeepEqual(correct, broken) {
					t.Log("distinguished by", c.ID)
					distinguished = true
					break
				}
			}
			if !distinguished {
				t.Fatal("no counterexample to mutant")
			}
		})
	}
}
func TestDeterministicDifferentialHistories(t *testing.T) {
	// These generated histories are non-normative. Neither implementation's
	// outputs are written to vectors. Both accepted and invalid candidates are compared.
	const histories = 24
	const steps = 24
	checks, accepted, rejected := 0, 0, 0
	for seed := int64(0); seed < histories; seed++ {
		rng := rand.New(rand.NewSource(0x544f5449504f + seed))
		us := []model.Update{{ID: "root", Token: "t", Parents: []string{}, Fields: map[string]string{"STATUS": "LIVE", "ISSUER": "issuer", "ACCOUNT": "account", "CREDENTIAL": "key0"}}}
		for step := 0; step < steps; step++ {
			view, e := model.Evaluate(us)
			if e != nil {
				t.Fatal(e)
			}
			parents := append([]string{}, view.Tokens["t"].Heads...)
			validIDs := []string{}
			for _, u := range us {
				if view.Validation[u.ID] == "FULLY_VALID" {
					validIDs = append(validIDs, u.ID)
				}
			}
			if rng.Intn(3) != 0 {
				parents = []string{validIDs[rng.Intn(len(validIDs))]}
			}
			f := map[string]string{}
			switch rng.Intn(7) {
			case 0:
				f["STATUS"] = "LIVE"
			case 1:
				f["STATUS"] = "TOMBSTONE"
			case 2:
				f["CREDENTIAL"] = fmt.Sprintf("key%d", step)
			case 3:
				f["ISSUER"] = fmt.Sprintf("issuer%d", step%3)
			case 4:
				f["ACCOUNT"] = fmt.Sprintf("account%d", step%2)
			case 5:
				f["STATUS"] = "LIVE"
				f["CREDENTIAL"] = "combined"
			case 6:
				f["STATUS"] = "LIVE"
				f["ISSUER"] = "mixed"
			}
			id := fmt.Sprintf("s%02d", step)
			candidate := append(append([]model.Update{}, us...), model.Update{ID: id, Token: "t", Parents: parents, Fields: f})
			a, e := model.Evaluate(candidate)
			if e != nil {
				t.Fatal(e)
			}
			b, e := referenceoracle.Evaluate(candidate)
			if e != nil {
				t.Fatal(e)
			}
			checks++
			if !reflect.DeepEqual(a, b) {
				t.Fatalf("seed %d step %d\nprimary %+v\noracle %+v", seed, step, a, b)
			}
			if a.Validation[id] == "FULLY_VALID" {
				accepted++
				us = candidate
			} else {
				rejected++
				for _, r := range a.Resolutions {
					if r == id {
						t.Fatal("invalid candidate leaked coverage")
					}
				}
			}
			// Hide/reveal the candidate as a view change; accepted history must remain
			// unchanged when the same immutable set is restored in a different arrival order.
			if len(us) > 2 {
				hidden := us[:len(us)-1]
				hx, e := model.Evaluate(hidden)
				if e != nil {
					t.Fatal(e)
				}
				hy, e := referenceoracle.Evaluate(hidden)
				if e != nil || !reflect.DeepEqual(hx, hy) {
					t.Fatal("hidden-view divergence")
				}
				checks++
				shuffled := append([]model.Update{}, us...)
				rng.Shuffle(len(shuffled), func(i, j int) { shuffled[i], shuffled[j] = shuffled[j], shuffled[i] })
				x, e := model.Evaluate(shuffled)
				if e != nil {
					t.Fatal(e)
				}
				y, e := referenceoracle.Evaluate(us)
				if e != nil || !reflect.DeepEqual(x, y) {
					t.Fatal("arrival-order divergence")
				}
				checks++
			}
		}
	}
	t.Logf("histories=%d candidate-checks=%d accepted=%d rejected=%d total-comparisons=%d", histories, histories*steps, accepted, rejected, checks)
	if accepted == 0 || rejected == 0 {
		t.Fatal("generator missed validity boundary")
	}
}

func TestCapacityPreservesSemanticFrontier(t *testing.T) {
	root := model.Update{ID: "root", Token: "t", Parents: []string{}, Fields: map[string]string{"STATUS": "LIVE", "ISSUER": "i", "ACCOUNT": "a", "CREDENTIAL": "k"}}
	us := []model.Update{root}
	heads := []string{}
	for n := 0; n < 33; n++ {
		id := fmt.Sprintf("h%02d", n)
		heads = append(heads, id)
		us = append(us, model.Update{ID: id, Token: "t", Parents: []string{"root"}, Fields: map[string]string{"CREDENTIAL": id}})
		if n == 31 {
			v, e := model.Evaluate(append(append([]model.Update{}, us...), model.Update{ID: "within", Token: "t", Parents: append([]string{}, heads...), Fields: map[string]string{"ISSUER": "edited"}}))
			if e != nil || v.Validation["within"] != "FULLY_VALID" {
				t.Fatal("32 parents rejected")
			}
		}
	}
	candidate := model.Update{ID: "oversize", Token: "t", Parents: heads, Fields: map[string]string{"ISSUER": "edited"}}
	us = append(us, candidate)
	a, e := model.Evaluate(us)
	if e != nil {
		t.Fatal(e)
	}
	b, e := referenceoracle.Evaluate(us)
	if e != nil || !reflect.DeepEqual(a, b) || a.Validation["oversize"] != "INVALID" || len(a.Tokens["t"].Heads) != 33 || len(a.Tokens["t"].Fields["CREDENTIAL"]) != 33 {
		t.Fatal("capacity boundary truncated history or permitted unencodable update")
	}
}
