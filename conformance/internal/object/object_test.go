package object

import (
	"bytes"
	"strings"
	"testing"
	"totipo/conformance/internal/tlv"
)

func fixture() Object {
	return Object{Identity: make([]byte, 32), Status: 1, Algorithm: 1, Digits: 6, Period: 30, Secret: []byte{1}}
}
func TestCanonicalBoundaries(t *testing.T) {
	o := fixture()
	name := strings.Repeat("n", 128)
	time := ^uint64(0)
	o.ClientName = &name
	o.ClientTime = &time
	o.Issuer = strings.Repeat("i", 256)
	o.Account = strings.Repeat("a", 256)
	o.Secret = bytes.Repeat([]byte{1}, 128)
	for i := 0; i < 4; i++ {
		o.Parents = append(o.Parents, bytes.Repeat([]byte{byte(i)}, 32))
	}
	p, e := o.Encode()
	if e != nil || len(p) != 1005 {
		t.Fatal(len(p), e)
	}
	o.Parents = append(o.Parents, bytes.Repeat([]byte{4}, 32))
	if _, e = o.Encode(); e == nil {
		t.Fatal("fifth parent")
	}
	o = fixture()
	p, _ = o.Encode()
	for _, bad := range [][]byte{p[:len(p)-1], append(bytes.Clone(p), 0), append(bytes.Clone(p), []byte{0, 13, 0, 0}...)} {
		if s, _ := Dispatch(bad); s != Invalid {
			t.Fatal("bad framing")
		}
	}
	for _, mutate := range []func(*Object){func(o *Object) { o.Period = 0 }, func(o *Object) { o.Digits = 9 }, func(o *Object) { o.Algorithm = 0 }, func(o *Object) { o.Status = 0 }, func(o *Object) { o.Secret = nil }, func(o *Object) { o.Issuer = "\xff" }, func(o *Object) { o.ClientName = new(string); *o.ClientName = strings.Repeat("a", 129) }, func(o *Object) { o.Parents = [][]byte{make([]byte, 32), make([]byte, 32)} }} {
		o = fixture()
		mutate(&o)
		if _, e = o.Encode(); e == nil {
			t.Fatal("invalid field")
		}
	}
}
func TestMetadataAndCompleteValue(t *testing.T) {
	o := fixture()
	before, _ := o.Encode()
	v := o.ValueBytes()
	empty := ""
	zero := uint64(0)
	o.ClientName = &empty
	o.ClientTime = &zero
	after, _ := o.Encode()
	if bytes.Equal(before, after) || !bytes.Equal(v, o.ValueBytes()) {
		t.Fatal("metadata equality")
	}
	o.Status = 2
	if bytes.Equal(v, o.ValueBytes()) {
		t.Fatal("status omitted")
	}
	if _, e := o.Encode(); e != nil {
		t.Fatal("complete tombstone")
	}
	o.Secret = nil
	if _, e := o.Encode(); e == nil {
		t.Fatal("incomplete tombstone")
	}
}
func TestOptionalOrderAndWidths(t *testing.T) {
	o := fixture()
	p, _ := o.Encode()
	for _, tail := range [][]byte{{0, 12, 0, 0}, {0, 11, 0, 1, 255}, {0, 12, 0, 8, 0, 0, 0, 0, 0, 0, 0, 0, 0, 11, 0, 0}} {
		if s, _ := Dispatch(append(bytes.Clone(p), tail...)); s != Invalid {
			t.Fatal("metadata grammar")
		}
	}
	f, _ := tlv.Encode(12, tlv.U64(0))
	if s, _ := Dispatch(append(p, f...)); s != Supported {
		t.Fatal("time without name")
	}
}
func FuzzDispatch(f *testing.F) {
	o := fixture()
	p, _ := o.Encode()
	f.Add(p)
	f.Add([]byte{})
	f.Fuzz(func(t *testing.T, p []byte) {
		s, o := Dispatch(p)
		if s == Supported {
			again, e := o.Encode()
			if e != nil || !bytes.Equal(p, again) {
				t.Fatal("canonical roundtrip")
			}
		} else if o != nil {
			t.Fatal("invalid object leaked")
		}
	})
}
