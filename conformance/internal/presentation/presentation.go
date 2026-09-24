// Package presentation models post-authentication DEVICE_UPDATE history only.
package presentation

import (
	"fmt"
	"sort"
)

type Update struct {
	ID      string   `json:"id"`
	Signer  string   `json:"signer"`
	Parents []string `json:"parents"`
	Name    string   `json:"name"`
	Kind    string   `json:"kind"`
}
type View struct {
	Validation map[string]string   `json:"validation"`
	Heads      map[string][]string `json:"heads"`
	Names      map[string][]string `json:"names"`
}

func Evaluate(updates []Update) (View, error) {
	objects := map[string]Update{}
	for _, u := range updates {
		if u.ID == "" || u.Signer == "" {
			return View{}, fmt.Errorf("missing identity")
		}
		if _, ok := objects[u.ID]; ok {
			return View{}, fmt.Errorf("duplicate identity")
		}
		objects[u.ID] = u
	}
	out := View{map[string]string{}, map[string][]string{}, map[string][]string{}}
	active := map[string]bool{}
	var before func(string, string) bool
	before = func(a, b string) bool {
		seen := map[string]bool{}
		todo := append([]string{}, objects[b].Parents...)
		for len(todo) > 0 {
			x := todo[0]
			todo = todo[1:]
			if x == a {
				return true
			}
			if seen[x] {
				continue
			}
			seen[x] = true
			todo = append(todo, objects[x].Parents...)
		}
		return false
	}
	var check func(string) string
	check = func(id string) string {
		if s := out.Validation[id]; s != "" {
			return s
		}
		u, ok := objects[id]
		if !ok {
			return "PENDING"
		}
		if active[id] {
			return "INVALID"
		}
		active[id] = true
		defer delete(active, id)
		bad := u.Kind != "DEVICE_UPDATE" || len(u.Parents) > 32
		pending := false
		seen := map[string]bool{}
		for _, p := range u.Parents {
			bad = bad || seen[p] || p == id
			seen[p] = true
			s := check(p)
			bad = bad || s == "INVALID"
			pending = pending || s == "PENDING"
			if v, ok := objects[p]; ok && (v.Signer != u.Signer || v.Kind != "DEVICE_UPDATE") {
				bad = true
			}
		}
		s := "FULLY_VALID"
		if bad {
			s = "INVALID"
		} else if pending {
			s = "PENDING"
		} else {
			for _, a := range u.Parents {
				for _, b := range u.Parents {
					if a != b && before(a, b) {
						s = "INVALID"
					}
				}
			}
		}
		out.Validation[id] = s
		return s
	}
	ids := []string{}
	for id := range objects {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	for _, id := range ids {
		check(id)
	}
	for _, id := range ids {
		if out.Validation[id] != "FULLY_VALID" {
			continue
		}
		u := objects[id]
		maximal := true
		for _, other := range ids {
			if out.Validation[other] == "FULLY_VALID" && objects[other].Signer == u.Signer && before(id, other) {
				maximal = false
				break
			}
		}
		if maximal {
			out.Heads[u.Signer] = append(out.Heads[u.Signer], id)
		}
	}
	for signer, heads := range out.Heads {
		values := map[string]bool{}
		for _, id := range heads {
			values[objects[id].Name] = true
		}
		names := []string{}
		for name := range values {
			names = append(names, name)
		}
		sort.Strings(names)
		out.Names[signer] = names
	}
	return out, nil
}
