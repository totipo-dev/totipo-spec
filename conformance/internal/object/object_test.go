package object

import (
	"bytes"
	"testing"
	"totipo/conformance/internal/tlv"
)

func sample() Object {
	return Object{Routing: Routing{Version: 1, Type: Token, Identity: make([]byte, 32), Author: make([]byte, 32)}, Status: 1, Algorithm: 1, Digits: 6, Period: 30, Secret: []byte{1}}
}
func change(p []byte, tag uint16, v []byte) []byte {
	out := []byte{}
	for len(p) > 0 {
		f, r, e := tlv.Take(p)
		if e != nil {
			panic(e)
		}
		if f.Tag == tag {
			f.Value = v
		}
		b, e := tlv.Encode(f.Tag, f.Value)
		if e != nil {
			panic(e)
		}
		out = append(out, b...)
		p = r
	}
	return out
}
func TestIntrinsicVersusProvenance(t *testing.T) {
	o := sample()
	for _, sig := range [][]byte{nil, {0xff}, {0x30, 0}, bytes.Repeat([]byte{0xaa}, 72)} {
		o.Signature = sig
		p, e := o.Encode()
		if e != nil {
			t.Fatal(e)
		}
		c, _ := Dispatch(p)
		if c != Supported {
			t.Fatal(c)
		}
	}
	o = Object{Routing: Routing{Version: 1, Type: Device}, PublicKey: make([]byte, 65)}
	o.Identity = DeviceID(o.PublicKey)
	p, e := o.Encode()
	if e != nil {
		t.Fatal(e)
	}
	c, _ := Dispatch(p)
	if c != Supported {
		t.Fatal("off-curve key must affect provenance only")
	}
}
func TestRejectGrammar(t *testing.T) {
	base, _ := sample().Encode()
	for _, v := range []struct {
		name string
		tag  uint16
		b    []byte
	}{
		{"time-width", 6, make([]byte, 7)}, {"identity-width", 0x101, make([]byte, 31)}, {"author-width", 0x102, make([]byte, 33)}, {"status", 0x103, []byte{0}}, {"invalid-utf8", 0x104, []byte{0xff}}, {"issuer-overflow", 0x104, bytes.Repeat([]byte{'x'}, 257)}, {"account-overflow", 0x105, bytes.Repeat([]byte{'x'}, 257)}, {"signature-overflow", 0xff01, make([]byte, 73)}, {"parent-count", 4, tlv.U16(33)},
	} {
		t.Run(v.name, func(t *testing.T) {
			c, _ := Dispatch(change(base, v.tag, v.b))
			if c != Invalid {
				t.Fatal(c)
			}
		})
	}
	o := sample()
	p, _ := o.Encode()
	_, tail, _ := ParseRouting(p)
	_ = tail
	for _, mutate := range []func(*Object){func(o *Object) { o.Algorithm = 4 }, func(o *Object) { o.Digits = 9 }, func(o *Object) { o.Period = 0 }, func(o *Object) { o.Secret = nil }, func(o *Object) { o.Secret = make([]byte, 129) }} {
		o := sample()
		mutate(&o)
		if _, e := o.Encode(); e == nil {
			t.Fatal("invalid credential encoded")
		}
	}
	for _, p := range [][]byte{append(bytes.Clone(base), 0), base[:len(base)-1], append(bytes.Clone(base), base[len(base)-4:]...)} {
		if c, _ := Dispatch(p); c != Invalid {
			t.Fatal("trailing/truncated/duplicate signature accepted")
		}
	}
}
func TestFrozenOpaqueTail(t *testing.T) {
	for _, typ := range []byte{Token, Device} {
		o := sample()
		o.Type = typ
		o.Version = 255
		p, e := o.Routing.Encode()
		if e != nil {
			t.Fatal(e)
		}
		for _, tail := range [][]byte{nil, {0xff}, {0, 0, 0xff, 0xff}, bytes.Repeat([]byte{0xaa}, 200)} {
			c, got := Dispatch(append(bytes.Clone(p), tail...))
			if c != Opaque || got.Version != 255 {
				t.Fatal("future tail interpreted")
			}
		}
	}
}
func TestTimestampAndCapacity(t *testing.T) {
	o := sample()
	o.AuthorTime = ^uint64(0)
	p, e := o.Encode()
	if e != nil {
		t.Fatal(e)
	}
	_, got := Dispatch(p)
	if got.AuthorTime != o.AuthorTime {
		t.Fatal("lost unsigned time")
	}
	d := Object{Routing: Routing{Version: 1, Type: Device}, PublicKey: make([]byte, 65), DisplayName: string(bytes.Repeat([]byte{'D'}, 254))}
	d.Identity = DeviceID(d.PublicKey)
	for i := 0; i < 15; i++ {
		b := make([]byte, 32)
		b[31] = byte(i)
		d.Parents = append(d.Parents, b)
	}
	// 70-byte signature yields 1005 bytes, but writer must reserve 1007.
	d.Signature = make([]byte, 70)
	p, e = d.Encode()
	if e != nil {
		t.Fatal(e)
	}
	n, e := d.ReservedSize()
	if e != nil || len(p) != 1005 || n != 1007 {
		t.Fatalf("actual %d reserved %d err %v", len(p), n, e)
	}
}
func FuzzDispatch(f *testing.F) {
	p, _ := sample().Encode()
	f.Add(p)
	f.Add([]byte{})
	f.Add([]byte{0, 1, 0, 1, 2})
	f.Fuzz(func(t *testing.T, p []byte) {
		c, o := Dispatch(p)
		if c == Supported {
			b, e := o.Encode()
			if e != nil || !bytes.Equal(p, b) {
				t.Fatal("canonical roundtrip failure")
			}
		}
		if c == Opaque {
			r, _, e := ParseRouting(p)
			if e != nil || r.Version == 1 {
				t.Fatal("invalid opaque classification")
			}
		}
	})
}
