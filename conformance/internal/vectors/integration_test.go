package vectors

import (
	"bytes"
	"encoding/hex"
	"reflect"
	"testing"
	"totipo/conformance/internal/cryptov1"
	"totipo/conformance/internal/graph"
	"totipo/conformance/internal/object"
	"totipo/conformance/internal/totp"
)

func TestRealFoldAndMetadataEquality(t *testing.T) {
	name, empty, decomposed := "Alice's laptop", "", "e\u0301 laptop"
	time, zero, max := uint64(1770000000), uint64(0), ^uint64(0)
	for _, trial := range []struct {
		label string
		name  *string
		time  *uint64
	}{
		{"present", &name, &time}, {"absent", nil, nil}, {"empty-zero", &empty, &zero},
		{"absent-name-max-time", nil, &max}, {"exact-utf8-absent-time", &decomposed, nil},
	} {
		t.Run(trial.label, func(t *testing.T) {
			k, _ := cryptov1.Derive(bytes.Repeat([]byte{3}, 32))
			operation := object.Object{Identity: make([]byte, 32), Parents: [][]byte{}, Status: 2, Algorithm: 1, Digits: 8, Period: 30, Secret: []byte("12345678901234567890"), ClientName: trial.name, ClientTime: trial.time}
			identity := hex.EncodeToString(operation.Identity)
			s := graph.New()
			frontier := [][]byte{}
			expected := map[string]graph.Node{}
			accept := func(id string, p []byte, parsed object.Object) {
				n, e := graph.FromObject(id, parsed)
				if e != nil {
					t.Fatal(e)
				}
				if n.Canonical != hex.EncodeToString(p) || !reflect.DeepEqual(n.ClientName, parsed.ClientName) || !reflect.DeepEqual(n.ClientTime, parsed.ClientTime) {
					t.Fatal("projection lost object information")
				}
				if e = s.Add(n); e != nil {
					t.Fatal(e)
				}
				expected[id] = n
			}
			for i := 0; i < 10; i++ {
				original := operation
				client := string(rune('A' + i))
				timestamp := uint64(i)
				original.ClientName = &client
				original.ClientTime = &timestamp
				p, e := original.Encode()
				if e != nil {
					t.Fatal(e)
				}
				if !bytes.Equal(operation.ValueBytes(), original.ValueBytes()) {
					t.Fatal("metadata affected value")
				}
				id, b, e := k.Seal(p)
				if e != nil {
					t.Fatal(e)
				}
				opened, e := k.Open(id, b)
				if e != nil {
					t.Fatal(e)
				}
				class, parsed := object.Dispatch(opened)
				if class != object.Supported {
					t.Fatal(class)
				}
				raw, _ := hex.DecodeString(id)
				frontier = append(frontier, raw)
				accept(id, opened, *parsed)
			}
			r := s.Evaluate(identity)
			if r.Conflicting || len(r.Values) != 1 || len(r.Heads) != 10 || len(r.HeadObjects) != 10 {
				t.Fatal(r)
			}
			for _, head := range r.HeadObjects {
				if !reflect.DeepEqual(head, expected[head.ID]) {
					t.Fatal("equal-valued head lost metadata")
				}
			}
			var final string
			stages, e := graph.FoldObjects(frontier, operation, func(stage object.Object) ([]byte, error) {
				p, e := stage.Encode()
				if e != nil {
					return nil, e
				}
				id, b, e := k.Seal(p)
				if e != nil {
					return nil, e
				}
				opened, e := k.Open(id, b)
				if e != nil {
					return nil, e
				}
				_, parsed := object.Dispatch(opened)
				if !bytes.Equal(parsed.ValueBytes(), operation.ValueBytes()) || !reflect.DeepEqual(parsed.ClientName, trial.name) || !reflect.DeepEqual(parsed.ClientTime, trial.time) {
					t.Fatal("fold changed value or metadata presence/value")
				}
				accept(id, opened, *parsed)
				final = id
				return hex.DecodeString(id)
			})
			if e != nil || len(stages) != 3 {
				t.Fatal(stages, e)
			}
			r = s.Evaluate(identity)
			if !reflect.DeepEqual(r.Heads, []string{final}) || r.Conflicting || !reflect.DeepEqual(r.HeadObjects, []graph.Node{expected[final]}) {
				t.Fatal(r)
			}
			// Historical originals and intermediate stages still expose their own exact
			// metadata. No synthetic current metadata overwrites the retained records.
			for id, want := range expected {
				got, ok := s.Node(id)
				if !ok || !reflect.DeepEqual(got, want) {
					t.Fatal("historical metadata changed", id)
				}
			}
			for _, status := range []byte{1, 2} {
				operation.Status = status
				code, e := totp.Code(operation.Algorithm, operation.Secret, operation.Digits, operation.Period, 59)
				if e != nil || code != "94287082" {
					t.Fatal(code, e)
				}
			}
		})
	}
}
