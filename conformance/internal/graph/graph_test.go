package graph

import (
	"bytes"
	"fmt"
	"reflect"
	"testing"
	"totipo/conformance/internal/object"
)

func TestCyclesAndMissingEdges(t *testing.T) {
	s := New()
	s.Add(Node{ID: "a", Identity: "t", Parents: []string{"b"}, Canonical: "a", Value: "X"})
	s.Add(Node{ID: "b", Identity: "t", Parents: []string{"a"}, Canonical: "b", Value: "Y"})
	r := s.Evaluate("t")
	if !r.Conflicting || !reflect.DeepEqual(r.Heads, []string{"a", "b"}) {
		t.Fatal(r)
	}
	s.Add(Node{ID: "d", Identity: "t", Parents: []string{"a"}, Canonical: "d", Value: "Z"})
	if r = s.Evaluate("t"); !reflect.DeepEqual(r.Heads, []string{"d"}) {
		t.Fatal(r)
	}
	s.Remove("a")
	r = s.Evaluate("t")
	if !reflect.DeepEqual(r.Heads, []string{"b", "d"}) || !reflect.DeepEqual(r.Unresolved, []string{"a"}) {
		t.Fatal(r)
	}
}
func TestSameIDLocalFailure(t *testing.T) {
	s := New()
	s.Add(Node{ID: "a", Identity: "t", Parents: nil, Canonical: "P", Value: "X"})
	if e := s.Add(Node{ID: "a", Identity: "t", Parents: nil, Canonical: "Q", Value: "Y"}); e == nil {
		t.Fatal("collision chosen")
	}
	s.Add(Node{ID: "b", Identity: "other", Parents: nil, Canonical: "B", Value: "Z"})
	if len(s.Evaluate("t").Heads) != 0 || len(s.Evaluate("other").Heads) != 1 {
		t.Fatal("wrong failure scope")
	}
}
func TestFoldFailureAndCoverage(t *testing.T) {
	ids := [][]byte{}
	for i := 0; i < 10; i++ {
		ids = append(ids, bytes.Repeat([]byte{byte(i)}, 32))
	}
	calls := 0
	stages, e := Fold(ids, func(p [][]byte) ([]byte, error) {
		calls++
		if len(p) != 4 {
			t.Fatal("capacity")
		}
		if calls == 3 {
			return nil, fmt.Errorf("persistence failure")
		}
		return bytes.Repeat([]byte{byte(100 + calls)}, 32), nil
	})
	if e == nil || len(stages) != 2 {
		t.Fatal("lost stages")
	}
}

// Arrival order and disappearance cannot change interpretation of the same
// observed set. Include arbitrary self-edges and cycles, without crypto claims.
func FuzzArrivalAndDisappearance(f *testing.F) {
	f.Add([]byte{1, 0, 2, 1, 0})
	f.Fuzz(func(t *testing.T, b []byte) {
		if len(b) > 64 {
			b = b[:64]
		}
		a, c := New(), New()
		nodes := []Node{}
		for i, v := range b {
			n := Node{ID: fmt.Sprint(i), Identity: "t", Parents: []string{fmt.Sprint(int(v) % 65)}, Canonical: fmt.Sprint(i), Value: fmt.Sprint(v % 3)}
			nodes = append(nodes, n)
			a.Add(n)
		}
		for i := len(nodes) - 1; i >= 0; i-- {
			c.Add(nodes[i])
		}
		if !reflect.DeepEqual(a.Evaluate("t"), c.Evaluate("t")) {
			t.Fatal("arrival order")
		}
		for i := 0; i < len(nodes); i += 2 {
			a.Remove(nodes[i].ID)
			c.Remove(nodes[i].ID)
		}
		if !reflect.DeepEqual(a.Evaluate("t"), c.Evaluate("t")) {
			t.Fatal("disappearance")
		}
	})
}

