package runner

import (
	"bytes"
	"encoding/binary"
	"testing"
	"totipo/conformance/internal/corpus"
	"totipo/conformance/internal/dispatch"
	"totipo/conformance/internal/envelope"
	"totipo/conformance/internal/tlv"
)

func seeds(f *testing.F, kind string) [][]byte {
	f.Helper()
	cases, e := corpus.Load("../../../vectors/v0")
	if e != nil {
		f.Fatal(e)
	}
	out := [][]byte{}
	for _, c := range cases {
		if c.Kind == kind {
			b, e := corpus.Hex(c.Input["bytes_hex"])
			if e != nil {
				f.Fatal(e)
			}
			out = append(out, b)
		}
	}
	return out
}
func FuzzTLVFraming(f *testing.F) {
	for _, b := range seeds(f, "tlv") {
		f.Add(b)
	}
	f.Fuzz(func(t *testing.T, b []byte) {
		if tlv.Framing(b) != nil {
			return
		}
		remaining := b
		var encoded []byte
		for len(remaining) > 0 {
			tag, v, r, e := tlv.Next(remaining)
			if e != nil {
				t.Fatal(e)
			}
			var h [4]byte
			binary.BigEndian.PutUint16(h[:2], tag)
			binary.BigEndian.PutUint16(h[2:], uint16(len(v)))
			encoded = append(encoded, h[:]...)
			encoded = append(encoded, v...)
			remaining = r
		}
		if !bytes.Equal(b, encoded) {
			t.Fatal("accepted framing changed bytes")
		}
	})
}
func FuzzV0Grammar(f *testing.F) {
	for _, b := range seeds(f, "tlv") {
		f.Add(b, true)
		f.Add(b, false)
	}
	f.Fuzz(func(t *testing.T, b []byte, signed bool) {
		o, e := tlv.Parse(b, signed)
		if e != nil {
			return
		}
		var encoded []byte
		for _, v := range o.Fields {
			var h [4]byte
			binary.BigEndian.PutUint16(h[:2], v.Tag)
			binary.BigEndian.PutUint16(h[2:], uint16(len(v.Value)))
			encoded = append(encoded, h[:]...)
			encoded = append(encoded, v.Value...)
		}
		if !bytes.Equal(b, encoded) {
			t.Fatal("accepted bytes changed")
		}
		if signed {
			if len(b) < 68 || !bytes.Equal(o.Unsigned, b[:len(b)-68]) {
				t.Fatal("unsigned omission")
			}
		}
	})
}
func FuzzCredentialGrammar(f *testing.F) {
	for _, b := range seeds(f, "tlv") {
		for len(b) > 0 {
			tag, v, r, e := tlv.Next(b)
			if e != nil {
				break
			}
			if tag == 0x105 {
				f.Add(v)
			}
			b = r
		}
	}
	f.Add([]byte{})
	f.Fuzz(func(t *testing.T, b []byte) {
		if tlv.Credential(b) == nil {
			if len(b) < 23 || len(b) > 150 || tlv.Framing(b) != nil {
				t.Fatal("credential bounds/framing")
			}
		}
	})
}
func FuzzEnvelopeUnframe(f *testing.F) {
	for _, b := range seeds(f, "envelope") {
		f.Add(b)
	}
	f.Fuzz(func(t *testing.T, b []byte) {
		p, e := envelope.Unframe(b)
		if e != nil {
			return
		}
		if len(b) != 2032 || len(p) != int(binary.BigEndian.Uint16(b)) {
			t.Fatal("range")
		}
		framed, e := envelope.Frame(p)
		if e != nil || !bytes.Equal(framed, b) {
			t.Fatal("noncanonical envelope accepted")
		}
	})
}
func FuzzCommonPrefixDispatch(f *testing.F) {
	for _, b := range seeds(f, "dispatch") {
		f.Add(b)
	}
	f.Fuzz(func(t *testing.T, b []byte) {
		_ = dispatch.Classify(b)
		if len(b) > 2020 {
			return
		}
		future := append([]byte{0, 1, 0, 1, 1, 0, 2, 0, 1, 255}, b...)
		if dispatch.Classify(future) != "AUTHENTICATED_UNSUPPORTED_FUTURE" {
			t.Fatal("parsed opaque tail")
		}
	})
}
