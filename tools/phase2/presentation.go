package main

import (
	"fmt"
	"sort"
)

type D struct {
	ID      string   `json:"id"`
	Signer  string   `json:"signer"`
	Parents []string `json:"parents"`
	Name    string   `json:"name"`
	Kind    string   `json:"kind"`
}
type DV struct {
	Validation M                   `json:"validation"`
	Heads      map[string][]string `json:"heads"`
	Names      map[string][]string `json:"names"`
}
type DS struct {
	Updates  []D `json:"updates"`
	Expected DV  `json:"expected"`
}
type DC struct {
	ID           string `json:"id"`
	Kind         string `json:"kind"`
	Description  string `json:"description"`
	Provenance   P      `json:"provenance"`
	Input        M      `json:"input"`
	Expected     M      `json:"expected"`
	Presentation DS     `json:"presentation"`
}

func deviceFixtures() {
	var cases []DC
	add := func(id string, us []D, heads, names map[string][]string, invalid M) {
		validation := M{}
		for _, u := range us {
			validation[u.ID] = "FULLY_VALID"
		}
		for k, v := range invalid {
			validation[k] = v
		}
		cases = append(cases, DC{"v0/presentation/" + id, "presentation", id, P{"review/phase2/REVIEW.md", "spec-derived-reviewed", "r36 21,49; " + id}, M{}, M{"disposition": "VERIFIED"}, DS{us, DV{validation, heads, names}}})
	}
	u := func(id, key, name string, parents ...string) D {
		if parents == nil {
			parents = []string{}
		}
		return D{id, key, parents, name, "DEVICE_UPDATE"}
	}
	r := u("r", "k", "root")
	a := u("a", "k", "alpha", "r")
	b := u("b", "k", "beta", "r")
	m := u("m", "k", "merged", "a", "b")
	add("root", []D{r}, map[string][]string{"k": {"r"}}, map[string][]string{"k": {"root"}}, nil)
	add("rename", []D{r, a}, map[string][]string{"k": {"a"}}, map[string][]string{"k": {"alpha"}}, nil)
	add("concurrent-names", []D{r, a, b}, map[string][]string{"k": {"a", "b"}}, map[string][]string{"k": {"alpha", "beta"}}, nil)
	add("rename-all-heads", []D{r, a, b, m}, map[string][]string{"k": {"m"}}, map[string][]string{"k": {"merged"}}, nil)
	equal := b
	equal.Name = "alpha"
	add("equal-values-distinct-heads", []D{r, a, equal}, map[string][]string{"k": {"a", "b"}}, map[string][]string{"k": {"alpha"}}, nil)
	p := u("p", "k", "pending", "missing")
	add("pending", []D{r, p}, map[string][]string{"k": {"r"}}, map[string][]string{"k": {"root"}}, M{"p": "PENDING"})
	other := u("other", "other-key", "other")
	cross := u("cross", "k", "bad", "other")
	add("cross-signer-invalid", []D{r, other, cross}, map[string][]string{"k": {"r"}, "other-key": {"other"}}, map[string][]string{"k": {"root"}, "other-key": {"other"}}, M{"cross": "INVALID"})
	redundant := u("bad", "k", "bad", "r", "a")
	add("redundant-parent-invalid", []D{r, a, redundant}, map[string][]string{"k": {"a"}}, map[string][]string{"k": {"alpha"}}, M{"bad": "INVALID"})
	duplicate := u("bad", "k", "bad", "r", "r")
	add("duplicate-parent-invalid", []D{r, duplicate}, map[string][]string{"k": {"r"}}, map[string][]string{"k": {"root"}}, M{"bad": "INVALID"})
	token := u("token", "k", "not-presentation")
	token.Kind = "TOKEN_UPDATE"
	cross = u("cross", "k", "bad", "token")
	add("token-parent-invalid", []D{r, token, cross}, map[string][]string{"k": {"r"}}, map[string][]string{"k": {"root"}}, M{"token": "INVALID", "cross": "INVALID"})
	roots := u("second", "k", "second")
	add("independent-roots", []D{r, roots}, map[string][]string{"k": {"r", "second"}}, map[string][]string{"k": {"root", "second"}}, nil)
	sort.Slice(cases, func(i, j int) bool { return cases[i].ID < cases[j].ID })
	write("vectors/v0/presentation/phase2.json", struct {
		Schema int  `json:"schema"`
		Cases  []DC `json:"cases"`
	}{1, cases})
	fmt.Println("authored", len(cases), "presentation cases")
}
