// Package graph evaluates valid observed TOKEN facts. Abstract IDs in semantic
// tests do not claim to be constructible cryptographic objects.
package graph

import (
	"bytes"
	"encoding/hex"
	"errors"
	"sort"
	"totipo/conformance/internal/object"
)

type Node struct {
	ID       string   `json:"id"`
	Identity string   `json:"identity"`
	Parents  []string `json:"parents"`
	// Canonical identifies exact canonical plaintext in defensive model fixtures.
	Canonical  string  `json:"canonical"`
	Value      string  `json:"value"`
	ClientName *string `json:"client_name,omitempty"`
	ClientTime *uint64 `json:"client_time,omitempty"`
}
type Result struct {
	Heads       []string `json:"heads"`
	HeadObjects []Node   `json:"head_objects"`
	Values      []string `json:"values"`
	Unresolved  []string `json:"unresolved"`
	Conflicting bool     `json:"conflicting"`
}
type Store struct {
	nodes  map[string]Node
	failed map[string]bool
}

func New() *Store { return &Store{map[string]Node{}, map[string]bool{}} }
func (s *Store) Add(n Node) error {
	if s.failed[n.ID] {
		return errors.New("failed object identity")
	}
	if old, ok := s.nodes[n.ID]; ok && old.Canonical != n.Canonical {
		delete(s.nodes, n.ID)
		s.failed[n.ID] = true
		return errors.New("same-ID integrity failure")
	}
	s.nodes[n.ID] = cloneNode(n)
	return nil
}

// Node returns a detached full record for current or historical presentation.
func (s *Store) Node(id string) (Node, bool) {
	n, ok := s.nodes[id]
	return cloneNode(n), ok
}
func cloneNode(n Node) Node {
	if n.Parents != nil {
		n.Parents = append([]string{}, n.Parents...)
	}
	if n.ClientName != nil {
		name := *n.ClientName
		n.ClientName = &name
	}
	if n.ClientTime != nil {
		time := *n.ClientTime
		n.ClientTime = &time
	}
	return n
}

// FromObject projects an already validated object without losing its metadata.
// Symbolic graph tests may construct Node directly instead of using wire IDs.
func FromObject(id string, o object.Object) (Node, error) {
	p, e := o.Encode()
	if e != nil {
		return Node{}, e
	}
	parents := []string{}
	for _, p := range o.Parents {
		parents = append(parents, hex.EncodeToString(p))
	}
	return cloneNode(Node{ID: id, Identity: hex.EncodeToString(o.Identity), Parents: parents,
		Canonical: hex.EncodeToString(p), Value: hex.EncodeToString(o.ValueBytes()),
		ClientName: o.ClientName, ClientTime: o.ClientTime}), nil
}
func (s *Store) Remove(id string) { delete(s.nodes, id) }
func sorted(m map[string]bool) []string {
	out := []string{}
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}
func (s *Store) Evaluate(identity string) Result {
	nodes := map[string]Node{}
	for id, n := range s.nodes {
		if n.Identity == identity {
			nodes[id] = n
		}
	}
	reach := map[string]map[string]bool{}
	unresolved := map[string]bool{}
	for id := range nodes {
		seen := map[string]bool{}
		todo := []string{id}
		for len(todo) > 0 {
			x := todo[len(todo)-1]
			todo = todo[:len(todo)-1]
			if seen[x] {
				continue
			}
			seen[x] = true
			for _, p := range nodes[x].Parents {
				if _, ok := nodes[p]; ok {
					todo = append(todo, p)
				} else {
					unresolved[p] = true
				}
			}
		}
		reach[id] = seen
	}
	heads := map[string]bool{}
	values := map[string]bool{}
	for id, n := range nodes {
		head := true
		for other := range nodes {
			if reach[other][id] && !reach[id][other] {
				head = false
				break
			}
		}
		if head {
			heads[id] = true
			values[n.Value] = true
		}
	}
	ids := sorted(heads)
	records := make([]Node, 0, len(ids))
	for _, id := range ids {
		records = append(records, cloneNode(nodes[id]))
	}
	return Result{Heads: ids, HeadObjects: records, Values: sorted(values), Unresolved: sorted(unresolved), Conflicting: len(values) > 1}
}

// Fold stages sorted original IDs with capacity four. The callback constructs
// and publishes a complete object with the operation's desired value and exact
// client metadata unchanged in every stage, returning
// its ID. Earlier published stages survive any later callback error.
func Fold(frontier [][]byte, publish func([][]byte) ([]byte, error)) ([][][]byte, error) {
	ids := make([][]byte, len(frontier))
	for i, id := range frontier {
		if len(id) != 32 {
			return nil, errors.New("ID width")
		}
		ids[i] = bytes.Clone(id)
	}
	sort.Slice(ids, func(i, j int) bool { return bytes.Compare(ids[i], ids[j]) < 0 })
	for i := 1; i < len(ids); i++ {
		if bytes.Equal(ids[i-1], ids[i]) {
			return nil, errors.New("duplicate frontier")
		}
	}
	stages := [][][]byte{}
	var carry []byte
	for len(ids) > 0 || len(stages) == 0 {
		n := 4
		if carry != nil {
			n = 3
		}
		if n > len(ids) {
			n = len(ids)
		}
		parents := append([][]byte{}, ids[:n]...)
		ids = ids[n:]
		if carry != nil {
			parents = append(parents, carry)
		}
		sort.Slice(parents, func(i, j int) bool { return bytes.Compare(parents[i], parents[j]) < 0 })
		next, e := publish(parents)
		if e != nil {
			return stages, e
		}
		if len(next) != 32 {
			return stages, errors.New("published ID width")
		}
		stages = append(stages, parents)
		carry = bytes.Clone(next)
		if len(ids) == 0 {
			break
		}
	}
	return stages, nil
}

// FoldObjects snapshots one complete operation. Only parent lists change between
// stages. Each callback receives detached fields so it cannot mutate later stages.
func FoldObjects(frontier [][]byte, operation object.Object, publish func(object.Object) ([]byte, error)) ([][][]byte, error) {
	operation.Parents = nil
	template, e := operation.Encode()
	if e != nil {
		return nil, e
	}
	return Fold(frontier, func(parents [][]byte) ([]byte, error) {
		_, stage := object.Dispatch(template)
		stage.Parents = make([][]byte, len(parents))
		for i, p := range parents {
			stage.Parents[i] = bytes.Clone(p)
		}
		return publish(*stage)
	})
}