func TestStronglyConnectedGroupsAgainstIndependentClosure(t *testing.T) {
	// Exhaust all 512 directed three-node graphs, including self edges. Compute
	// reachability by matrix closure as an independent oracle for the traversal.
	for mask := 0; mask < 512; mask++ {
		s := New()
		var reach [3][3]bool
		for i := 0; i < 3; i++ {
			parents := []string{}
			reach[i][i] = true
			for j := 0; j < 3; j++ {
				if mask&(1<<(3*i+j)) != 0 {
					parents = append(parents, fmt.Sprint(j))
					reach[i][j] = true
				}
			}
			s.Add(Node{ID: fmt.Sprint(i), Identity: "T", Parents: parents, Canonical: fmt.Sprint(i), Value: "X"})
		}
		for k := 0; k < 3; k++ {
			for i := 0; i < 3; i++ {
				for j := 0; j < 3; j++ {
					reach[i][j] = reach[i][j] || (reach[i][k] && reach[k][j])
				}
			}
		}
		want := []string{}
		for i := 0; i < 3; i++ {
			head := true
			for j := 0; j < 3; j++ {
				if reach[j][i] && !reach[i][j] {
					head = false
				}
			}
			if head {
				want = append(want, fmt.Sprint(i))
			}
		}
		if got := s.Evaluate("T"); !reflect.DeepEqual(got.Heads, want) || got.Conflicting {
			t.Fatalf("graph %d: %+v want %v", mask, got, want)
		}
	}
}

func TestMetadataSnapshotsAndSCC(t *testing.T) {
	empty := ""
	zero := uint64(0)
	a := Node{ID: "A", Identity: "T", Parents: []string{"B"}, Canonical: "A", Value: "X", ClientName: &empty, ClientTime: &zero}
	b := Node{ID: "B", Identity: "T", Parents: []string{"A"}, Canonical: "B", Value: "X"}
	s := New()
	s.Add(a)
	s.Add(b)
	// Caller mutation cannot change retained validated object facts.
	empty = "changed"
	zero = 99
	a.Parents[0] = "missing"
	r := s.Evaluate("T")
	if r.Conflicting || len(r.Values) != 1 || len(r.HeadObjects) != 2 {
		t.Fatal(r)
	}
	if r.HeadObjects[0].ClientName == nil || *r.HeadObjects[0].ClientName != "" || r.HeadObjects[0].ClientTime == nil || *r.HeadObjects[0].ClientTime != 0 {
		t.Fatal("present-empty/zero lost")
	}
	if r.HeadObjects[1].ClientName != nil || r.HeadObjects[1].ClientTime != nil {
		t.Fatal("absence synthesized")
	}
	*r.HeadObjects[0].ClientName = "mutated result"
	*r.HeadObjects[0].ClientTime = 123
	retained, _ := s.Node("A")
	if *retained.ClientName != "" || *retained.ClientTime != 0 {
		t.Fatal("result aliases store")
	}
	s.Add(Node{ID: "D", Identity: "T", Parents: []string{"A"}, Canonical: "D", Value: "Y"})
	if r = s.Evaluate("T"); !reflect.DeepEqual(r.Heads, []string{"D"}) {
		t.Fatal(r)
	}
	historical, _ := s.Node("A")
	if *historical.ClientName != "" || *historical.ClientTime != 0 {
		t.Fatal("historical metadata lost")
	}
}

func TestFoldOperationIsSnapshot(t *testing.T) {
	name := "operation"
	time := uint64(0)
	op := object.Object{Identity: make([]byte, 32), Status: 1, Algorithm: 1, Digits: 6, Period: 30, Secret: []byte{1}, ClientName: &name, ClientTime: &time}
	frontier := [][]byte{}
	for i := 0; i < 10; i++ {
		frontier = append(frontier, bytes.Repeat([]byte{byte(i)}, 32))
	}
	calls := 0
	_, e := FoldObjects(frontier, op, func(stage object.Object) ([]byte, error) {
		if stage.ClientName == nil || *stage.ClientName != "operation" || stage.ClientTime == nil || *stage.ClientTime != 0 || stage.Secret[0] != 1 {
			t.Fatal("operation changed between stages")
		}
		// Mutate both the caller's input and this callback's copy. Neither may alter
		// the operation snapshot used to construct the following stage.
		name = "changed"
		time = 999
		op.Secret[0] = 2
		*stage.ClientName = "callback"
		*stage.ClientTime = 3
		stage.Secret[0] = 4
		calls++
		return bytes.Repeat([]byte{byte(100 + calls)}, 32), nil
	})
	if e != nil || calls != 3 {
		t.Fatal(calls, e)
	}
}
