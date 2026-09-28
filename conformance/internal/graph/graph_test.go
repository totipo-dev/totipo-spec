package graph

import (
	"fmt"
	"reflect"
	"testing"
)

func TestEdgeResolutionWithinSnapshot(t *testing.T) {
	s := New()
	a := Node{ID: "a", Version: 1, Type: "TOKEN", Identity: "T", Class: "SUPPORTED_VALID", Parents: []string{"b", "c", "d"}, Digest: "A"}
	if e := s.Learn(a, &Value{}); e != nil {
		t.Fatal(e)
	}
	if s.Edge("a", "b") != "UNRESOLVED" {
		t.Fatal("missing")
	}
	for _, n := range []Node{{ID: "b", Type: "DEVICE", Identity: "T", Class: "OPAQUE_ROUTABLE", Digest: "B"}, {ID: "c", Type: "TOKEN", Identity: "X", Class: "SUPPORTED_VALID", Digest: "C"}, {ID: "d", Type: "TOKEN", Identity: "T", Class: "OPAQUE_ROUTABLE", Digest: "D"}} {
		if e := s.Learn(n, &Value{}); e != nil {
			t.Fatal(e)
		}
	}
	if s.Edge("a", "b") != "REJECTED" || s.Edge("a", "c") != "REJECTED" || s.Edge("a", "d") != "RESOLVED" {
		t.Fatal("edge classification")
	}
	s.Disappear("d")
	if s.Edge("a", "d") != "UNRESOLVED" {
		t.Fatal("absent parent resolved")
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
	if e := s.Learn(a, &Value{}); e != nil {
		t.Fatal(e)
	}
	if e := s.Learn(b, &Value{}); e != ErrIntegrity || !s.IntegrityFailure {
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
			if e := a.Learn(nodes[i], &Value{}); e != nil {
				t.Fatal(e)
			}
			if e := b.Learn(nodes[len(nodes)-1-i], &Value{}); e != nil {
				t.Fatal(e)
			}
		}
		if !reflect.DeepEqual(a.Heads("TOKEN", "T"), b.Heads("TOKEN", "T")) {
			t.Fatal("arrival changed heads")
		}

		for id := range a.Nodes {
			a.Disappear(id)
		}
		if len(a.Heads("TOKEN", "T")) != 0 {
			t.Fatal("absent objects remained current")
		}
	})
}

func TestSnapshotIsolationAndAdvisorySeparation(t *testing.T) {
	s := New()
	v := Value{Status: "LIVE", Account: "old"}
	a := Node{ID: "a", Type: "TOKEN", Identity: "T", Class: "SUPPORTED_VALID", Digest: "a"}
	if err := s.Learn(a, &v); err != nil {
		t.Fatal(err)
	}
	snapshot := s.Snapshot()
	s.Remember()
	s.Disappear("a")
	if len(s.Nodes) != 0 || len(snapshot.Nodes) != 1 {
		t.Fatal("history or later observations altered accepted snapshot")
	}
	if got := s.Warnings(); !reflect.DeepEqual(got, []string{"HISTORY_REGRESSION"}) {
		t.Fatal(got)
	}
	s.LoseHistory()
	if len(snapshot.Nodes) != 1 || len(s.Nodes) != 0 {
		t.Fatal("cache loss changed accepted evidence")
	}
}

func TestConfirmationBindsDesiredValueAndIntent(t *testing.T) {
	for _, change := range []string{"value", "intent"} {
		s := New()
		for _, id := range []string{"a", "b"} {
			v := Value{Status: "LIVE", Account: id}
			if err := s.Learn(Node{ID: id, Type: "TOKEN", Identity: "T", Class: "SUPPORTED_VALID", Digest: id}, &v); err != nil {
				t.Fatal(err)
			}
		}
		p, err := s.Plan(Node{ID: "c", Type: "TOKEN", Identity: "T", Class: "SUPPORTED_VALID", Digest: "c"}, Value{Status: "LIVE", Account: "chosen"}, "resolve", true)
		if err != nil {
			t.Fatal(err)
		}
		if change == "value" {
			p.Value.Account = "different"
		} else {
			p.Intent = "restore"
		}
		if p.Publish(s, true) {
			t.Fatal("changed confirmed decision published")
		}
	}
}

func TestNoHistoryFeatureIsNotMemoryLoss(t *testing.T) {
	s := New()
	s.LoseHistory()
	if len(s.Warnings()) != 0 {
		t.Fatal("absence of feature synthesized history warning")
	}
	s.Remember()
	s.LoseHistory()
	if got := s.Warnings(); !reflect.DeepEqual(got, []string{"HISTORY_MEMORY_LOST"}) {
		t.Fatal(got)
	}
	s.ClearHistory()
	if len(s.Warnings()) != 0 {
		t.Fatal("deliberate clearing retained warning state")
	}
}
