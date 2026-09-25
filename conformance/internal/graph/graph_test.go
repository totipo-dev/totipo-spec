package graph

import (
	"fmt"
	"reflect"
	"testing"
)

func TestEdgeResolutionAndDurability(t *testing.T) {
	s := New()
	a := Node{ID: "a", Version: 1, Type: "TOKEN", Identity: "T", Class: "SUPPORTED_VALID", Parents: []string{"b", "c", "d"}, Digest: "A"}
	if e := s.Learn(a, nil, true); e != nil {
		t.Fatal(e)
	}
	if s.Edge("a", "b") != "UNRESOLVED" {
		t.Fatal("missing")
	}
	for _, n := range []Node{{ID: "b", Type: "DEVICE", Identity: "T", Class: "OPAQUE_ROUTABLE", Digest: "B"}, {ID: "c", Type: "TOKEN", Identity: "X", Class: "SUPPORTED_VALID", Digest: "C"}, {ID: "d", Type: "TOKEN", Identity: "T", Class: "OPAQUE_ROUTABLE", Digest: "D"}} {
		if e := s.Learn(n, nil, true); e != nil {
			t.Fatal(e)
		}
	}
	if s.Edge("a", "b") != "REJECTED" || s.Edge("a", "c") != "REJECTED" || s.Edge("a", "d") != "RESOLVED" {
		t.Fatal("edge classification")
	}
	s.Disappear("d")
	if s.Edge("a", "d") != "RESOLVED" {
		t.Fatal("durable edge lost")
	}
	a.Parents[0] = "mutation"
	if s.Nodes["a"].Parents[0] != "b" {
		t.Fatal("aliased input")
	}
}
func TestDeviceCycle(t *testing.T) {
	s := New()
	a := Node{ID: "a", Type: "DEVICE", Identity: "D", Class: "OPAQUE_ROUTABLE", Parents: []string{"b"}}
	b := Node{ID: "b", Type: "DEVICE", Identity: "D", Class: "SUPPORTED_VALID", Parents: []string{"a"}}
	if e := s.Learn(a, nil, true); e != nil {
		t.Fatal(e)
	}
	if e := s.Learn(b, nil, true); e != ErrIntegrity || !s.ContinuityUnknown {
		t.Fatal("cycle did not block")
	}
}

func FuzzArrivalAndDisappearance(f *testing.F) {
	f.Add([]byte{1, 2, 3, 4})
	f.Add([]byte{255, 0, 99})
	f.Fuzz(func(t *testing.T, data []byte) {
		if len(data) > 24 {
			data = data[:24]
		}
		if len(data) == 0 {
			return
		}
		nodes := make([]Node, len(data))
		for i, b := range data {
			n := Node{ID: fmt.Sprint(i), Version: 1, Type: "TOKEN", Identity: "T", Class: "SUPPORTED_VALID", Digest: fmt.Sprint(i)}
			if i > 0 {
				n.Parents = []string{fmt.Sprint(int(b) % i)}
			}
			nodes[i] = n
		}
		a, b := New(), New()
		for i := range nodes {
			if e := a.Learn(nodes[i], nil, true); e != nil {
				t.Fatal(e)
			}
			if e := b.Learn(nodes[len(nodes)-1-i], nil, true); e != nil {
				t.Fatal(e)
			}
		}
		if !reflect.DeepEqual(a.Heads("TOKEN", "T"), b.Heads("TOKEN", "T")) {
			t.Fatal("arrival changed heads")
		}
		before := a.Heads("TOKEN", "T")
		for id := range a.Nodes {
			a.Disappear(id)
		}
		if !reflect.DeepEqual(before, a.Heads("TOKEN", "T")) {
			t.Fatal("disappearance erased topology")
		}
	})
}
